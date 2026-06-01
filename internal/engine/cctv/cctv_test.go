package cctv

import (
	"os"
	"testing"
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
	if got := results[0].Title; got != "\u54c8\u5c14\u6ee8\u697c\u9876\u574d\u584c\u4e8b\u6545\u81f44\u6b7b7\u4f24" {
		t.Fatalf("unexpected title %q", got)
	}
	if got := results[0].URL; got != "https://news.cctv.com/2021/08/09/ARTIJ7P0HV4wlGfQQKqNKTqT210809.shtml" {
		t.Fatalf("unexpected url %q", got)
	}
	if got := results[1].Description; got == "" {
		t.Fatal("second description is empty")
	}
}
