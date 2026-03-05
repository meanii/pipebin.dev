package config

import "github.com/meanii/pipebin.dev/libs/config"

type Config struct {
	APP_PORT string
	LOGGER   string
}

func LoadConfig() *Config {
	return &Config{
		APP_PORT: config.GetEnv("APP_PORT", "8001"),
		LOGGER:   config.GetEnv("LOGGER", "development"),
	}
}
