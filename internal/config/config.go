package config

import (
	"os"
	"time"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Server ServerConfig `yaml:"server"`
	Search SearchConfig `yaml:"search"`
}

type ServerConfig struct {
	Address string `yaml:"address"`
}

type SearchConfig struct {
	Timeout  time.Duration `yaml:"timeout"`
	CacheTTL time.Duration `yaml:"cache_ttl"`
	Engines  []string      `yaml:"engines"`
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
	return cfg, nil
}

func withDefaults(cfg Config) Config {
	if cfg.Server.Address == "" {
		cfg.Server.Address = "127.0.0.1:8090"
	}
	if cfg.Search.Timeout <= 0 {
		cfg.Search.Timeout = 10 * time.Second
	}
	if cfg.Search.CacheTTL <= 0 {
		cfg.Search.CacheTTL = 5 * time.Minute
	}
	if len(cfg.Search.Engines) == 0 {
		cfg.Search.Engines = []string{"bing", "so360", "sogou", "zhwikipedia"}
	}
	return cfg
}
