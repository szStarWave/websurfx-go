package main

import (
	"context"
	"fmt"

	websurfx "github.com/szStarWave/websurfx-go"
)

func main() {
	client, err := websurfx.NewFromConfigFile("")
	if err != nil {
		panic(err)
	}

	response := client.Search(context.Background(), websurfx.Query{Text: "中国", Page: 1})
	for _, result := range response.Results {
		fmt.Printf("%s\n%s\n\n", result.Title, result.URL)
	}
	for _, engineErr := range response.EngineErrorsInfo {
		fmt.Printf("engine error: %s %s %s\n", engineErr.Engine, engineErr.Kind, engineErr.Message)
	}
}
