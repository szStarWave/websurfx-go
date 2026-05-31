package wikipedia

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/szStarWave/websurfx-go/internal/engine/common"
	"github.com/szStarWave/websurfx-go/internal/search"
)

const name = "zhwikipedia"

type Engine struct{}

type response struct {
	Query struct {
		Search []item `json:"search"`
	} `json:"query"`
}

type item struct {
	Title   string `json:"title"`
	Snippet string `json:"snippet"`
	PageID  int64  `json:"pageid"`
}

func NewZH() Engine {
	return Engine{}
}

func (Engine) Name() string {
	return name
}

func (Engine) Search(ctx context.Context, client *http.Client, query search.Query) ([]search.Result, *search.EngineError) {
	body, err := common.Get(ctx, client, searchURL(query), "https://zh.wikipedia.org/")
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
	offset := (query.Page - 1) * 10
	return fmt.Sprintf("https://zh.wikipedia.org/w/api.php?action=query&list=search&format=json&utf8=1&srlimit=10&sroffset=%d&srsearch=%s", offset, search.EncodeQuery(query.Text))
}

func Parse(body []byte) ([]search.Result, *search.EngineError) {
	var parsed response
	if err := json.Unmarshal(body, &parsed); err != nil {
		return nil, &search.EngineError{Kind: search.ErrorUnexpected, Message: err.Error()}
	}
	if len(parsed.Query.Search) == 0 {
		return nil, &search.EngineError{Kind: search.ErrorEmptyResult}
	}
	results := make([]search.Result, 0, len(parsed.Query.Search))
	for _, item := range parsed.Query.Search {
		results = append(results, search.Result{
			Title:       item.Title,
			URL:         fmt.Sprintf("https://zh.wikipedia.org/?curid=%d", item.PageID),
			Description: search.TextFromHTML(item.Snippet),
			Engine:      []string{name},
		})
	}
	return results, nil
}
