package htmlengine

import (
	"bytes"
	"context"
	"fmt"
	"net/http"

	"github.com/PuerkitoBio/goquery"

	"github.com/szStarWave/websurfx-go/internal/engine/common"
	"github.com/szStarWave/websurfx-go/internal/search"
)

type Engine struct {
	name            string
	baseURL         string
	referer         string
	pathFormat      string
	pageParamOffset int
	pageValue       func(page int) int
	resultSelector  string
	titleSelector   string
	descSelector    string
	emptySelector   string
}

type Config struct {
	Name            string
	BaseURL         string
	Referer         string
	PathFormat      string
	PageParamOffset int
	PageValue       func(page int) int
	ResultSelector  string
	TitleSelector   string
	DescSelector    string
	EmptySelector   string
}

func New(cfg Config) Engine {
	return Engine{
		name:            cfg.Name,
		baseURL:         cfg.BaseURL,
		referer:         cfg.Referer,
		pathFormat:      cfg.PathFormat,
		pageParamOffset: cfg.PageParamOffset,
		pageValue:       cfg.PageValue,
		resultSelector:  cfg.ResultSelector,
		titleSelector:   cfg.TitleSelector,
		descSelector:    cfg.DescSelector,
		emptySelector:   cfg.EmptySelector,
	}
}

func (e Engine) Name() string {
	return e.name
}

func (e Engine) Search(ctx context.Context, client *http.Client, query search.Query) ([]search.Result, *search.EngineError) {
	body, err := common.Get(ctx, client, e.searchURL(query), e.referer)
	if err != nil {
		return nil, common.WithEngine(e.name, err)
	}
	results, parseErr := e.Parse(body)
	if parseErr != nil {
		return nil, common.WithEngine(e.name, parseErr)
	}
	return results, nil
}

func (e Engine) searchURL(query search.Query) string {
	pageValue := query.Page + e.pageParamOffset
	if e.pageValue != nil {
		pageValue = e.pageValue(query.Page)
	}
	return e.baseURL + fmt.Sprintf(e.pathFormat, search.EncodeQuery(query.Text), pageValue)
}

func (e Engine) Parse(body []byte) ([]search.Result, *search.EngineError) {
	doc, err := goquery.NewDocumentFromReader(bytes.NewReader(body))
	if err != nil {
		return nil, &search.EngineError{Kind: search.ErrorUnexpected, Message: err.Error()}
	}
	if e.emptySelector != "" && doc.Find(e.emptySelector).Length() > 0 {
		return nil, &search.EngineError{Kind: search.ErrorEmptyResult}
	}
	var results []search.Result
	doc.Find(e.resultSelector).Each(func(_ int, item *goquery.Selection) {
		titleNode := item.Find(e.titleSelector).First()
		descNode := item.Find(e.descSelector).First()
		href, ok := titleNode.Attr("href")
		title := search.Text(titleNode)
		if !ok || title == "" {
			return
		}
		results = append(results, search.Result{
			Title:       title,
			URL:         search.AbsURL(e.baseURL, href),
			Description: search.Text(descNode),
			Engine:      []string{e.name},
		})
	})
	if len(results) == 0 {
		return nil, &search.EngineError{Kind: search.ErrorUnexpected, Message: "selector matched zero results"}
	}
	return results, nil
}
