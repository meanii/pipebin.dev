package main

import (
	"github.com/meanii/pipebin.dev/libs/logger"
	"github.com/meanii/pipebin.dev/services/api/internal/config"
	"go.uber.org/zap"
)

func main() {
	cfg := config.LoadConfig()
	logger.SetupLogger(cfg.LOGGER)

	zap.S().Info("config loaded.")
}
