package config

import (
	"os"
	"slices"

	toml "github.com/pelletier/go-toml/v2"
)

type Config struct {
	Cache  CacheConfig
	Master MasterConfig
}

type CacheConfig struct {
	Provider string
	DbPath   string
}

type MasterConfig struct {
	Protocol       string
	Addr           string
	Send_Timeout   int
	Retry_Interval int
}

func NewDefaultConfig() *Config {
	return &Config{
		Cache: CacheConfig{
			Provider: "in_memory",
			DbPath:   "",
		},
		Master: MasterConfig{
			Send_Timeout:   1,
			Retry_Interval: 2,
		},
	}
}

func Load(path string) (*Config, error) {
	contents, err := loadConfig(path)
	if err != nil {
		return NewDefaultConfig(), err
	}
	cfg, err := parseConfig(contents)
	if err != nil {
		return nil, err
	}
	return validateConfig(cfg)
}

func loadConfig(path string) ([]byte, error) {
	return os.ReadFile(path)
}

func parseConfig(fileContents []byte) (*Config, error) {
	var cfg Config
	err := toml.Unmarshal(fileContents, &cfg)
	return &cfg, err
}

func validateConfig(cfg *Config) (*Config, error) {
	defaultCfg := NewDefaultConfig()
	if !slices.Contains([]string{"sqlite", "in_memory"}, cfg.Cache.Provider) {
		cfg.Cache.Provider = defaultCfg.Cache.Provider
	}

	return cfg, nil
}
