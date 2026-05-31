package main

import (
	"context"
	"encoding/json"
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
	once := flag.String("once", "", "run one search query and print JSON instead of starting the HTTP server")
	page := flag.Int("page", 1, "search page for -once")
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

	client, err := websurfx.NewFromConfig(cfg)
	if err != nil {
		slog.Error("build client", "error", err)
		os.Exit(1)
	}

	if *once != "" {
		response := client.Search(context.Background(), websurfx.Query{Text: *once, Page: *page})
		encoder := json.NewEncoder(os.Stdout)
		encoder.SetIndent("", "  ")
		if err := encoder.Encode(response); err != nil {
			slog.Error("write response", "error", err)
			os.Exit(1)
		}
		return
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
