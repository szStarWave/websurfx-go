package yahoo

import "testing"

func TestParseFixture(t *testing.T) {
	results, err := New().Parse([]byte(`
<div class="algo">
  <h3><a href="https://example.com/yahoo">Yahoo result</a></h3>
  <p class="compText">Yahoo summary</p>
</div>`))
	if err != nil {
		t.Fatalf("Parse returned error: %#v", err)
	}
	if results[0].Title != "Yahoo result" || results[0].Description != "Yahoo summary" {
		t.Fatalf("unexpected result %#v", results[0])
	}
}
