package main

import (
	"flag"
	"log/slog"
	"net/http"
	"os"

	"github.com/szStarWave/websurfx-go/internal/cache"
	"github.com/szStarWave/websurfx-go/internal/config"
	"github.com/szStarWave/websurfx-go/internal/engine"
	"github.com/szStarWave/websurfx-go/internal/httpapi"
	"github.com/szStarWave/websurfx-go/internal/search"
)

func main() {
	configPath := flag.String("config", "config.yaml", "path to YAML config")
	flag.Parse()

	cfg, err := config.Load(*configPath)
	if err != nil {
		slog.Error("load config", "error", err)
		os.Exit(1)
	}

	engines, err := engine.Build(cfg.Search.Engines)
	if err != nil {
		slog.Error("build engines", "error", err)
		os.Exit(1)
	}

	service := search.NewAggregator(
		cfg.Search.Timeout,
		cache.NewMemory(cfg.Search.CacheTTL),
		engines,
	)
	server, err := httpapi.New(service)
	if err != nil {
		slog.Error("build server", "error", err)
		os.Exit(1)
	}

	slog.Info("starting websurfx go v2", "address", cfg.Server.Address)
	if err := http.ListenAndServe(cfg.Server.Address, server.Routes()); err != nil {
		slog.Error("server stopped", "error", err)
		os.Exit(1)
	}
}
