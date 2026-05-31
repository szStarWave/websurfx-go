package httpapi

import (
	"compress/gzip"
	"encoding/json"
	"html/template"
	"io"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/szStarWave/websurfx-go/internal/search"
)

type Server struct {
	service     search.Service
	index       *template.Template
	engineNames []string
}

type engineNameProvider interface {
	EngineNames() []string
}

func New(service search.Service) (*Server, error) {
	index, err := template.New("index").Funcs(template.FuncMap{
		"plus":  func(a, b int) int { return a + b },
		"minus": func(a, b int) int { return a - b },
	}).Parse(indexHTML)
	if err != nil {
		return nil, err
	}
	var names []string
	if provider, ok := service.(engineNameProvider); ok {
		names = provider.EngineNames()
	}
	return &Server{service: service, index: index, engineNames: names}, nil
}

func (s *Server) Routes() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/", s.indexHandler)
	mux.HandleFunc("/api/search", s.apiSearchHandler)
	mux.HandleFunc("/search", s.searchHandler)
	mux.HandleFunc("/healthz", s.healthHandler)
	mux.HandleFunc("/robots.txt", s.robotsHandler)
	mux.HandleFunc("/opensearch.xml", s.openSearchHandler)
	mux.HandleFunc("/about", s.aboutHandler)
	mux.HandleFunc("/settings", s.settingsHandler)
	return mux
}

func (s *Server) apiSearchHandler(w http.ResponseWriter, r *http.Request) {
	s.writeJSONSearch(w, r)
}

func (s *Server) searchHandler(w http.ResponseWriter, r *http.Request) {
	if r.URL.Query().Get("json") == "true" {
		s.writeJSONSearch(w, r)
		return
	}
	s.indexHandler(w, r)
}

func (s *Server) writeJSONSearch(w http.ResponseWriter, r *http.Request) {
	query := parseQuery(r)
	if strings.TrimSpace(query.Text) == "" {
		http.Error(w, "missing q query parameter", http.StatusBadRequest)
		return
	}
	response := s.service.Search(r.Context(), query)
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	_ = json.NewEncoder(w).Encode(response)
}

func (s *Server) indexHandler(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" && r.URL.Path != "/search" {
		s.notFoundHandler(w, r)
		return
	}
	query := parseQuery(r)
	var response search.Response
	if strings.TrimSpace(query.Text) != "" {
		response = s.service.Search(r.Context(), query)
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	_ = s.index.Execute(w, struct {
		Query    string
		Response search.Response
		HasQuery bool
		Engines  []string
	}{
		Query:    query.Text,
		Response: response,
		HasQuery: strings.TrimSpace(query.Text) != "",
		Engines:  s.engineNames,
	})
}

func (s *Server) healthHandler(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	_, _ = w.Write([]byte(`{"ok":true}`))
}

func (s *Server) aboutHandler(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	_, _ = w.Write([]byte(aboutHTML))
}

func (s *Server) settingsHandler(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	_, _ = w.Write([]byte(settingsHTML))
}

func (s *Server) robotsHandler(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	_, _ = w.Write([]byte("User-agent: *\nDisallow: /search\nDisallow: /api/search\n"))
}

func (s *Server) openSearchHandler(w http.ResponseWriter, r *http.Request) {
	scheme := "http"
	if r.TLS != nil {
		scheme = "https"
	}
	base := scheme + "://" + r.Host
	w.Header().Set("Content-Type", "application/opensearchdescription+xml; charset=utf-8")
	_, _ = w.Write([]byte(`<?xml version="1.0" encoding="UTF-8"?>
<OpenSearchDescription xmlns="http://a9.com/-/spec/opensearch/1.1/">
  <ShortName>Websurfx Go</ShortName>
  <Description>Web search through Websurfx Go</Description>
  <InputEncoding>UTF-8</InputEncoding>
  <Url type="text/html" method="get" template="` + base + `/search?q={searchTerms}"/>
  <Url type="application/json" method="get" template="` + base + `/search?q={searchTerms}&amp;json=true"/>
</OpenSearchDescription>`))
}

