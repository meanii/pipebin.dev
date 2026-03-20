package server

import (
	"net/http"

	"github.com/meanii/pipebin.dev/services/api/handler"
)

type Dependencies struct {
	PasteHandler handler.PasteHandler
}

type Middleware func(http.Handler) http.Handler

func NewRouter(
	deps Dependencies,
) *http.ServeMux {
	mux := http.NewServeMux()

	// root endpoint (POST api.local.pipebin.dev)
	mux.HandleFunc("POST /", deps.PasteHandler.CreatePaste)
	mux.HandleFunc("GET /p/{public_id}", deps.PasteHandler.GetPasteByPublicID)

	return mux
}

func EnableGlobalMiddlewares(handler http.Handler, middlewares ...Middleware) http.Handler {
	for i := len(middlewares) - 1; i >= 0; i-- {
		handler = middlewares[i](handler)
	}
	return handler
}
