package chinanews

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/szStarWave/websurfx-go/internal/engine/common"
	"github.com/szStarWave/websurfx-go/internal/search"
)

const name = "chinanews"

type Engine struct{}

func New() Engine {
	return Engine{}
}

func (Engine) Name() string {
	return name
}

func (Engine) Search(ctx context.Context, client *http.Client, query search.Query) ([]search.Result, *search.EngineError) {
	body, err := common.Get(ctx, client, searchURL(query), "https://sou.chinanews.com.cn/")
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
	return fmt.Sprintf("https://sou.chinanews.com.cn/search.do?q=%s&page=%d", search.EncodeQuery(query.Text), query.Page)
}

func Parse(body []byte) ([]search.Result, *search.EngineError) {
	raw := string(body)
	block, ok := extractJavaScriptArray(raw, "docArr")
	if !ok {
		return nil, &search.EngineError{Kind: search.ErrorUnexpected, Message: "docArr not found"}
	}
	var docs []chinanewsDocument
	if err := json.Unmarshal([]byte(block), &docs); err != nil {
		return nil, &search.EngineError{Kind: search.ErrorUnexpected, Message: err.Error()}
	}
	var results []search.Result
	for _, doc := range docs {
		result := doc.result()
		if result.Title == "" || result.URL == "" {
			continue
		}
		results = append(results, result)
	}
	if len(results) == 0 {
		return nil, &search.EngineError{Kind: search.ErrorEmptyResult}
	}
	return results, nil
}

type chinanewsDocument struct {
	Title             any    `json:"title"`
	URL               string `json:"url"`
	ContentWithoutTag string `json:"content_without_tag"`
	CreateTime        string `json:"createtime"`
	PubTime           string `json:"pubtime"`
}

func (d chinanewsDocument) result() search.Result {
	description := search.TextFromHTML(d.ContentWithoutTag)
	date := firstNonEmpty(d.PubTime, d.CreateTime)
	if date != "" && description != "" {
		description = date + " " + description
	}
	return search.Result{
		Title:       search.TextFromHTML(titleString(d.Title)),
		URL:         strings.TrimSpace(d.URL),
		Description: description,
		Engine:      []string{name},
	}
}

func titleString(value any) string {
	switch typed := value.(type) {
	case string:
		return typed
	case []any:
		var parts []string
		for _, item := range typed {
			if text, ok := item.(string); ok {
				parts = append(parts, text)
			}
		}
		return strings.Join(parts, " ")
	default:
		return ""
	}
}

func extractJavaScriptArray(raw, name string) (string, bool) {
	marker := "var " + name + " = "
	index := strings.Index(raw, marker)
	if index < 0 {
		return "", false
	}
	start := index + len(marker)
	block, end := balancedJSON(raw, start, '[', ']')
	return block, end > start
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
