package bing

import (
	"bytes"
	"context"
	"fmt"
	"net/http"

	"github.com/PuerkitoBio/goquery"

	"github.com/szStarWave/websurfx-go/internal/engine/common"
	"github.com/szStarWave/websurfx-go/internal/search"
)

const name = "bing"

type Engine struct{}

func New() Engine {
	return Engine{}
}

func (Engine) Name() string {
	return name
}

func (Engine) Search(ctx context.Context, client *http.Client, query search.Query) ([]search.Result, *search.EngineError) {
	body, err := common.Get(ctx, client, searchURL(query), "https://www.bing.com/")
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
	first := (query.Page-1)*10 + 1
	url := fmt.Sprintf("https://www.bing.com/search?q=%s&mkt=zh-CN&setlang=zh-CN", search.EncodeQuery(query.Text))
	if query.Page > 1 {
		url += fmt.Sprintf("&first=%d", first)
	}
	return url
}

func Parse(body []byte) ([]search.Result, *search.EngineError) {
	doc, err := goquery.NewDocumentFromReader(bytes.NewReader(body))
	if err != nil {
		return nil, &search.EngineError{Type: search.ErrorParse, Detail: err.Error()}
	}
	if doc.Find(".b_no").Length() > 0 {
		return nil, &search.EngineError{Type: search.ErrorEmptyResult}
	}
	var results []search.Result
	doc.Find("li.b_algo").Each(func(_ int, item *goquery.Selection) {
		titleNode := item.Find("h2 a").First()
		descNode := item.Find(".b_caption p").First()
		href, ok := titleNode.Attr("href")
		title := search.Text(titleNode)
		if !ok || title == "" {
			return
		}
		results = append(results, search.Result{
			Title:       title,
			URL:         href,
			Description: search.Text(descNode),
			Engine:      []string{name},
		})
	})
	if len(results) == 0 {
		return nil, &search.EngineError{Type: search.ErrorParse, Detail: "selector matched zero results"}
	}
	return results, nil
}
