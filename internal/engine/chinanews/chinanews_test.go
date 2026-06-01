package chinanews

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
	if got := results[0].Title; got != "\u6ee1\u8f7d28\u5428\u7164\u7126\u6cb9\u7f50\u8f66\u7740\u706b \u4ea4\u8b66\u9006\u5411\u51b2\u950b\u59a5\u5584\u5904\u7f6e" {
		t.Fatalf("unexpected title %q", got)
	}
	if got := results[0].Description; got == "" || got[:10] != "2022-03-03" {
		t.Fatalf("unexpected description %q", got)
	}
	if got := results[1].Title; got != "\u6c14\u6e29\u56de\u5347\u677e\u82b1\u6c5f\u51b0\u878d\u6c34\u9614 \u54c8\u5c14\u6ee8\u6d77\u4e8b\u5c40\u591a\u63aa\u5e76\u4e3e" {
		t.Fatalf("unexpected array title %q", got)
	}
}
