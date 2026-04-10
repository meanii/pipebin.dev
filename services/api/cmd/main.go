package main

import (
	"context"
	"fmt"
	"log"
	"log/slog"
	"net/http"
	"time"

	"github.com/meanii/pipebin.dev/libs/logger"
	"github.com/meanii/pipebin.dev/services/api/handler"
	"github.com/meanii/pipebin.dev/services/api/internal/config"
	"github.com/meanii/pipebin.dev/services/api/internal/database"
	"github.com/meanii/pipebin.dev/services/api/internal/middleware"
	"github.com/meanii/pipebin.dev/services/api/internal/server"
	"github.com/meanii/pipebin.dev/services/api/internal/services"
	"github.com/meanii/pipebin.dev/services/api/repository"
)

func main() {
	cfg := config.LoadConfig()
	logger.Setup(cfg.LOGGER)

	pool, err := database.New(cfg.POSTGRESQL_DSN)
	if err != nil {
		log.Fatal(err)
	}
	defer pool.Close()

	pasteRepo := repository.NewPasteRespository(pool)
	pasteService := services.NewPastesService(pasteRepo)

	// Background goroutine: delete expired pastes every 10 minutes.
	go runExpiryCleanup(pasteRepo)

	h := server.NewRouter(server.Dependencies{
		PasteHandler:  *handler.NewPasteHandler(pasteService),
		HealthHandler: *handler.NewHealthHandler(pool),
	})

	mux := server.EnableGlobalMiddlewares(
		h,
		middleware.LoggerMiddleware,
		middleware.RateLimitMiddleware,
		middleware.RequestIDMiddleware,
	)

	slog.Info("api listening", slog.String("addr", fmt.Sprintf("http://0.0.0.0:%s", cfg.APP_PORT)))
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%s", cfg.APP_PORT), mux))
}

func runExpiryCleanup(repo repository.Repository) {
	ticker := time.NewTicker(10 * time.Minute)
	defer ticker.Stop()
	for range ticker.C {
		n, err := repo.DeleteExpired(context.Background())
		if err != nil {
			slog.Error("expiry cleanup failed", slog.String("error", err.Error()))
			continue
		}
		if n > 0 {
			slog.Info("expiry cleanup", slog.Int64("deleted", n))
		}
	}
}
