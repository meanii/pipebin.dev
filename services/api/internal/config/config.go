package config

import "github.com/meanii/pipebin.dev/libs/config"

type Config struct {
	APP_PORT       string
	POSTGRESQL_DSN string
	LOGGER         string
}

func LoadConfig() *Config {
	config.LoadDotEnv(".env", "configs")
	return &Config{
		APP_PORT:       config.GetEnv("APP_PORT", "8001"),
		POSTGRESQL_DSN: config.MustGetEnv("POSTGRESQL_DSN"),
		LOGGER:         config.GetEnv("LOGGER", "development"),
	}
}
