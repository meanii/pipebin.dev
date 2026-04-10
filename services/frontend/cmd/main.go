package main

import (
	"fmt"
	"io/fs"
	"log"
	"log/slog"
	"net/http"

	"github.com/meanii/pipebin.dev/libs/logger"
	"github.com/meanii/pipebin.dev/services/frontend"
	"github.com/meanii/pipebin.dev/services/frontend/handlers"
	"github.com/meanii/pipebin.dev/services/frontend/internal/config"
	"github.com/meanii/pipebin.dev/services/frontend/internal/server"
)

func main() {
	cfg := config.LoadConfig()
	logger.Setup(cfg.LOGGER)

	templateFS, _ := fs.Sub(frontend.TemplateFS, "templates")
	staticFS, _ := fs.Sub(frontend.StaticFS, "static")

	handler := handlers.NewFrontendHandler(templateFS, cfg.API_BASE_URL)
	router := server.NewRouter(handler, staticFS)

	slog.Info("frontend listening", slog.String("addr", fmt.Sprintf("http://0.0.0.0:%s", cfg.FA_PORT)))
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%s", cfg.FA_PORT), router))
}