func (s *Server) notFoundHandler(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusNotFound)
	_, _ = w.Write([]byte("<!doctype html><title>Not found</title><main><h1>Not found</h1><p>The requested page does not exist.</p></main>"))
}

func parseQuery(r *http.Request) search.Query {
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	return search.NormalizeQuery(search.Query{
		Text: strings.TrimSpace(r.URL.Query().Get("q")),
		Page: page,
	})
}

func WithRateLimit(next http.Handler, requestsPerMinute int) http.Handler {
	if requestsPerMinute <= 0 {
		return next
	}
	limiter := &ipLimiter{
		limit:  requestsPerMinute,
		window: time.Minute,
		hits:   map[string]rateEntry{},
	}
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !limiter.allow(clientIP(r)) {
			http.Error(w, "rate limit exceeded", http.StatusTooManyRequests)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func WithCORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func WithCacheHeaders(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/robots.txt", "/opensearch.xml":
			w.Header().Set("Cache-Control", "public, max-age=3600")
		case "/healthz", "/api/search", "/search":
			w.Header().Set("Cache-Control", "no-store")
		default:
			w.Header().Set("Cache-Control", "no-cache")
		}
		next.ServeHTTP(w, r)
	})
}

func WithCompression(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !strings.Contains(r.Header.Get("Accept-Encoding"), "gzip") {
			next.ServeHTTP(w, r)
			return
		}
		w.Header().Add("Vary", "Accept-Encoding")
		w.Header().Set("Content-Encoding", "gzip")
		gz := gzip.NewWriter(w)
		defer gz.Close()
		next.ServeHTTP(gzipResponseWriter{ResponseWriter: w, Writer: gz}, r)
	})
}

type gzipResponseWriter struct {
	http.ResponseWriter
	io.Writer
}

func (w gzipResponseWriter) Write(data []byte) (int, error) {
	return w.Writer.Write(data)
}

type ipLimiter struct {
	mu     sync.Mutex
	limit  int
	window time.Duration
	hits   map[string]rateEntry
}

type rateEntry struct {
	count int
	reset time.Time
}

func (l *ipLimiter) allow(key string) bool {
	now := time.Now()
	l.mu.Lock()
	defer l.mu.Unlock()
	entry := l.hits[key]
	if now.After(entry.reset) {
		entry = rateEntry{reset: now.Add(l.window)}
	}
	entry.count++
	l.hits[key] = entry
	return entry.count <= l.limit
}

func clientIP(r *http.Request) string {
	if forwarded := r.Header.Get("X-Forwarded-For"); forwarded != "" {
		return strings.TrimSpace(strings.Split(forwarded, ",")[0])
	}
	host := r.RemoteAddr
	if idx := strings.LastIndex(host, ":"); idx > -1 {
		return host[:idx]
	}
	return host
}

