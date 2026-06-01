package toutiao

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/szStarWave/websurfx-go/internal/engine/common"
	"github.com/szStarWave/websurfx-go/internal/search"
)

const name = "toutiao"

type Engine struct{}

func New() Engine {
	return Engine{}
}

func (Engine) Name() string {
	return name
}

func (Engine) Search(ctx context.Context, client *http.Client, query search.Query) ([]search.Result, *search.EngineError) {
	body, err := common.Get(ctx, client, searchURL(query), "https://so.toutiao.com/")
	if err != nil {
		return nil, common.WithEngine(name, err)
	}
	results, parseErr := Parse(body)
	if parseErr != nil {
		return nil, common.WithEngine(name, parseErr)
	}
	return results, nil
}

func searchURL(query search.Query) string {
	return fmt.Sprintf("https://so.toutiao.com/search?keyword=%s", search.EncodeQuery(query.Text))
}

func Parse(body []byte) ([]search.Result, *search.EngineError) {
	raw := string(body)
	if strings.Contains(raw, "captcha") || strings.Contains(raw, "\u5b89\u5168\u9a8c\u8bc1") {
		return nil, &search.EngineError{Kind: search.ErrorRequest, Message: "captcha or security verification page"}
	}
	var results []search.Result
	seen := map[string]bool{}
	for _, block := range extractMergeArticleBlocks(raw) {
		var articles []toutiaoArticle
		if err := json.Unmarshal([]byte(block), &articles); err != nil {
			continue
		}
		for _, article := range articles {
			result := article.result()
			if result.Title == "" || result.URL == "" {
				continue
			}
			key := search.CanonicalURL(result.URL)
			if key == "" || seen[key] {
				continue
			}
			seen[key] = true
			results = append(results, result)
		}
	}
	if len(results) == 0 {
		return nil, &search.EngineError{Kind: search.ErrorUnexpected, Message: "selector matched zero results"}
	}
	return results, nil
}

type toutiaoArticle struct {
	Title       string `json:"title"`
	Abstract    string `json:"abstract"`
	ArticleURL  string `json:"article_url"`
	Source      string `json:"source"`
	MediaName   string `json:"media_name"`
	PublishTime int64  `json:"publish_time"`
}

func (a toutiaoArticle) result() search.Result {
	description := strings.TrimSpace(a.Abstract)
	source := strings.TrimSpace(firstNonEmpty(a.Source, a.MediaName))
	if source != "" && description != "" {
		description = source + ": " + description
	}
	return search.Result{
		Title:       strings.TrimSpace(a.Title),
		URL:         strings.TrimSpace(a.ArticleURL),
		Description: description,
		Engine:      []string{name},
	}
}

func extractMergeArticleBlocks(raw string) []string {
	const marker = `"merge_article":`
	var blocks []string
	offset := 0
	for {
		index := strings.Index(raw[offset:], marker)
		if index < 0 {
			break
		}
		start := offset + index + len(marker)
		block, end := balancedJSON(raw, start, '[', ']')
		if end <= start {
			offset = start
			continue
		}
		blocks = append(blocks, block)
		offset = end
	}
	return blocks
}

func balancedJSON(raw string, start int, open, close byte) (string, int) {
	depth := 0
	inString := false
	escaped := false
	for i := start; i < len(raw); i++ {
		c := raw[i]
		if inString {
			if escaped {
				escaped = false
				continue
			}
			if c == '\\' {
				escaped = true
				continue
			}
			if c == '"' {
				inString = false
			}
			continue
		}
		if c == '"' {
			inString = true
			continue
		}
		if c == open {
			depth++
		}
		if c == close {
			depth--
			if depth == 0 {
				return raw[start : i+1], i + 1
			}
		}
	}
	return "", -1
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return value
		}
	}
	return ""
}
