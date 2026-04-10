package server

import (
	"net/http"

	"github.com/meanii/pipebin.dev/services/api/handler"
)

type Dependencies struct {
	PasteHandler  handler.PasteHandler
	HealthHandler handler.HealthHandler
}

type Middleware func(http.Handler) http.Handler

func NewRouter(deps Dependencies) *http.ServeMux {
	mux := http.NewServeMux()

	mux.HandleFunc("GET /healthz", deps.HealthHandler.Health)
	mux.HandleFunc("POST /", deps.PasteHandler.CreatePaste)
	mux.HandleFunc("PUT /", deps.PasteHandler.CreatePaste) // curl -T - URL
	mux.HandleFunc("GET /p/{public_id}", deps.PasteHandler.GetPasteByPublicID)

	return mux
}

func EnableGlobalMiddlewares(handler http.Handler, middlewares ...Middleware) http.Handler {
	for i := len(middlewares) - 1; i >= 0; i-- {
		handler = middlewares[i](handler)
	}
	return handler
}
