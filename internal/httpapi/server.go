package httpapi

import (
	"encoding/json"
	"html/template"
	"net/http"
	"strconv"
	"strings"

	"github.com/szStarWave/websurfx-go/internal/search"
)

type Server struct {
	service search.Service
	index   *template.Template
}

func New(service search.Service) (*Server, error) {
	index, err := template.New("index").Parse(indexHTML)
	if err != nil {
		return nil, err
	}
	return &Server{service: service, index: index}, nil
}

func (s *Server) Routes() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/", s.indexHandler)
	mux.HandleFunc("/api/search", s.apiSearchHandler)
	return mux
}

func (s *Server) apiSearchHandler(w http.ResponseWriter, r *http.Request) {
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
	}{
		Query:    query.Text,
		Response: response,
		HasQuery: strings.TrimSpace(query.Text) != "",
	})
}

func parseQuery(r *http.Request) search.Query {
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	return search.NormalizeQuery(search.Query{
		Text: strings.TrimSpace(r.URL.Query().Get("q")),
		Page: page,
	})
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
  <form action="/" method="get">
    <input name="q" value="{{.Query}}" placeholder="输入中文关键词" autofocus>
    <button type="submit">搜索</button>
  </form>

  {{if .HasQuery}}
    <p class="meta">找到 {{len .Response.Results}} 条结果，用时 {{.Response.Duration}}{{if .Response.Cached}}，来自缓存{{end}}</p>
    {{if .Response.Errors}}
      <section class="errors">
        <h2>部分搜索引擎失败</h2>
        {{range .Response.Errors}}
          <div class="error">{{.Engine}}: {{.Type}}{{if .Detail}} - {{.Detail}}{{end}}</div>
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
  {{else}}
    <p class="meta">默认引擎：Bing 中文、360 搜索、搜狗、中文 Wikipedia。当前只做网页搜索。</p>
  {{end}}
</main>
</body>
</html>`
