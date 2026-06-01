package shenma

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
	if len(results) != 2 {
		t.Fatalf("results len = %d, want 2: %#v", len(results), results)
	}
	if got := results[0].Title; got != "极端大风突袭哈尔滨：过山车停摆倒挂半空" {
		t.Fatalf("unexpected first title %q", got)
	}
	if got := results[0].Description; got != "5月31日，受强对流天气影响，哈尔滨国际会展中心体育场相关设施受到损坏。" {
		t.Fatalf("unexpected first description %q", got)
	}
	if got := results[1].URL; got != "https://k.sina.cn/article_7857201856_1d45362c0019067db4.html" {
		t.Fatalf("unexpected second url %q", got)
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
