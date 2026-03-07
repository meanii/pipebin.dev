package logger

import (
	"os"

	"go.uber.org/zap"
)

var Log *zap.Logger

func Setup(env string) {
	var logger *zap.Logger
	var err error

	if env == "development" {
		logger, err = zap.NewDevelopment()
	} else {
		logger, err = zap.NewProduction()
	}

	if err != nil {
		panic(err)
	}

	Log = logger
	zap.ReplaceGlobals(logger)
}

func Sync() {
	if Log != nil {
		_ = Log.Sync()
	}
}

func Env() string {
	env := os.Getenv("ENV")

	if env == "" {
		env = "production"
	}

	return env
}
