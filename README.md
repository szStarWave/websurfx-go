# Websurfx Go

A small Chinese-first web search library and executable. It uses the Rust Websurfx implementation as a behavior reference, but intentionally keeps the first Go version compact and easy to embed.

## Scope

- Web search only.
- Default engines: Bing Chinese, 360 Search, Sogou, and Chinese Wikipedia.
- JSON API and a minimal server-rendered frontend when used as an exe.
- Public Go package for embedding in other Go programs.
- YAML configuration for the executable.
- In-memory cache only.
- Engine failures are returned in responses instead of being hidden as "no results".

Out of scope for this version: image/video/news search, user cookie settings, theme system, complex safe search, Redis, and Docker.

## Run As An Exe

```powershell
go test ./...
go run ./cmd/websearch -config config.yaml
```

Open:

```text
http://127.0.0.1:8090/
```

JSON API:

```text
http://127.0.0.1:8090/api/search?q=中国
```

Build a normal executable:

```powershell
go build -o websurfx-go.exe ./cmd/websearch
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
    engines, err := websurfx.BuildEngines(websurfx.DefaultEngines())
    if err != nil {
        panic(err)
    }

    svc := websurfx.NewService(10*time.Second, 5*time.Minute, engines)
    resp := svc.Search(context.Background(), websurfx.Query{Text: "中国", Page: 1})

    for _, result := range resp.Results {
        fmt.Println(result.Title, result.URL)
    }
    for _, err := range resp.Errors {
        fmt.Println(err.Engine, err.Type, err.Detail)
    }
}
```

## Response Shape

```json
{
  "query": "中国",
  "page": 1,
  "results": [
    {
      "title": "中国新闻",
      "url": "https://example.com",
      "description": "摘要",
      "engine": ["sogou"]
    }
  ],
  "errors": [
    {
      "engine": "zhwikipedia",
      "type": "RequestError",
      "detail": "timeout"
    }
  ],
  "cached": false,
  "duration": "1.2s"
}
```