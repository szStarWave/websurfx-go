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
	Engine string `json:"engine"`
	Type   string `json:"type"`
	Detail string `json:"detail,omitempty"`
}

type Response struct {
	Query    string        `json:"query"`
	Page     int           `json:"page"`
	Results  []Result      `json:"results"`
	Errors   []EngineError `json:"errors"`
	Cached   bool          `json:"cached"`
	Duration string        `json:"duration"`
}

type Engine interface {
	Name() string
	Search(ctx context.Context, client *http.Client, query Query) ([]Result, *EngineError)
}

const (
	ErrorEmptyResult = "EmptyResultSet"
	ErrorRequest     = "RequestError"
	ErrorParse       = "ParseError"
	ErrorUnexpected  = "UnexpectedError"
)

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
