# Websurfx Go v2

This is a small Chinese-first web search prototype that uses the Rust implementation as a behavior reference without porting every feature.

## Scope

- Web search only.
- Default engines: Bing Chinese, 360 Search, Sogou, and Chinese Wikipedia.
- JSON API and a minimal server-rendered frontend.
- YAML configuration.
- In-memory cache.
- Engine failures are returned in responses instead of being hidden as "no results".

Out of scope for this first version: image/video/news search, user cookie settings, theme system, complex safe search, and Redis.

## Run

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
