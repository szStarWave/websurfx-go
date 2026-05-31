package duckduckgo

import "testing"

func TestParseFixture(t *testing.T) {
	results, err := New().Parse([]byte(`
<div class="result">
  <h2 class="result__title"><a href="https://example.com/ddg">Duck result</a></h2>
  <a class="result__snippet">Duck summary</a>
</div>`))
	if err != nil {
		t.Fatalf("Parse returned error: %#v", err)
	}
	if results[0].Title != "Duck result" || results[0].Description != "Duck summary" {
		t.Fatalf("unexpected result %#v", results[0])
	}
}
