package config

import (
	"fmt"
	"os"
	"time"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Server ServerConfig `yaml:"server"`
	Search SearchConfig `yaml:"search"`
}

type ServerConfig struct {
	Address       string          `yaml:"address"`
	RateLimit     RateLimitConfig `yaml:"rate_limit"`
	LogStructured bool            `yaml:"log_structured"`
}

type SearchConfig struct {
	Timeout         time.Duration `yaml:"timeout"`
	RequestTimeout  time.Duration `yaml:"request_timeout"`
	CacheTTL        time.Duration `yaml:"cache_ttl"`
	Engines         []string      `yaml:"engines"`
	EnabledEngines  []string      `yaml:"enabled_engines"`
	ProxyURL        string        `yaml:"proxy_url"`
	UserAgentPolicy string        `yaml:"user_agent_policy"`
	Filters         FilterConfig  `yaml:"filters"`
}

type RateLimitConfig struct {
	Enabled           bool `yaml:"enabled"`
	RequestsPerMinute int  `yaml:"requests_per_minute"`
}

type FilterConfig struct {
	Allowlist []string `yaml:"allowlist"`
	Blocklist []string `yaml:"blocklist"`
}

func Load(path string) (Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return Config{}, err
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return Config{}, err
	}
	cfg = withDefaults(cfg)
	if err := Validate(cfg); err != nil {
		return Config{}, err
	}
	return cfg, nil
}

func Validate(cfg Config) error {
	if cfg.Server.Address == "" {
		return fmt.Errorf("server address is required")
	}
	if cfg.Search.Timeout <= 0 {
		return fmt.Errorf("search timeout must be positive")
	}
	if cfg.Search.CacheTTL <= 0 {
		return fmt.Errorf("cache ttl must be positive")
	}
	if len(cfg.Search.Engines) == 0 {
		return fmt.Errorf("at least one search engine is required")
	}
	if cfg.Server.RateLimit.Enabled && cfg.Server.RateLimit.RequestsPerMinute <= 0 {
		return fmt.Errorf("rate_limit.requests_per_minute must be positive when rate limiter is enabled")
	}
	return nil
}

func withDefaults(cfg Config) Config {
	if cfg.Server.Address == "" {
		cfg.Server.Address = "127.0.0.1:8090"
	}
	if cfg.Search.Timeout <= 0 && cfg.Search.RequestTimeout > 0 {
		cfg.Search.Timeout = cfg.Search.RequestTimeout
	}
	if cfg.Search.RequestTimeout <= 0 && cfg.Search.Timeout > 0 {
		cfg.Search.RequestTimeout = cfg.Search.Timeout
	}
	if cfg.Search.Timeout <= 0 {
		cfg.Search.Timeout = 10 * time.Second
		cfg.Search.RequestTimeout = cfg.Search.Timeout
	}
	if cfg.Search.CacheTTL <= 0 {
		cfg.Search.CacheTTL = 5 * time.Minute
	}
	if len(cfg.Search.Engines) == 0 && len(cfg.Search.EnabledEngines) > 0 {
		cfg.Search.Engines = cfg.Search.EnabledEngines
	}
	if len(cfg.Search.EnabledEngines) == 0 && len(cfg.Search.Engines) > 0 {
		cfg.Search.EnabledEngines = cfg.Search.Engines
	}
	if len(cfg.Search.Engines) == 0 {
		cfg.Search.Engines = []string{"bing", "so360", "sogou", "zhwikipedia"}
		cfg.Search.EnabledEngines = cfg.Search.Engines
	}
	if cfg.Search.UserAgentPolicy == "" {
		cfg.Search.UserAgentPolicy = "desktop"
	}
	if cfg.Server.RateLimit.Enabled && cfg.Server.RateLimit.RequestsPerMinute == 0 {
		cfg.Server.RateLimit.RequestsPerMinute = 60
	}
	return cfg
}
