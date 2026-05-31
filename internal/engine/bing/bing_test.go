package bing

import (
	"os"
	"strings"
	"testing"

	"github.com/szStarWave/websurfx-go/internal/search"
)

func TestParseFixture(t *testing.T) {
	body, err := os.ReadFile("testdata/result.html")
	if err != nil {
		t.Fatal(err)
	}
	results, engineErr := Parse(body)
	if engineErr != nil {
		t.Fatalf("Parse returned error: %#v", engineErr)
	}
	if got := results[0].Title; got != "中国新闻 - Bing" {
		t.Fatalf("unexpected title %q", got)
	}
}

func TestParseEmptyFixture(t *testing.T) {
	body, err := os.ReadFile("testdata/empty.html")
	if err != nil {
		t.Fatal(err)
	}
	_, engineErr := Parse(body)
	if engineErr == nil || engineErr.Type != search.ErrorEmptyResult {
		t.Fatalf("expected empty result error, got %#v", engineErr)
	}
}

func TestSearchURLEncodesChineseQuery(t *testing.T) {
	url := searchURL(search.Query{Text: "特朗普 c++", Page: 2})
	if !strings.Contains(url, "q=%E7%89%B9%E6%9C%97%E6%99%AE+c%2B%2B") {
		t.Fatalf("query was not encoded: %s", url)
	}
	if !strings.Contains(url, "mkt=zh-CN") || !strings.Contains(url, "setlang=zh-CN") {
		t.Fatalf("missing Chinese market options: %s", url)
	}
}
