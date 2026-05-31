package httpapi

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/szStarWave/websurfx-go/internal/search"
)

type fakeService struct{}

func (fakeService) Search(_ context.Context, query search.Query) search.Response {
	return search.Response{
		Query:       query.Text,
		Page:        query.Page,
		HasNextPage: true,
		Results: []search.Result{{
			Title:       "China news",
			URL:         "https://example.com/china",
			Description: "Summary",
			Engine:      []string{"fixture"},
		}},
		EngineErrorsInfo: []search.EngineError{{
			Engine:  "broken",
			Kind:    search.ErrorRequest,
			Message: "timeout",
		}},
		Duration: "1ms",
	}
}

func TestSearchJSONRoutesMatch(t *testing.T) {
	handler := mustRoutes(t)
	apiResp := request(t, handler, "/api/search?q=china&page=2")
	searchResp := request(t, handler, "/search?q=china&page=2&json=true")
	if apiResp.Code != http.StatusOK || searchResp.Code != http.StatusOK {
		t.Fatalf("unexpected status api=%d search=%d", apiResp.Code, searchResp.Code)
	}
	if apiResp.Body.String() != searchResp.Body.String() {
		t.Fatalf("JSON routes differ:\napi=%s\nsearch=%s", apiResp.Body.String(), searchResp.Body.String())
	}
	var parsed search.Response
	if err := json.Unmarshal(apiResp.Body.Bytes(), &parsed); err != nil {
		t.Fatal(err)
	}
	if parsed.Page != 2 || !parsed.HasNextPage || len(parsed.EngineErrorsInfo) != 1 {
		t.Fatalf("unexpected response %#v", parsed)
	}
}

func TestBasicHTTPRoutes(t *testing.T) {
	handler := mustRoutes(t)
	cases := map[string]int{
		"/api/search":     http.StatusBadRequest,
		"/healthz":        http.StatusOK,
		"/robots.txt":     http.StatusOK,
		"/opensearch.xml": http.StatusOK,
		"/missing":        http.StatusNotFound,
	}
	for path, want := range cases {
		got := request(t, handler, path)
		if got.Code != want {
			t.Fatalf("%s: want %d got %d body=%s", path, want, got.Code, got.Body.String())
		}
	}
}

func TestIndexShowsErrorsAndPagination(t *testing.T) {
	handler := mustRoutes(t)
	resp := request(t, handler, "/search?q=china&page=2")
	body := resp.Body.String()
	if resp.Code != http.StatusOK {
		t.Fatalf("unexpected status %d", resp.Code)
	}
	for _, want := range []string{"部分搜索引擎失败", "下一页", "上一页", "China news"} {
		if !strings.Contains(body, want) {
			t.Fatalf("page missing %q: %s", want, body)
		}
	}
}

func TestRateLimit(t *testing.T) {
	handler := WithRateLimit(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	}), 1)
	if first := request(t, handler, "/"); first.Code != http.StatusOK {
		t.Fatalf("first request status %d", first.Code)
	}
	if second := request(t, handler, "/"); second.Code != http.StatusTooManyRequests {
		t.Fatalf("second request status %d", second.Code)
	}
}

func mustRoutes(t *testing.T) http.Handler {
	t.Helper()
	server, err := New(fakeService{})
	if err != nil {
		t.Fatal(err)
	}
	return server.Routes()
}

func request(t *testing.T, handler http.Handler, path string) *httptest.ResponseRecorder {
	t.Helper()
	req := httptest.NewRequest(http.MethodGet, path, nil)
	req.RemoteAddr = "127.0.0.1:12345"
	resp := httptest.NewRecorder()
	handler.ServeHTTP(resp, req)
	return resp
}
