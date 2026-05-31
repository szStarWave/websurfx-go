package startpage

import "testing"

func TestParseFixture(t *testing.T) {
	results, err := New().Parse([]byte(`
<div class="w-gl__result">
  <a class="w-gl__result-title" href="https://example.com/startpage">Startpage result</a>
  <p class="w-gl__description">Startpage summary</p>
</div>`))
	if err != nil {
		t.Fatalf("Parse returned error: %#v", err)
	}
	if results[0].Title != "Startpage result" || results[0].Description != "Startpage summary" {
		t.Fatalf("unexpected result %#v", results[0])
	}
}
