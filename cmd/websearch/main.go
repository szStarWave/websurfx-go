package main

import (
	"context"
	"flag"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

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
	if cfg.Server.LogStructured {
		slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stderr, nil)))
	} else {
		slog.SetDefault(slog.New(slog.NewTextHandler(os.Stderr, nil)))
	}

	engines, err := websurfx.BuildEngines(cfg.Search.Engines)
	if err != nil {
		slog.Error("build engines", "error", err)
		os.Exit(1)
	}
	client, err := websurfx.New(websurfx.Options{
		Engines:         engines,
		Timeout:         cfg.Search.Timeout,
		CacheTTL:        cfg.Search.CacheTTL,
		ProxyURL:        cfg.Search.ProxyURL,
		UserAgentPolicy: cfg.Search.UserAgentPolicy,
		RateLimit:       cfg.Server.RateLimit,
		CORS:            cfg.Server.CORS,
		Compression:     cfg.Server.Compression,
		CacheHeaders:    cfg.Server.CacheHeaders,
		Filters: websurfx.FilterOptions{
			Allowlist: cfg.Search.Filters.Allowlist,
			Blocklist: cfg.Search.Filters.Blocklist,
		},
	})
	if err != nil {
		slog.Error("build client", "error", err)
		os.Exit(1)
	}

	server := &http.Server{
		Addr:              cfg.Server.Address,
		Handler:           client.Handler(),
		ReadHeaderTimeout: 5 * time.Second,
	}
	stop, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	go func() {
		<-stop.Done()
		ctx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer shutdownCancel()
		if err := server.Shutdown(ctx); err != nil {
			slog.Error("shutdown server", "error", err)
		}
	}()

	slog.Info("starting websurfx go v2", "address", cfg.Server.Address)
	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		slog.Error("server stopped", "error", err)
		os.Exit(1)
	}
}
