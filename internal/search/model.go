package search

import (
	"context"
	"net/http"
	"time"
)

type Query struct {
	Text string
	Page int
}

type Result struct {
	Title       string   `json:"title"`
	URL         string   `json:"url"`
	Description string   `json:"description"`
	Engine      []string `json:"engine"`
}

type EngineError struct {
	Engine  string `json:"engine"`
	Kind    string `json:"kind"`
	Message string `json:"message,omitempty"`
}

type Response struct {
	Query            string        `json:"query"`
	Page             int           `json:"page"`
	HasNextPage      bool          `json:"hasNextPage"`
	Engines          []string      `json:"engines"`
	Results          []Result      `json:"results"`
	EngineErrorsInfo []EngineError `json:"engineErrorsInfo"`
	Errors           []EngineError `json:"-"`
	Cached           bool          `json:"cached"`
	Duration         string        `json:"duration"`
}

type Engine interface {
	Name() string
	Search(ctx context.Context, client *http.Client, query Query) ([]Result, *EngineError)
}

const (
	ErrorEmptyResult = "EmptyResultSet"
	ErrorRequest     = "RequestError"
	ErrorUnexpected  = "UnexpectedError"
)

type FilterOptions struct {
	Allowlist []string
	Blocklist []string
}

type contextKey string

const userAgentKey contextKey = "user-agent"

func WithUserAgent(ctx context.Context, userAgent string) context.Context {
	if userAgent == "" {
		return ctx
	}
	return context.WithValue(ctx, userAgentKey, userAgent)
}

func UserAgentFromContext(ctx context.Context) string {
	value, _ := ctx.Value(userAgentKey).(string)
	if value == "" {
		return "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/125.0 Safari/537.36"
	}
	return value
}

type Service interface {
	Search(ctx context.Context, query Query) Response
}

func NormalizeQuery(q Query) Query {
	if q.Page < 1 {
		q.Page = 1
	}
	return q
}

func Since(start time.Time) string {
	return time.Since(start).Round(time.Millisecond).String()
}
