package shenma

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/PuerkitoBio/goquery"

	"github.com/szStarWave/websurfx-go/internal/engine/common"
	"github.com/szStarWave/websurfx-go/internal/search"
)

const name = "shenma"

type Engine struct{}

func New() Engine {
	return Engine{}
}

func (Engine) Name() string {
	return name
}

func (Engine) Search(ctx context.Context, client *http.Client, query search.Query) ([]search.Result, *search.EngineError) {
	body, err := common.Get(ctx, client, searchURL(query), "https://yz.m.sm.cn/")
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
	return fmt.Sprintf("https://yz.m.sm.cn/s?q=%s&page=%d", search.EncodeQuery(query.Text), query.Page)
}

func Parse(body []byte) ([]search.Result, *search.EngineError) {
	doc, err := goquery.NewDocumentFromReader(bytes.NewReader(body))
	if err != nil {
		return nil, &search.EngineError{Kind: search.ErrorUnexpected, Message: err.Error()}
	}
	if doc.Find(".empty, .no-result").Length() > 0 {
		return nil, &search.EngineError{Kind: search.ErrorEmptyResult}
	}
	var results []search.Result
	seen := map[string]struct{}{}
	doc.Find(`script[type="application/json"][id^="s-data-"]`).Each(func(_ int, script *goquery.Selection) {
		result, ok := parseResultScript(script.Text())
		if !ok {
			return
		}
		key := search.CanonicalURL(result.URL)
		if key == "" {
			key = strings.ToLower(result.Title)
		}
		if _, exists := seen[key]; exists {
			return
		}
		seen[key] = struct{}{}
		results = append(results, result)
	})
	if len(results) == 0 {
		return nil, &search.EngineError{Kind: search.ErrorUnexpected, Message: "selector matched zero results"}
	}
	return results, nil
}

func parseResultScript(raw string) (search.Result, bool) {
	var payload shenmaPayload
	if err := json.Unmarshal([]byte(strings.TrimSpace(raw)), &payload); err != nil {
		return search.Result{}, false
	}
	initial := payload.Data.InitialData
	title := firstText(
		initial.Title,
		initial.ArticleTitle,
		initial.TitleProps.Content,
	)
	url := firstURL(
		initial.URL,
		initial.TitleProps.DestURL,
		initial.SummaryProps.DestURL,
		initial.SourceProps.DestURL,
		initial.NUProps.RU,
		initial.NUProps.NU,
	)
	if title == "" || url == "" {
		return search.Result{}, false
	}
	description := firstText(
		initial.Desc,
		initial.SummaryProps.Content,
	)
	return search.Result{
		Title:       title,
		URL:         url,
		Description: description,
		Engine:      []string{name},
	}, true
}

func firstText(values ...string) string {
	for _, value := range values {
		value = strings.TrimSpace(value)
		if value == "" {
			continue
		}
		return search.TextFromHTML(value)
	}
	return ""
}

func firstURL(values ...string) string {
	for _, value := range values {
		value = strings.TrimSpace(value)
		if strings.HasPrefix(value, "http://") || strings.HasPrefix(value, "https://") {
			return value
		}
	}
	return ""
}

type shenmaPayload struct {
	Data struct {
		InitialData shenmaInitialData `json:"initialData"`
	} `json:"data"`
}

type shenmaInitialData struct {
	Title        string       `json:"title"`
	Desc         string       `json:"desc"`
	URL          string       `json:"url"`
	ArticleTitle string       `json:"ARTICLE_TITLE"`
	TitleProps   shenmaLink   `json:"titleProps"`
	SummaryProps shenmaLink   `json:"summaryProps"`
	SourceProps  shenmaSource `json:"sourceProps"`
	NUProps      shenmaNU     `json:"nuProps"`
}

type shenmaLink struct {
	Content string `json:"content"`
	DestURL string `json:"dest_url"`
}

type shenmaSource struct {
	DestURL string `json:"dest_url"`
}

type shenmaNU struct {
	RU string `json:"ru"`
	NU string `json:"nu"`
}
