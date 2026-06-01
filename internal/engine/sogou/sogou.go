package sogou

import (
	"bytes"
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/PuerkitoBio/goquery"

	"github.com/szStarWave/websurfx-go/internal/engine/common"
	"github.com/szStarWave/websurfx-go/internal/search"
)

const name = "sogou"

type Engine struct{}

func New() Engine {
	return Engine{}
}

func (Engine) Name() string {
	return name
}

func (Engine) Search(ctx context.Context, client *http.Client, query search.Query) ([]search.Result, *search.EngineError) {
	body, err := common.Get(ctx, client, searchURL(query), "https://www.sogou.com/")
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
	return fmt.Sprintf("https://www.sogou.com/web?query=%s&page=%d", search.EncodeQuery(query.Text), query.Page)
}

func Parse(body []byte) ([]search.Result, *search.EngineError) {
	doc, err := goquery.NewDocumentFromReader(bytes.NewReader(body))
	if err != nil {
		return nil, &search.EngineError{Kind: search.ErrorUnexpected, Message: err.Error()}
	}
	if doc.Find(".noresult_part1_container, .no-result").Length() > 0 {
		return nil, &search.EngineError{Kind: search.ErrorEmptyResult}
	}
	var results []search.Result
	doc.Find(".vrwrap, .rb, .result").Each(func(_ int, item *goquery.Selection) {
		if skipResultBlock(item) {
			return
		}
		titleNode := item.Find("h2 a, h3 a, .pt a, .vr-title a, [id^='sogou_vr_'][id*='_title_']").First()
		descNode := item.Find(".str_info, .txt-info, .ft, .text-layout, [id^='cacheresult_summary'], .base-ellipsis, .summary").First()
		href, ok := titleNode.Attr("href")
		title := search.Text(titleNode)
		if !ok || title == "" || skipResultLink(href) {
			return
		}
		results = append(results, search.Result{
			Title:       title,
			URL:         search.AbsURL("https://www.sogou.com", href),
			Description: search.Text(descNode),
			Engine:      []string{name},
		})
	})
	if len(results) == 0 {
		return nil, &search.EngineError{Kind: search.ErrorUnexpected, Message: "selector matched zero results"}
	}
	return results, nil
}

func skipResultBlock(item *goquery.Selection) bool {
	class, _ := item.Attr("class")
	id, _ := item.Attr("id")
	haystack := strings.ToLower(class + " " + id)
	for _, marker := range []string{"hintbox", "better-hint", "click-better-sugg", "ext_query"} {
		if strings.Contains(haystack, marker) {
			return true
		}
	}
	return false
}

func skipResultLink(href string) bool {
	href = strings.TrimSpace(strings.ToLower(href))
	if href == "" {
		return true
	}
	if strings.HasPrefix(href, "?") {
		return true
	}
	if strings.Contains(href, "s_from=hint") || strings.Contains(href, "pc_wendaka") {
		return true
	}
	return false
}
