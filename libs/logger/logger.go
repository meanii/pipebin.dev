package logger

import (
	"os"

	"go.uber.org/zap"
)

func SetupLogger(env string) {
	logger := zap.Must(zap.NewProduction())
	if os.Getenv("env") == "development" {
		logger = zap.Must(zap.NewDevelopment())
	}
	zap.ReplaceGlobals(logger)
}
