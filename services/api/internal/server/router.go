package server

import "net/http"

type Dependencies struct{}

func NewRouter(deps Dependencies) *http.ServeMux {
	mux := http.NewServeMux()

	// root endpoint (POST api.pipebin.dev)
	mux.HandleFunc("POST /", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("X-Server", "api.internal.pipebin.dev")
		w.WriteHeader(200)
		w.Write([]byte("Created."))
	})

	return mux
}
