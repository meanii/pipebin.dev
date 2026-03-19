package httpx

import (
	"encoding/json"
	"net/http"
)

func Response(w http.ResponseWriter, data map[string]interface{}, status int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	resp := map[string]interface{}{
		"data":   data,
		"status": http.StatusText(status),
	}
	json.NewEncoder(w).Encode(resp)
}

func EResponse(w http.ResponseWriter, msg string, status int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	resp := map[string]interface{}{
		"error":  msg,
		"status": http.StatusText(status),
	}
	json.NewEncoder(w).Encode(resp)
}
