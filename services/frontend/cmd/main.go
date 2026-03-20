package main

import (
	"fmt"
	"io/fs"
	"net/http"

	"github.com/meanii/pipebin.dev/libs/logger"
	"github.com/meanii/pipebin.dev/services/frontend"
	"github.com/meanii/pipebin.dev/services/frontend/handlers"
	"github.com/meanii/pipebin.dev/services/frontend/internal/config"
	"github.com/meanii/pipebin.dev/services/frontend/internal/server"
	"go.uber.org/zap"
)

func main() {
	cfg := config.LoadConfig()
	logger.Setup(cfg.LOGGER)
	defer logger.Sync()

	templateFS, _ := fs.Sub(frontend.TemplateFS, "templates")
	staticFS, _ := fs.Sub(frontend.StaticFS, "static")

	handler := handlers.NewFrontendHandler(templateFS, cfg.API_BASE_URL)
	router := server.NewRouter(handler, staticFS)

	zap.L().Sugar().Infof("pipebin.dev running on http://0.0.0.0:%s", cfg.FA_PORT)
	http.ListenAndServe(fmt.Sprintf(":%s", cfg.FA_PORT), router)
}
