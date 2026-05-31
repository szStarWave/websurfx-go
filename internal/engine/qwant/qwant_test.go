package qwant

import "testing"

func TestParseFixture(t *testing.T) {
	results, err := New().Parse([]byte(`
<article>
  <h2><a href="https://example.com/qwant">Qwant result</a></h2>
  <p>Qwant summary</p>
</article>`))
	if err != nil {
		t.Fatalf("Parse returned error: %#v", err)
	}
	if results[0].Title != "Qwant result" || results[0].Description != "Qwant summary" {
		t.Fatalf("unexpected result %#v", results[0])
	}
}
