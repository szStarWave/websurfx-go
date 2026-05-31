package main

import (
	"log"
	"net/http"

	websurfx "github.com/szStarWave/websurfx-go"
)

func main() {
	client, err := websurfx.NewFromConfigFile("")
	if err != nil {
		log.Fatal(err)
	}

	log.Println("listening on http://127.0.0.1:8091")
	log.Fatal(http.ListenAndServe("127.0.0.1:8091", client.Handler()))
}
