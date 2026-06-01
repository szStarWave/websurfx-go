package sogou

import (
	"os"
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
	if got := results[0].Title; got != "中国新闻 - 搜狗" {
		t.Fatalf("unexpected title %q", got)
	}
}

func TestParseModernFixtureSkipsHintsAndKeepsSummaries(t *testing.T) {
	body, err := os.ReadFile("testdata/modern.html")
	if err != nil {
		t.Fatal(err)
	}
	results, engineErr := Parse(body)
	if engineErr != nil {
		t.Fatalf("Parse returned error: %#v", engineErr)
	}
	if len(results) != 2 {
		t.Fatalf("results len = %d, want 2: %#v", len(results), results)
	}
	if got := results[0].Title; got != "\u54c8\u5c14\u6ee8\u906d\u9047\u6781\u7aef\u5927\u98ce, \u8fc7\u5c71\u8f66\u60ac\u505c\u534a\u7a7a" {
		t.Fatalf("unexpected first title %q", got)
	}
	if got := results[0].Description; got == "" || got == "\u817e\u8baf\u7f51 https://mp.weixin.qq.com/c... 14\u5c0f\u65f6\u524d" {
		t.Fatalf("unexpected first description %q", got)
	}
	if got := results[1].Description; got != "\u54c8\u5c14\u6ee8\u5927\u98ce\u9020\u6210\u591a\u8d77\u4e8b\u6545, \u6709\u5173\u90e8\u95e8\u53d1\u5e03\u9884\u8b66\u3002" {
		t.Fatalf("unexpected second description %q", got)
	}
}

func TestParseEmptyFixture(t *testing.T) {
	body, err := os.ReadFile("testdata/empty.html")
	if err != nil {
		t.Fatal(err)
	}
	_, engineErr := Parse(body)
	if engineErr == nil || engineErr.Kind != search.ErrorEmptyResult {
		t.Fatalf("expected empty result error, got %#v", engineErr)
	}
}
