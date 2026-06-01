package config

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestLoadMissingFileReturnsDefaultWithAllEngines(t *testing.T) {
	cfg, err := Load(filepath.Join(t.TempDir(), "missing.yaml"))
	if err != nil {
		t.Fatal(err)
	}
	if cfg.Server.Address != "127.0.0.1:8090" {
		t.Fatalf("unexpected address %q", cfg.Server.Address)
	}
	if cfg.Search.Timeout != 10*time.Second {
		t.Fatalf("unexpected timeout %s", cfg.Search.Timeout)
	}
	if len(cfg.Search.Engines) != 14 {
		t.Fatalf("expected all engines, got %v", cfg.Search.Engines)
	}
	if cfg.Search.Engines[0] != "bing" || cfg.Search.Engines[len(cfg.Search.Engines)-1] != "searx" {
		t.Fatalf("unexpected engine order %v", cfg.Search.Engines)
	}
}

func TestLoadEmptyPathReturnsDefault(t *testing.T) {
	cfg, err := Load("")
	if err != nil {
		t.Fatal(err)
	}
	if len(cfg.Search.Engines) != 14 {
		t.Fatalf("expected all engines, got %v", cfg.Search.Engines)
	}
}

func TestLoadCustomPath(t *testing.T) {
	path := filepath.Join(t.TempDir(), "custom.yaml")
	data := []byte(`
server:
  address: "127.0.0.1:18080"
search:
  timeout: 2s
  cache_ttl: 30s
  engines:
    - bing
`)
	if err := os.WriteFile(path, data, 0o600); err != nil {
		t.Fatal(err)
	}
	cfg, err := Load(path)
	if err != nil {
		t.Fatal(err)
	}
	if cfg.Server.Address != "127.0.0.1:18080" {
		t.Fatalf("unexpected address %q", cfg.Server.Address)
	}
	if cfg.Search.Timeout != 2*time.Second || cfg.Search.CacheTTL != 30*time.Second {
		t.Fatalf("unexpected durations timeout=%s cache=%s", cfg.Search.Timeout, cfg.Search.CacheTTL)
	}
	if len(cfg.Search.Engines) != 1 || cfg.Search.Engines[0] != "bing" {
		t.Fatalf("unexpected engines %v", cfg.Search.Engines)
	}
}

func TestValidateRateLimit(t *testing.T) {
	cfg := Default()
	cfg.Server.RateLimit.Enabled = true
	cfg.Server.RateLimit.RequestsPerMinute = -1
	if err := Validate(cfg); err == nil {
		t.Fatal("expected validation error")
	}
}
