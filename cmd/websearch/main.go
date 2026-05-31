package main

import (
	"flag"
	"log/slog"
	"net/http"
	"os"

	websurfx "github.com/szStarWave/websurfx-go"
)

func main() {
	configPath := flag.String("config", "config.yaml", "path to YAML config")
	flag.Parse()

	cfg, err := websurfx.LoadConfig(*configPath)
	if err != nil {
		slog.Error("load config", "error", err)
		os.Exit(1)
	}

	service, err := websurfx.NewServiceFromConfig(cfg)
	if err != nil {
		slog.Error("build service", "error", err)
		os.Exit(1)
	}
	handler, err := websurfx.NewHTTPHandler(service)
	if err != nil {
		slog.Error("build http handler", "error", err)
		os.Exit(1)
	}

	slog.Info("starting websurfx go v2", "address", cfg.Server.Address)
	if err := http.ListenAndServe(cfg.Server.Address, handler); err != nil {
		slog.Error("server stopped", "error", err)
		os.Exit(1)
	}
}
