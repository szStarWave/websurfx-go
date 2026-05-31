package main

import (
	"context"
	"fmt"
	"time"

	websurfx "github.com/szStarWave/websurfx-go"
)

func main() {
	cfg := websurfx.DefaultConfig()
	cfg.Search.Engines = websurfx.ChineseDefaultEngines()
	cfg.Search.Timeout = 8 * time.Second
	cfg.Search.CacheTTL = 2 * time.Minute
	cfg.Search.Filters.Blocklist = []string{"example-spam-domain"}

	client, err := websurfx.NewFromConfig(cfg)
	if err != nil {
		panic(err)
	}

	response := client.Search(context.Background(), websurfx.Query{Text: "特朗普", Page: 1})
	fmt.Printf("results=%d errors=%d cached=%v\n", len(response.Results), len(response.EngineErrorsInfo), response.Cached)
}
