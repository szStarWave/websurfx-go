package search

import (
	"context"
	"net/http"
	"net/url"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"
)

type Cache interface {
	Get(key string) (Response, bool)
	Set(key string, value Response)
}

type Aggregator struct {
	client  *http.Client
	cache   Cache
	engines []Engine
	filters FilterOptions
}

func NewAggregator(timeout time.Duration, cache Cache, engines []Engine) *Aggregator {
	return &Aggregator{
		client:  &http.Client{Timeout: timeout},
		cache:   cache,
		engines: engines,
	}
}

func NewAggregatorWithClient(client *http.Client, cache Cache, engines []Engine, filters FilterOptions) *Aggregator {
	return &Aggregator{
		client:  client,
		cache:   cache,
		engines: engines,
		filters: filters,
	}
}

func (a *Aggregator) Search(ctx context.Context, query Query) Response {
	start := time.Now()
	query = NormalizeQuery(query)
	key := cacheKey(query, a.engines)
	if a.cache != nil {
		if cached, ok := a.cache.Get(key); ok {
			cached.Duration = Since(start)
			return cached
		}
	}

	type engineResponse struct {
		engine  string
		results []Result
		err     *EngineError
	}

	responses := make(chan engineResponse, len(a.engines))
	var wg sync.WaitGroup
	for _, engine := range a.engines {
		engine := engine
		wg.Add(1)
		go func() {
			defer wg.Done()
			results, err := engine.Search(ctx, a.client, query)
			responses <- engineResponse{engine: engine.Name(), results: results, err: err}
		}()
	}
	wg.Wait()
	close(responses)

	resultMap := map[string]Result{}
	var errors []EngineError
	for response := range responses {
		if response.err != nil {
			errors = append(errors, *response.err)
			continue
		}
		for _, result := range response.results {
			result.URL = CanonicalURL(result.URL)
			if result.URL == "" || !a.allowed(result) {
				continue
			}
			key := result.URL
			if existing, ok := resultMap[key]; ok {
				existing.Engine = mergeEngines(existing.Engine, response.engine)
				if existing.Description == "" {
					existing.Description = result.Description
				}
				resultMap[result.URL] = existing
				continue
			}
			if len(result.Engine) == 0 {
				result.Engine = []string{response.engine}
			}
			result.Engine = mergeEngines(result.Engine)
			resultMap[key] = result
		}
	}

	results := make([]Result, 0, len(resultMap))
	for _, result := range resultMap {
		results = append(results, result)
	}
	sort.SliceStable(results, func(i, j int) bool {
		left := scoreResult(query.Text, results[i])
		right := scoreResult(query.Text, results[j])
		if left == right {
			return strings.ToLower(results[i].Title) < strings.ToLower(results[j].Title)
		}
		return left > right
	})
	sort.SliceStable(errors, func(i, j int) bool {
		return errors[i].Engine < errors[j].Engine
	})

	out := Response{
		Query:            query.Text,
		Page:             query.Page,
		HasNextPage:      len(results) >= 10,
		Engines:          engineNames(a.engines),
		Results:          results,
		EngineErrorsInfo: errors,
		Errors:           errors,
		Duration:         Since(start),
	}
	if a.cache != nil {
		a.cache.Set(key, out)
	}
	return out
}

func (a *Aggregator) EngineNames() []string {
	return engineNames(a.engines)
}

func engineNames(engines []Engine) []string {
	names := make([]string, 0, len(engines))
	for _, engine := range engines {
		names = append(names, engine.Name())
	}
	sort.Strings(names)
	return names
}

func (a *Aggregator) allowed(result Result) bool {
	haystack := strings.ToLower(strings.Join([]string{result.URL, result.Title, result.Description}, " "))
	if len(a.filters.Allowlist) > 0 {
		matched := false
		for _, item := range a.filters.Allowlist {
			item = strings.ToLower(strings.TrimSpace(item))
			if item != "" && strings.Contains(haystack, item) {
				matched = true
				break
			}
		}
		if !matched {
			return false
		}
	}
	for _, item := range a.filters.Blocklist {
		item = strings.ToLower(strings.TrimSpace(item))
		if item != "" && strings.Contains(haystack, item) {
			return false
		}
	}
	return true
}

