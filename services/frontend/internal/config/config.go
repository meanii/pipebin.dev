package config

import "github.com/meanii/pipebin.dev/libs/config"

type Config struct {
	FA_PORT string
	LOGGER  string
}

func LoadConfig() *Config {
	return &Config{
		FA_PORT: config.GetEnv("FA_PORT", "8002"),
		LOGGER:  config.GetEnv("LOGGER", "development"),
	}
}
