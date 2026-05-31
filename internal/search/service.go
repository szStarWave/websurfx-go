package search

import (
	"context"
	"net/http"
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
}

func NewAggregator(timeout time.Duration, cache Cache, engines []Engine) *Aggregator {
	return &Aggregator{
		client:  &http.Client{Timeout: timeout},
		cache:   cache,
		engines: engines,
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
			if existing, ok := resultMap[result.URL]; ok {
				existing.Engine = append(existing.Engine, response.engine)
				resultMap[result.URL] = existing
				continue
			}
			if len(result.Engine) == 0 {
				result.Engine = []string{response.engine}
			}
			resultMap[result.URL] = result
		}
	}

	results := make([]Result, 0, len(resultMap))
	for _, result := range resultMap {
		results = append(results, result)
	}
	sort.SliceStable(results, func(i, j int) bool {
		return strings.ToLower(results[i].Title) < strings.ToLower(results[j].Title)
	})
	sort.SliceStable(errors, func(i, j int) bool {
		return errors[i].Engine < errors[j].Engine
	})

	out := Response{
		Query:    query.Text,
		Page:     query.Page,
		Results:  results,
		Errors:   errors,
		Duration: Since(start),
	}
	if a.cache != nil {
		a.cache.Set(key, out)
	}
	return out
}

func cacheKey(query Query, engines []Engine) string {
	names := make([]string, 0, len(engines))
	for _, engine := range engines {
		names = append(names, engine.Name())
	}
	sort.Strings(names)
	return strings.Join([]string{query.Text, time.Now().Format("2006-01-02"), strings.Join(names, ","), strconv.Itoa(query.Page)}, "|")
}
