package wikipedia

import (
	"os"
	"strings"
	"testing"

	"github.com/szStarWave/websurfx-go/internal/search"
)

func TestParseFixture(t *testing.T) {
	body, err := os.ReadFile("testdata/result.json")
	if err != nil {
		t.Fatal(err)
	}
	results, engineErr := Parse(body)
	if engineErr != nil {
		t.Fatalf("Parse returned error: %#v", engineErr)
	}
	if got := results[0].Title; got != "中国" {
		t.Fatalf("unexpected title %q", got)
	}
	if got := results[0].Description; got != "中华人民共和国，简称中国" {
		t.Fatalf("unexpected description %q", got)
	}
}

func TestParseEmptyFixture(t *testing.T) {
	body, err := os.ReadFile("testdata/empty.json")
	if err != nil {
		t.Fatal(err)
	}
	_, engineErr := Parse(body)
	if engineErr == nil || engineErr.Type != search.ErrorEmptyResult {
		t.Fatalf("expected empty result error, got %#v", engineErr)
	}
}

func TestSearchURLTargetsZHAPI(t *testing.T) {
	url := searchURL(search.Query{Text: "中国", Page: 1})
	if !strings.Contains(url, "zh.wikipedia.org/w/api.php") {
		t.Fatalf("unexpected url: %s", url)
	}
	if !strings.Contains(url, "srsearch=%E4%B8%AD%E5%9B%BD") {
		t.Fatalf("query was not encoded: %s", url)
	}
}
