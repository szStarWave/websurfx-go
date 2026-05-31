package htmlengine

import (
	"testing"

	"github.com/szStarWave/websurfx-go/internal/search"
)

func TestParseResultEmptyAndUnexpected(t *testing.T) {
	engine := New(Config{
		Name:           "fixture",
		BaseURL:        "https://example.com",
		PathFormat:     "/search?q=%s&page=%d",
		ResultSelector: ".result",
		TitleSelector:  "h2 a[href]",
		DescSelector:   "p",
		EmptySelector:  ".empty",
	})

	results, err := engine.Parse([]byte(`<div class="result"><h2><a href="/one">China news</a></h2><p>Summary text</p></div>`))
	if err != nil {
		t.Fatalf("Parse returned error: %#v", err)
	}
	if results[0].URL != "https://example.com/one" {
		t.Fatalf("unexpected URL %q", results[0].URL)
	}
	if results[0].Engine[0] != "fixture" {
		t.Fatalf("unexpected engine %v", results[0].Engine)
	}

	_, err = engine.Parse([]byte(`<div class="empty">No results</div>`))
	if err == nil || err.Kind != search.ErrorEmptyResult {
		t.Fatalf("expected EmptyResultSet, got %#v", err)
	}

	_, err = engine.Parse([]byte(`<main>layout changed</main>`))
	if err == nil || err.Kind != search.ErrorUnexpected {
		t.Fatalf("expected UnexpectedError, got %#v", err)
	}
}
