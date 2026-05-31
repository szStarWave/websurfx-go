package websurfx

import (
	"path/filepath"
	"testing"
)

func TestPublicDefaultsUseAllEngines(t *testing.T) {
	if len(AllEngines()) <= len(ChineseDefaultEngines()) {
		t.Fatalf("expected all engines to include optional engines: %v", AllEngines())
	}
	if len(DefaultEngines()) != len(AllEngines()) {
		t.Fatalf("DefaultEngines should return all engines")
	}
	cfg := DefaultConfig()
	if len(cfg.Search.Engines) != len(AllEngines()) {
		t.Fatalf("default config engines = %v", cfg.Search.Engines)
	}
}

func TestNewFromMissingConfigFile(t *testing.T) {
	client, err := NewFromConfigFile(filepath.Join(t.TempDir(), "missing.yaml"))
	if err != nil {
		t.Fatal(err)
	}
	if client == nil {
		t.Fatal("nil client")
	}
}

func TestNewFromEmptyConfigUsesDefaults(t *testing.T) {
	client, err := NewFromConfig(Config{})
	if err != nil {
		t.Fatal(err)
	}
	if client == nil {
		t.Fatal("nil client")
	}
	service, err := NewServiceFromConfig(Config{})
	if err != nil {
		t.Fatal(err)
	}
	if service == nil {
		t.Fatal("nil service")
	}
}
