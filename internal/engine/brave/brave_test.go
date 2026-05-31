package brave

import "testing"

func TestParseFixture(t *testing.T) {
	results, err := New().Parse([]byte(`
<div class="snippet">
  <a class="heading-serpresult" href="https://example.com/brave">Brave result</a>
  <p class="snippet-description">Brave summary</p>
</div>`))
	if err != nil {
		t.Fatalf("Parse returned error: %#v", err)
	}
	if results[0].Title != "Brave result" || results[0].Description != "Brave summary" {
		t.Fatalf("unexpected result %#v", results[0])
	}
}
