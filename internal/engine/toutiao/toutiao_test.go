package toutiao

import (
	"os"
	"testing"

	"github.com/szStarWave/websurfx-go/internal/search"
)

func TestParseMergeArticleFixture(t *testing.T) {
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
	if got := results[0].Title; got != "\u54c8\u5c14\u6ee8\u906d\u9047\u6781\u7aef\u5927\u98ce\uff0c\u4e00\u6e38\u4e50\u56ed\u8fc7\u5c71\u8f66\u60ac\u505c\u534a\u7a7a" {
		t.Fatalf("unexpected first title %q", got)
	}
	if got := results[0].Description; got != "\u6f47\u6e58\u6668\u62a5: 5\u670831\u65e5\u4e0b\u5348\uff0c\u54c8\u5c14\u6ee8\u5e02\u906d\u9047\u6781\u7aef\u5927\u98ce\u5929\u6c14\u3002" {
		t.Fatalf("unexpected first description %q", got)
	}
	if got := results[1].URL; got != "https://m.toutiao.com/group/2/" {
		t.Fatalf("unexpected second url %q", got)
	}
}

func TestParseCaptchaFixture(t *testing.T) {
	_, engineErr := Parse([]byte("<title>\u5b89\u5168\u9a8c\u8bc1</title>"))
	if engineErr == nil || engineErr.Kind != search.ErrorRequest {
		t.Fatalf("expected request error, got %#v", engineErr)
	}
}
