# Websurfx Go

A Chinese-first web search library and executable inspired by the Rust Websurfx project. This Go version is intentionally compact: it runs as a normal exe and can also be embedded as `github.com/szStarWave/websurfx-go` from other Go programs.

## Scope

- Web search only.
- Default engines: Bing Chinese, 360 Search, Sogou, and Chinese Wikipedia.
- Optional engines: DuckDuckGo, Brave, Qwant, Startpage, Yahoo, and configurable Searx.
- JSON API, `/search` compatibility route, and a minimal server-rendered frontend.
- In-memory cache, proxy support, simple allow/block filters, and in-process HTTP rate limiting.
- No Redis, no Docker, no image/video/news search, no cookie settings, no theme system, and no complex safe-search tiers.

## Run As An Exe

```powershell
go test ./...
go run ./cmd/websearch -config config.yaml
```

Open:

```text
http://127.0.0.1:8090/
```

JSON APIs:

```text
http://127.0.0.1:8090/api/search?q=中国
http://127.0.0.1:8090/search?q=中国&json=true
```

Build a normal executable:

```powershell
go build -o bin/websurfx-go.exe ./cmd/websearch
```

## Use As A Library

```go
package main

import (
    "context"
    "fmt"
    "time"

    websurfx "github.com/szStarWave/websurfx-go"
)

func main() {
    client, err := websurfx.New(websurfx.Options{
        Timeout:  10 * time.Second,
        CacheTTL: 5 * time.Minute,
    })
    if err != nil {
        panic(err)
    }

    resp := client.Search(context.Background(), websurfx.Query{Text: "中国", Page: 1})
    for _, result := range resp.Results {
        fmt.Println(result.Title, result.URL)
    }
    for _, err := range resp.EngineErrorsInfo {
        fmt.Println(err.Engine, err.Kind, err.Message)
    }
}
```

## Response Shape

```json
{
  "query": "中国",
  "page": 1,
  "hasNextPage": true,
  "results": [
    {
      "title": "中国新闻",
      "url": "https://example.com",
      "description": "摘要",
      "engine": ["sogou"]
    }
  ],
  "engineErrorsInfo": [
    {
      "engine": "zhwikipedia",
      "kind": "RequestError",
      "message": "timeout"
    }
  ],
  "cached": false,
  "duration": "1.2s"
}
```
