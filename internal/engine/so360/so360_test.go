package so360

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
	if got := results[0].Title; got != "中国新闻 - 360" {
		t.Fatalf("unexpected title %q", got)
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
