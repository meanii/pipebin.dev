package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/meanii/pipebin.dev/libs/logger"
	"github.com/meanii/pipebin.dev/services/api/internal/config"
	"github.com/meanii/pipebin.dev/services/api/internal/database"
	"github.com/meanii/pipebin.dev/services/api/internal/server"
	"go.uber.org/zap"
)

func main() {
	cfg := config.LoadConfig()
	logger.Setup(cfg.LOGGER)
	defer logger.Sync()

	_, err := database.New(cfg.POSTGRESQL_DSN)
	if err != nil {
		log.Fatal(err)
	}

	mux := server.NewRouter(server.Dependencies{})

	zap.S().Infof("api.internal.pipebin.dev listening on http://0.0.0.0:%s", cfg.APP_PORT)
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%s", cfg.APP_PORT), mux))
}
