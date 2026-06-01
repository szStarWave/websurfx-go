package cctv

import (
	"bytes"
	"context"
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"github.com/PuerkitoBio/goquery"

	"github.com/szStarWave/websurfx-go/internal/engine/common"
	"github.com/szStarWave/websurfx-go/internal/search"
)

const name = "cctv"

type Engine struct{}

func New() Engine {
	return Engine{}
}

func (Engine) Name() string {
	return name
}

func (Engine) Search(ctx context.Context, client *http.Client, query search.Query) ([]search.Result, *search.EngineError) {
	body, err := common.Get(ctx, client, searchURL(query), "https://search.cctv.com/")
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
	return fmt.Sprintf("https://search.cctv.com/search.php?qtext=%s&type=web&page=%d", search.EncodeQuery(query.Text), query.Page)
}

func Parse(body []byte) ([]search.Result, *search.EngineError) {
	doc, err := goquery.NewDocumentFromReader(bytes.NewReader(body))
	if err != nil {
		return nil, &search.EngineError{Kind: search.ErrorUnexpected, Message: err.Error()}
	}
	if strings.Contains(search.Text(doc.Selection), "\u62b1\u6b49\uff0c\u6ca1\u6709\u627e\u5230") {
		return nil, &search.EngineError{Kind: search.ErrorEmptyResult}
	}
	var results []search.Result
	doc.Find("h3.tit").Each(func(_ int, item *goquery.Selection) {
		titleNode := item.Find("a[href]").First()
		href, ok := titleNode.Attr("href")
		title := search.Text(titleNode)
		if !ok || title == "" {
			return
		}
		container := item.Parent()
		desc := search.Text(container.Find("p.bre").First())
		if desc == "" {
			desc = search.Text(item.NextAllFiltered("p.bre").First())
		}
		results = append(results, search.Result{
			Title:       title,
			URL:         cctvTargetURL(href),
			Description: desc,
			Engine:      []string{name},
		})
	})
	if len(results) == 0 {
		return nil, &search.EngineError{Kind: search.ErrorUnexpected, Message: "selector matched zero results"}
	}
	return results, nil
}

func cctvTargetURL(href string) string {
	raw := search.AbsURL("https://search.cctv.com", href)
	parsed, err := url.Parse(raw)
	if err != nil {
		return raw
	}
	if target := parsed.Query().Get("targetpage"); target != "" {
		if decoded, err := url.QueryUnescape(target); err == nil {
			target = decoded
		}
		if strings.HasPrefix(target, "http://") || strings.HasPrefix(target, "https://") {
			return target
		}
	}
	return raw
}
