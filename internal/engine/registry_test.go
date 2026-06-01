package engine

import "testing"

func TestBuildAdditionalWebEngines(t *testing.T) {
	names := []string{"duckduckgo", "brave", "qwant", "startpage", "yahoo", "toutiao", "searx:https://example.com"}
	engines, err := Build(names)
	if err != nil {
		t.Fatal(err)
	}
	if len(engines) != len(names) {
		t.Fatalf("expected %d engines, got %d", len(names), len(engines))
	}
	if engines[len(engines)-1].Name() != "searx" {
		t.Fatalf("unexpected searx engine name %q", engines[len(engines)-1].Name())
	}
}
