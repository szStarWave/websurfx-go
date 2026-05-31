package search

import (
	"context"
	"net/http"
	"testing"
)

type fakeEngine struct {
	name    string
	results []Result
	err     *EngineError
}

func (e fakeEngine) Name() string {
	return e.name
}

func (e fakeEngine) Search(context.Context, *http.Client, Query) ([]Result, *EngineError) {
	return e.results, e.err
}

func TestAggregatorMergesRanksFiltersAndKeepsErrors(t *testing.T) {
	engines := []Engine{
		fakeEngine{name: "a", results: []Result{
			{Title: "Other", URL: "https://example.com/other#frag", Description: "plain"},
			{Title: "China news", URL: "https://redirect.test/?url=https%3A%2F%2Fexample.com%2Fchina%23frag", Description: "China summary"},
		}},
		fakeEngine{name: "b", results: []Result{
			{Title: "China news duplicate", URL: "https://example.com/china/", Description: ""},
		}},
		fakeEngine{name: "broken", err: &EngineError{Engine: "broken", Kind: ErrorRequest, Message: "timeout"}},
	}
	aggregator := NewAggregatorWithClient(http.DefaultClient, nil, engines, FilterOptions{Blocklist: []string{"other"}})

	response := aggregator.Search(context.Background(), Query{Text: "china", Page: 1})
	if len(response.Results) != 1 {
		t.Fatalf("expected one filtered/merged result, got %#v", response.Results)
	}
	if response.Results[0].URL != "https://example.com/china" {
		t.Fatalf("unexpected canonical URL %q", response.Results[0].URL)
	}
	if len(response.Results[0].Engine) != 2 {
		t.Fatalf("expected merged engines, got %v", response.Results[0].Engine)
	}
	if len(response.EngineErrorsInfo) != 1 || response.EngineErrorsInfo[0].Kind != ErrorRequest {
		t.Fatalf("expected retained engine error, got %#v", response.EngineErrorsInfo)
	}
	if len(response.EngineErrorsInfo) != len(response.Errors) {
		t.Fatalf("compat errors mirror not populated: %#v", response.Errors)
	}
}

func TestEncodeQuerySpecialCharacters(t *testing.T) {
	if got := EncodeQuery("特朗普 open ai c++"); got != "%E7%89%B9%E6%9C%97%E6%99%AE+open+ai+c%2B%2B" {
		t.Fatalf("unexpected encoded query %q", got)
	}
}