func mergeEngines(values ...interface{}) []string {
	seen := map[string]bool{}
	var out []string
	for _, value := range values {
		switch typed := value.(type) {
		case string:
			if typed != "" && !seen[typed] {
				seen[typed] = true
				out = append(out, typed)
			}
		case []string:
			for _, item := range typed {
				if item != "" && !seen[item] {
					seen[item] = true
					out = append(out, item)
				}
			}
		}
	}
	sort.Strings(out)
	return out
}

func scoreResult(query string, result Result) int {
	score := len(result.Engine) * 5
	terms := queryTerms(query)
	title := strings.ToLower(result.Title)
	description := strings.ToLower(result.Description)
	host := ""
	if parsed, err := url.Parse(result.URL); err == nil {
		host = strings.ToLower(parsed.Host)
	}
	for _, term := range terms {
		if strings.Contains(title, term) {
			score += 20
		}
		if strings.Contains(description, term) {
			score += 8
		}
		if strings.Contains(host, term) {
			score += 3
		}
	}
	if isTechnicalQuery(query) {
		if trustedTechnicalResult(host, title, description) {
			score += 90
		}
		if lowValueTechnicalResult(host, title, description) {
			score -= 100
		}
	}
	return score
}

func queryTerms(query string) []string {
	raw := strings.Fields(strings.ToLower(query))
	if len(raw) == 0 && strings.TrimSpace(query) != "" {
		raw = []string{strings.ToLower(strings.TrimSpace(query))}
	}
	return raw
}

func isTechnicalQuery(query string) bool {
	query = strings.ToLower(query)
	return containsAny(query,
		"api", "reference", "documentation", "docs", "developer",
		"winml", "windows ml", "windows machine learning",
		"onnx", "directml", "openvino", "cuda", "rocm",
		"runtime", "inference", "sdk",
	)
}

func trustedTechnicalResult(host, title, description string) bool {
	haystack := strings.Join([]string{host, title, description}, " ")
	if containsAny(host,
		"learn.microsoft.com",
		"docs.microsoft.com",
		"github.com",
		"developer.microsoft.com",
		"onnxruntime.ai",
		"docs.nvidia.com",
		"docs.openvino.ai",
		"rocm.docs.amd.com",
	) {
		if containsAny(haystack, "microsoft learn", "official", "api", "reference", "documentation", "docs", "github", "windows ml", "winml", "onnx", "directml") {
			return true
		}
	}
	return containsAny(title, "microsoft learn", "official documentation", "api reference", "github - microsoft")
}

func lowValueTechnicalResult(host, title, description string) bool {
	haystack := strings.Join([]string{host, title, description}, " ")
	return containsAny(haystack,
		"360文库",
		"360翻译",
		"360图片",
		"360视频",
		"wenku.so.com",
		"fanyi.so.com",
		"image.so.com",
		"baike.baidu.com",
		"blog.csdn.net",
		"csdn文库",
		"m365.cloud.microsoft",
		"office.com",
		"signup.live.com",
		"microsoftstore.com.cn",
		"microsoft 365 copilot",
		"microsoft 365",
		"surface_windows_office",
	)
}

func containsAny(value string, needles ...string) bool {
	for _, needle := range needles {
		needle = strings.TrimSpace(strings.ToLower(needle))
		if needle != "" && strings.Contains(value, needle) {
			return true
		}
	}
	return false
}

func cacheKey(query Query, engines []Engine) string {
	names := engineNames(engines)
	return strings.Join([]string{query.Text, time.Now().Format("2006-01-02"), strings.Join(names, ","), strconv.Itoa(query.Page)}, "|")
}
