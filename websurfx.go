package websurfx

import (
	"net/http"
	"time"

	"github.com/szStarWave/websurfx-go/internal/cache"
	"github.com/szStarWave/websurfx-go/internal/config"
	"github.com/szStarWave/websurfx-go/internal/engine"
	"github.com/szStarWave/websurfx-go/internal/httpapi"
	"github.com/szStarWave/websurfx-go/internal/search"
)

type Query = search.Query
type Result = search.Result
type EngineError = search.EngineError
type Response = search.Response
type Engine = search.Engine
type Service = search.Service

const (
	ErrorEmptyResult = search.ErrorEmptyResult
	ErrorRequest     = search.ErrorRequest
	ErrorParse       = search.ErrorParse
	ErrorUnexpected  = search.ErrorUnexpected
)

type Config = config.Config
type ServerConfig = config.ServerConfig
type SearchConfig = config.SearchConfig

func LoadConfig(path string) (Config, error) {
	return config.Load(path)
}

func DefaultEngines() []string {
	return []string{"bing", "so360", "sogou", "zhwikipedia"}
}

func BuildEngines(names []string) ([]Engine, error) {
	return engine.Build(names)
}

func NewService(timeout time.Duration, cacheTTL time.Duration, engines []Engine) Service {
	return search.NewAggregator(timeout, cache.NewMemory(cacheTTL), engines)
}

func NewServiceFromConfig(cfg Config) (Service, error) {
	engines, err := BuildEngines(cfg.Search.Engines)
	if err != nil {
		return nil, err
	}
	return NewService(cfg.Search.Timeout, cfg.Search.CacheTTL, engines), nil
}

func NewHTTPHandler(service Service) (http.Handler, error) {
	server, err := httpapi.New(service)
	if err != nil {
		return nil, err
	}
	return server.Routes(), nil
}
