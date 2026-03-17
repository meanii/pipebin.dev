package server

import (
	"net/http"

	"github.com/meanii/pipebin.dev/services/api/handler"
)

type Dependencies struct {
	PasteHandler handler.PasteHandler
}

func NewRouter(deps Dependencies) *http.ServeMux {
	mux := http.NewServeMux()

	// root endpoint (POST api.local.pipebin.dev)
	mux.HandleFunc("POST /", deps.PasteHandler.CreatePaste)
	mux.HandleFunc("GET /p/{public_id}", deps.PasteHandler.GetPasteByPublicID)

	return mux
}
