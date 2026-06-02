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

func TestAggregatorPrioritizesOfficialTechnicalSources(t *testing.T) {
	engines := []Engine{
		fakeEngine{name: "so360", results: []Result{
			{
				Title:       "winml api reference microsoft - 360\u7ffb\u8bd1",
				URL:         "https://fanyi.so.com/?src=onebox",
				Description: "translation onebox",
			},
			{
				Title:       "winml api reference microsoft - 360\u6587\u5e93",
				URL:         "https://wenku.so.com/s?q=winml%20api",
				Description: "document search landing page",
			},
		}},
		fakeEngine{name: "bing", results: []Result{
			{
				Title:       "\u5fae\u8f6f_\u767e\u5ea6\u767e\u79d1",
				URL:         "https://baike.baidu.com/item/%E5%BE%AE%E8%BD%AF/124767",
				Description: "company encyclopedia page",
			},
			{
				Title:       "Windows ML APIs in Windows.AI.MachineLearning | Microsoft Learn",
				URL:         "https://learn.microsoft.com/windows/ai/windows-ml/api-reference",
				Description: "Windows ML API reference for WinML model inference.",
			},
		}},
	}
	aggregator := NewAggregatorWithClient(http.DefaultClient, nil, engines, FilterOptions{})

	response := aggregator.Search(context.Background(), Query{Text: "WinML API reference Microsoft", Page: 1})
	if len(response.Results) == 0 {
		t.Fatalf("expected ranked results")
	}
	if response.Results[0].URL != "https://learn.microsoft.com/windows/ai/windows-ml/api-reference" {
		t.Fatalf("top result = %#v, want Microsoft Learn first; all=%#v", response.Results[0], response.Results)
	}
}

func TestEncodeQuerySpecialCharacters(t *testing.T) {
	if got := EncodeQuery("特朗普 open ai c++"); got != "%E7%89%B9%E6%9C%97%E6%99%AE+open+ai+c%2B%2B" {
		t.Fatalf("unexpected encoded query %q", got)
	}
}
