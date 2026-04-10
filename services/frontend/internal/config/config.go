package config

import "github.com/meanii/pipebin.dev/libs/config"

type Config struct {
	FA_PORT      string
	API_BASE_URL string
	LOGGER       string
}

func LoadConfig() *Config {
	return &Config{
		FA_PORT:      config.GetEnv("FA_PORT", "8002"),
		API_BASE_URL: config.GetEnv("API_BASE_URL", "http://localhost:8001"),
		LOGGER:       config.GetEnv("LOGGER", "development"),
	}
}
