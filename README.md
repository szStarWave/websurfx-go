# Websurfx Go

**中文文档:** [README.zh-CN.md](README.zh-CN.md)

Websurfx Go is a Chinese-first web search library and normal executable inspired by the Rust Websurfx project. It is designed to stay small enough to maintain long term: no Redis, no Docker, and no required service stack.

## Features

- Runs as a single executable.
- Can be embedded as the Go package `github.com/szStarWave/websurfx-go`.
- Web search only.
- Built-in default config enables every implemented engine: Bing Chinese, 360 Search, Sogou, Shenma, Chinese Wikipedia, DuckDuckGo, Brave, Qwant, Startpage, Yahoo, and Searx.
- `ChineseDefaultEngines()` is available when you want the smaller Chinese-first set.
- JSON API, `/search` compatibility route, OpenSearch metadata, health check, robots.txt, about page, and read-only settings page.
- In-memory cache, proxy support, configurable User-Agent policy, simple allow/block filters, optional CORS, gzip compression, cache headers, and in-process HTTP rate limiting.
- Upstream failures are returned as `engineErrorsInfo` instead of being hidden as empty results.

## Quick Start

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

Run a one-off search from the terminal:

```powershell
go run ./cmd/websearch -config config.yaml -once "中国" -page 1
```

## Configuration

The executable reads a YAML file with `-config`:

```powershell
go run ./cmd/websearch -config config.yaml
```

The config path is fully customizable. If the file does not exist, Websurfx Go falls back to a built-in default config that enables every implemented engine.

Minimal config:

```yaml
server:
  address: "127.0.0.1:8090"
search:
  engines:
    - bing
    - so360
    - sogou
    - shenma
    - zhwikipedia
    - duckduckgo
    - brave
    - qwant
    - startpage
    - yahoo
    - searx
```

Full example:

```yaml
server:
  address: "127.0.0.1:8090"
  log_structured: true
  cors: false
  compression: true
  cache_headers: true
  rate_limit:
    enabled: true
    requests_per_minute: 60

search:
  timeout: 10s
  request_timeout: 10s
  cache_ttl: 5m
  proxy_url: ""
  user_agent_policy: desktop
  engines:
    - bing
    - so360
    - sogou
    - shenma
    - zhwikipedia
    - duckduckgo
    - brave
    - qwant
    - startpage
    - yahoo
    - searx
  filters:
    allowlist: []
    blocklist: []
```

Supported engine names:

```text
bing
so360
sogou
shenma
zhwikipedia
duckduckgo
brave
qwant
startpage
yahoo
searx
searx:https://your-searx-instance.example
```

Configuration notes:

- `server.address`: HTTP listen address.
- `server.log_structured`: write JSON logs when true, text logs when false.
- `server.cors`: add permissive CORS headers for HTTP API use from browsers.
- `server.compression`: gzip responses when the client supports it.
- `server.cache_headers`: add basic `Cache-Control` headers.
- `server.rate_limit.enabled`: enable in-process per-client HTTP rate limiting.
- `server.rate_limit.requests_per_minute`: request budget per client IP.
- `search.timeout` / `search.request_timeout`: upstream HTTP timeout. `request_timeout` is accepted as a clearer alias.
- `search.cache_ttl`: in-memory cache lifetime.
- `search.proxy_url`: optional HTTP proxy URL, for example `http://127.0.0.1:7890`.
- `search.user_agent_policy`: `desktop`, `mobile`, or a custom User-Agent string.
- `search.engines`: enabled engines in request fan-out order.
- `search.filters.allowlist`: optional substrings that result URL/title/description must contain.
- `search.filters.blocklist`: optional substrings that drop matching results.

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
    client, err := websurfx.NewFromConfigFile("my-config.yaml")
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

`NewFromConfigFile` accepts any path. Passing an empty path or a path that does not exist uses `DefaultConfig()`, which enables `AllEngines()`. If you want the smaller Chinese-first set, use `ChineseDefaultEngines()`.

You can also use the HTTP handler from a Go application:

```go
handler := client.Handler()
```

`Client.Search` does not apply HTTP-only middleware such as CORS, compression, cache headers, or rate limiting. Those only apply to `Client.Handler`.

Runnable examples:

- `examples/basic`: run one library search.
- `examples/custom-config`: create a config in Go code and use the Chinese-first engine set.
- `examples/http-server`: embed the HTTP handler in your own Go server.

## API Response

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

Error kinds:

- `RequestError`: HTTP/network/upstream status failure.
- `EmptyResultSet`: the upstream clearly reported no results.
- `UnexpectedError`: parser or page structure did not match expectations.

## Deliberately Out Of Scope

- Redis.
- Docker deployment files.
- Image, video, or news search.
- Cookie-backed settings.
- Theme system.
- Complex safe-search tiers.
