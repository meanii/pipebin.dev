package server

import (
	"io/fs"
	"net/http"

	"github.com/meanii/pipebin.dev/services/frontend/handlers"
)

func NewRouter(handler *handlers.FrontendHandler, static fs.FS) http.Handler {
	mux := http.NewServeMux()

	return mux
}
