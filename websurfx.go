package websurfx

import (
	"context"
	"net/http"
	"net/url"
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
type FilterOptions = search.FilterOptions

const (
	ErrorEmptyResult = search.ErrorEmptyResult
	ErrorRequest     = search.ErrorRequest
	ErrorUnexpected  = search.ErrorUnexpected
)

type Config = config.Config
type ServerConfig = config.ServerConfig
type SearchConfig = config.SearchConfig
type RateLimitOptions = config.RateLimitConfig

type Options struct {
	Engines         []Engine
	Timeout         time.Duration
	CacheTTL        time.Duration
	ProxyURL        string
	UserAgentPolicy string
	RateLimit       RateLimitOptions
	Filters         FilterOptions
	CORS            bool
	Compression     bool
	CacheHeaders    bool
}

type Client struct {
	service         Service
	rateLimit       RateLimitOptions
	userAgentPolicy string
	cors            bool
	compression     bool
	cacheHeaders    bool
}

type userAgentService struct {
	next      Service
	userAgent string
}

func (s userAgentService) Search(ctx context.Context, q Query) Response {
	return s.next.Search(search.WithUserAgent(ctx, s.userAgent), q)
}

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
	client, err := httpClient(cfg.Search.Timeout, cfg.Search.ProxyURL)
	if err != nil {
		return nil, err
	}
	service := search.NewAggregatorWithClient(client, cache.NewMemory(cfg.Search.CacheTTL), engines, search.FilterOptions{
		Allowlist: cfg.Search.Filters.Allowlist,
		Blocklist: cfg.Search.Filters.Blocklist,
	})
	return userAgentService{next: service, userAgent: userAgentForPolicy(cfg.Search.UserAgentPolicy)}, nil
}

func NewHTTPHandler(service Service) (http.Handler, error) {
	server, err := httpapi.New(service)
	if err != nil {
		return nil, err
	}
	return server.Routes(), nil
}

func New(opts Options) (*Client, error) {
	timeout := opts.Timeout
	if timeout <= 0 {
		timeout = 10 * time.Second
	}
	cacheTTL := opts.CacheTTL
	if cacheTTL <= 0 {
		cacheTTL = 5 * time.Minute
	}
	engines := opts.Engines
	if len(engines) == 0 {
		var err error
		engines, err = BuildEngines(DefaultEngines())
		if err != nil {
			return nil, err
		}
	}
	httpClient, err := httpClient(timeout, opts.ProxyURL)
	if err != nil {
		return nil, err
	}
	return &Client{
		service:         search.NewAggregatorWithClient(httpClient, cache.NewMemory(cacheTTL), engines, opts.Filters),
		rateLimit:       opts.RateLimit,
		userAgentPolicy: opts.UserAgentPolicy,
		cors:            opts.CORS,
		compression:     opts.Compression,
		cacheHeaders:    opts.CacheHeaders,
	}, nil
}

func (c *Client) Search(ctx context.Context, q Query) Response {
	return c.service.Search(search.WithUserAgent(ctx, userAgentForPolicy(c.userAgentPolicy)), q)
}

func (c *Client) Handler() http.Handler {
	handler, err := NewHTTPHandler(userAgentService{
		next:      c.service,
		userAgent: userAgentForPolicy(c.userAgentPolicy),
	})
	if err != nil {
		return http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		})
	}
	if c.rateLimit.Enabled {
		handler = httpapi.WithRateLimit(handler, c.rateLimit.RequestsPerMinute)
	}
	if c.cacheHeaders {
		handler = httpapi.WithCacheHeaders(handler)
	}
	if c.cors {
		handler = httpapi.WithCORS(handler)
	}
	if c.compression {
		handler = httpapi.WithCompression(handler)
	}
	return handler
}

func httpClient(timeout time.Duration, proxyURL string) (*http.Client, error) {
	transport := http.DefaultTransport.(*http.Transport).Clone()
	if proxyURL != "" {
		parsed, err := url.Parse(proxyURL)
		if err != nil {
			return nil, err
		}
		transport.Proxy = http.ProxyURL(parsed)
	}
	return &http.Client{Timeout: timeout, Transport: transport}, nil
}

func userAgentForPolicy(policy string) string {
	switch policy {
	case "desktop", "":
		return "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/125.0 Safari/537.36"
	case "mobile":
		return "Mozilla/5.0 (iPhone; CPU iPhone OS 17_0 like Mac OS X) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/17.0 Mobile/15E148 Safari/604.1"
	default:
		return policy
	}
}
