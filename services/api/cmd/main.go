package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/meanii/pipebin.dev/libs/logger"
	"github.com/meanii/pipebin.dev/services/api/handler"
	"github.com/meanii/pipebin.dev/services/api/internal/config"
	"github.com/meanii/pipebin.dev/services/api/internal/database"
	"github.com/meanii/pipebin.dev/services/api/internal/server"
	"github.com/meanii/pipebin.dev/services/api/internal/services"
	"github.com/meanii/pipebin.dev/services/api/repository"
	"go.uber.org/zap"
)

func main() {
	cfg := config.LoadConfig()
	logger.Setup(cfg.LOGGER)
	defer logger.Sync()

	database, err := database.New(cfg.POSTGRESQL_DSN)
	if err != nil {
		log.Fatal(err)
	}

	mux := server.NewRouter(server.Dependencies{
		PasteHandler: *handler.NewPasteHandler(*services.NewPastesService(
			*repository.NewPasteRespository(database),
		)),
	})

	zap.S().Infof("api.internal.pipebin.dev listening on http://0.0.0.0:%s", cfg.APP_PORT)
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%s", cfg.APP_PORT), mux))
}