const indexHTML = `<!doctype html>
<html lang="zh-CN">
<head>
  <meta charset="utf-8">
  <meta name="viewport" content="width=device-width, initial-scale=1">
  <title>Websurfx Go 中文搜索</title>
  <style>
    :root { color-scheme: light dark; font-family: system-ui, -apple-system, BlinkMacSystemFont, "Segoe UI", sans-serif; }
    body { margin: 0; background: Canvas; color: CanvasText; }
    main { max-width: 920px; margin: 0 auto; padding: 32px 20px; }
    form { display: flex; gap: 8px; margin-bottom: 24px; }
    input { flex: 1; font-size: 16px; padding: 10px 12px; border: 1px solid #8886; border-radius: 6px; }
    button { font-size: 16px; padding: 10px 18px; border: 1px solid #8886; border-radius: 6px; cursor: pointer; }
    nav { display: flex; gap: 14px; margin: 22px 0; }
    .meta, .url, .engine, .error { color: #667085; font-size: 13px; }
    .result { padding: 16px 0; border-top: 1px solid #8883; }
    .result h2 { font-size: 20px; margin: 0 0 6px; }
    .result p { margin: 8px 0; line-height: 1.55; }
    .errors { margin: 18px 0; padding: 12px; border: 1px solid #f59e0b77; border-radius: 6px; background: #f59e0b12; }
    .errors h2 { font-size: 16px; margin: 0 0 8px; }
  </style>
</head>
<body>
<main>
  <h1>Websurfx Go 中文搜索</h1>
  <form action="/search" method="get">
    <input name="q" value="{{.Query}}" placeholder="输入中文关键词" autofocus>
    <button type="submit">搜索</button>
  </form>

  {{if .HasQuery}}
    <p class="meta">找到 {{len .Response.Results}} 条结果，用时 {{.Response.Duration}}{{if .Response.Cached}}，来自缓存{{end}}</p>
    {{if .Response.EngineErrorsInfo}}
      <section class="errors">
        <h2>部分搜索引擎失败</h2>
        {{range .Response.EngineErrorsInfo}}
          <div class="error">{{.Engine}}: {{.Kind}}{{if .Message}} - {{.Message}}{{end}}</div>
        {{end}}
      </section>
    {{end}}
    {{range .Response.Results}}
      <article class="result">
        <h2><a href="{{.URL}}" target="_blank" rel="noreferrer">{{.Title}}</a></h2>
        <div class="url">{{.URL}}</div>
        <p>{{.Description}}</p>
        <div class="engine">{{range .Engine}}{{.}} {{end}}</div>
      </article>
    {{else}}
      <p>没有结果。若上游失败，错误会显示在上方。</p>
    {{end}}
    <nav class="meta">
      {{if gt .Response.Page 1}}<a href="/search?q={{.Query}}&page={{minus .Response.Page 1}}">上一页</a>{{end}}
      <span>第 {{.Response.Page}} 页</span>
      {{if .Response.HasNextPage}}<a href="/search?q={{.Query}}&page={{plus .Response.Page 1}}">下一页</a>{{end}}
    </nav>
  {{else}}
    {{if .Engines}}
      <p class="meta">当前启用引擎：{{range .Engines}}<code>{{.}}</code> {{end}}</p>
    {{else}}
      <p class="meta">默认启用全部已实现网页搜索引擎；如需更小的中文优先集合，请在配置中只保留 Bing 中文、360 搜索、搜狗、中文 Wikipedia。</p>
    {{end}}
  {{end}}
</main>
</body>
</html>`

const aboutHTML = `<!doctype html>
<html lang="zh-CN">
<head><meta charset="utf-8"><meta name="viewport" content="width=device-width, initial-scale=1"><title>About Websurfx Go</title></head>
<body>
<main style="max-width:760px;margin:40px auto;padding:0 20px;font-family:system-ui,sans-serif;line-height:1.6">
  <h1>Websurfx Go</h1>
  <p>Websurfx Go is a lightweight web-search executable and embeddable Go library.</p>
  <p>It focuses on normal web search, Chinese-first defaults, clear upstream error reporting, and simple local operation.</p>
  <p><a href="/">Search</a> · <a href="/settings">Runtime settings</a> · <a href="/opensearch.xml">OpenSearch</a></p>
</main>
</body>
</html>`

const settingsHTML = `<!doctype html>
<html lang="zh-CN">
<head><meta charset="utf-8"><meta name="viewport" content="width=device-width, initial-scale=1"><title>Websurfx Go Settings</title></head>
<body>
<main style="max-width:760px;margin:40px auto;padding:0 20px;font-family:system-ui,sans-serif;line-height:1.6">
  <h1>Runtime settings</h1>
  <p>This Go edition is configured through <code>config.yaml</code> or the public <code>websurfx.Options</code> API.</p>
  <p>It does not store browser cookies or expose a mutable settings UI. Edit the YAML file and restart the exe to change engines, proxy, filters, timeout, compression, CORS, cache headers, or rate limiting.</p>
  <p><a href="/">Back to search</a></p>
</main>
</body>
</html>`
