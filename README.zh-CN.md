# Websurfx Go

Websurfx Go 是一个中文优先的网页搜索库和普通可执行程序，行为上参考 Rust 版 Websurfx，但刻意保持轻量，方便长期维护。

它的定位很明确：

- 可以编译成一个普通 exe 直接运行。
- 可以作为 Go 包 `github.com/szStarWave/websurfx-go` 被其他 Go 程序引用。
- 不依赖 Redis。
- 不提供 Docker 部署文件。
- 第一阶段只做网页搜索，不做图片、视频、新闻等垂直搜索。

## 功能概览

- 内置默认配置启用全部已实现引擎：Bing 中文、360 搜索、搜狗、中文 Wikipedia、DuckDuckGo、Brave、Qwant、Startpage、Yahoo、Searx。
- 如果你想使用更小的中文优先集合，可以调用 `ChineseDefaultEngines()`。
- 提供首页、结果页、JSON API、`/search` 兼容路由、OpenSearch、健康检查、robots.txt、about 页和只读 settings 页。
- 支持内存缓存、代理、User-Agent 策略、简单 allowlist/blocklist 过滤、可选 CORS、gzip 压缩、基础 Cache-Control、进程内 HTTP 限流。
- 上游失败会返回在 `engineErrorsInfo` 中，不再伪装成“没有结果”。

## 快速运行

```powershell
go test ./...
go run ./cmd/websearch -config config.yaml
```

打开页面：

```text
http://127.0.0.1:8090/
```

JSON API：

```text
http://127.0.0.1:8090/api/search?q=中国
http://127.0.0.1:8090/search?q=中国&json=true
```

构建普通 exe：

```powershell
go build -o bin/websurfx-go.exe ./cmd/websearch
```

在命令行里执行一次搜索并输出 JSON：

```powershell
go run ./cmd/websearch -config config.yaml -once "中国" -page 1
```

## 配置文件

程序通过 `-config` 读取 YAML：

```powershell
go run ./cmd/websearch -config config.yaml
```

配置文件路径可以完全自定义。如果文件不存在，Websurfx Go 会使用内置默认配置，并启用当前已经实现的全部搜索引擎。

最小配置：

```yaml
server:
  address: "127.0.0.1:8090"
search:
  engines:
    - bing
    - so360
    - sogou
    - zhwikipedia
    - duckduckgo
    - brave
    - qwant
    - startpage
    - yahoo
    - searx
```

完整示例：

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

## 配置字段说明

- `server.address`：HTTP 监听地址，例如 `127.0.0.1:8090`。
- `server.log_structured`：为 `true` 时输出 JSON 日志，为 `false` 时输出普通文本日志。
- `server.cors`：是否为 HTTP API 增加宽松 CORS 头，适合浏览器前端调用。
- `server.compression`：客户端支持 gzip 时压缩响应。
- `server.cache_headers`：添加基础 `Cache-Control` 响应头。
- `server.rate_limit.enabled`：是否启用进程内限流。
- `server.rate_limit.requests_per_minute`：每个客户端 IP 每分钟允许的请求数。
- `search.timeout`：上游搜索请求超时。
- `search.request_timeout`：`timeout` 的清晰别名，两个字段都会被接受。
- `search.cache_ttl`：内存缓存有效期。
- `search.proxy_url`：可选 HTTP 代理，例如 `http://127.0.0.1:7890`。
- `search.user_agent_policy`：可选 `desktop`、`mobile`，也可以写完整自定义 User-Agent 字符串。
- `search.engines`：启用的搜索引擎列表。
- `search.filters.allowlist`：可选字符串列表，结果 URL、标题或摘要必须命中其中之一。
- `search.filters.blocklist`：可选字符串列表，结果 URL、标题或摘要命中后会被过滤。

## 搜索引擎名称

如果你只想保留中文优先集合，可以这样写：

```yaml
engines:
  - bing
  - so360
  - sogou
  - zhwikipedia
```

可选引擎：

```text
bing
so360
sogou
zhwikipedia
duckduckgo
brave
qwant
startpage
yahoo
searx
searx:https://your-searx-instance.example
```

`searx` 默认使用内置实例地址。更推荐写成 `searx:https://your-searx-instance.example`，方便你选择自己信任且稳定的实例。

## 作为 Go 库使用

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

`NewFromConfigFile` 可以传任意路径。传空字符串或文件不存在时，会使用 `DefaultConfig()`，它默认启用 `AllEngines()`。如果你想保留更小的中文优先集合，可以使用 `ChineseDefaultEngines()`。

如果你想把它嵌到自己的 HTTP 服务里：

```go
handler := client.Handler()
```

注意：`Client.Search` 是纯搜索调用，不会套用 CORS、gzip、cache header、rate limit 等 HTTP 中间件。这些只作用于 `Client.Handler()` 返回的 HTTP handler。

可直接运行的示例：

- `examples/basic`：作为库执行一次搜索。
- `examples/custom-config`：在 Go 代码里自定义配置，并使用中文优先引擎集合。
- `examples/http-server`：把 HTTP handler 嵌入你自己的 Go 服务。

## JSON 响应格式

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

错误类型：

- `RequestError`：网络请求失败、HTTP 状态异常、上游超时等。
- `EmptyResultSet`：上游明确返回无结果。
- `UnexpectedError`：HTML/JSON 解析失败，或页面结构和选择器不匹配。

## 当前不做的事情

- 不接 Redis。
- 不提供 Docker。
- 不做图片、视频、新闻搜索。
- 不做 Cookie 持久化设置页。
- 不做主题系统。
- 不做复杂 SafeSearch 分级。
