package so360

import (
	"bytes"
	"context"
	"fmt"
	"net/http"

	"github.com/PuerkitoBio/goquery"

	"github.com/szStarWave/websurfx-go/internal/engine/common"
	"github.com/szStarWave/websurfx-go/internal/search"
)

const name = "so360"

type Engine struct{}

func New() Engine {
	return Engine{}
}

func (Engine) Name() string {
	return name
}

func (Engine) Search(ctx context.Context, client *http.Client, query search.Query) ([]search.Result, *search.EngineError) {
	body, err := common.Get(ctx, client, searchURL(query), "https://www.so.com/")
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
	return fmt.Sprintf("https://www.so.com/s?q=%s&pn=%d", search.EncodeQuery(query.Text), query.Page)
}

func Parse(body []byte) ([]search.Result, *search.EngineError) {
	doc, err := goquery.NewDocumentFromReader(bytes.NewReader(body))
	if err != nil {
		return nil, &search.EngineError{Kind: search.ErrorUnexpected, Message: err.Error()}
	}
	if doc.Find(".res-none").Length() > 0 {
		return nil, &search.EngineError{Kind: search.ErrorEmptyResult}
	}
	var results []search.Result
	doc.Find("li.res-list, .result").Each(func(_ int, item *goquery.Selection) {
		titleNode := item.Find("h3 a, .res-title a, .js-title a").First()
		descNode := item.Find(".res-desc, .content, .summary").First()
		href, ok := titleNode.Attr("href")
		title := search.Text(titleNode)
		if !ok || title == "" {
			return
		}
		results = append(results, search.Result{
			Title:       title,
			URL:         search.AbsURL("https://www.so.com", href),
			Description: search.Text(descNode),
			Engine:      []string{name},
		})
	})
	if len(results) == 0 {
		return nil, &search.EngineError{Kind: search.ErrorUnexpected, Message: "selector matched zero results"}
	}
	return results, nil
}
