package searx

import "testing"

func TestParseFixture(t *testing.T) {
	results, err := New("https://searx.example").Parse([]byte(`
<article class="result">
  <h3><a href="/url">Searx result</a></h3>
  <p class="content">Searx summary</p>
</article>`))
	if err != nil {
		t.Fatalf("Parse returned error: %#v", err)
	}
	if results[0].Title != "Searx result" || results[0].URL != "https://searx.example/url" {
		t.Fatalf("unexpected result %#v", results[0])
	}
}
