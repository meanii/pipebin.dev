package server

import (
	"io/fs"
	"net/http"

	"github.com/meanii/pipebin.dev/services/frontend/handlers"
)

func NewRouter(handler *handlers.FrontendHandler, static fs.FS) http.Handler {
	mux := http.NewServeMux()

	mux.Handle("GET /static/", http.StripPrefix("/static/", http.FileServerFS(static)))

	mux.HandleFunc("GET /{$}", handler.Home)
	mux.HandleFunc("POST /", handler.CreatePaste)
	mux.HandleFunc("PUT /", handler.CreatePaste) // curl -T - URL
	mux.HandleFunc("GET /p/{id}", handler.Paste)
	mux.HandleFunc("GET /raw/{id}", handler.RawPaste)

	return mux
}
