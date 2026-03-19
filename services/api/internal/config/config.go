package config

import (
	"sync"

	"github.com/meanii/pipebin.dev/libs/config"
)

type Config struct {
	APP_PORT                string
	POSTGRESQL_DSN          string
	MAX_PASTE_SIZE_IN_BYTES int
	MAX_NANO_ID_LENGTH      int
	FRONTEND_URL            string
	LOGGER                  string
}

var GlobalConfig *Config

func LoadConfig() *Config {
	once := sync.Once{}
	once.Do(func() {
		config.LoadDotEnv(".env", "configs")
		GlobalConfig = &Config{
			APP_PORT:                config.GetEnv("APP_PORT", "8001"),
			POSTGRESQL_DSN:          config.MustGetEnv("POSTGRESQL_DSN"),
			MAX_PASTE_SIZE_IN_BYTES: config.GetEnv("MAX_PASTE_SIZE_IN_BYTES", 10<<20), // 10 MB
			MAX_NANO_ID_LENGTH:      config.GetEnv("MAX_NANO_ID_LENGTH", 24),
			FRONTEND_URL:            config.GetEnv("FRONTEND_URL", "http://localhost:8001"),
			LOGGER:                  config.GetEnv("LOGGER", "development"),
		}
	})
	return GlobalConfig
}
