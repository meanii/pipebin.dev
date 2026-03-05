package main

import (
	"github.com/meanii/pipebin.dev/libs/logger"
	"github.com/meanii/pipebin.dev/services/frontend/internal/config"
)

func main() {
	cfg := config.LoadConfig()
	logger.SetupLogger(cfg.LOGGER)
}
