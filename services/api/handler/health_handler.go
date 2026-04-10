package handler

import (
	"log/slog"
	"net/http"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/meanii/pipebin.dev/services/api/internal/httpx"
)

type HealthHandler struct {
	db *pgxpool.Pool
}

func NewHealthHandler(db *pgxpool.Pool) *HealthHandler {
	return &HealthHandler{db: db}
}

func (h *HealthHandler) Health(w http.ResponseWriter, r *http.Request) {
	if err := h.db.Ping(r.Context()); err != nil {
		slog.ErrorContext(r.Context(), "healthz: db ping failed", slog.String("error", err.Error()))
		httpx.EResponse(w, "database unavailable", http.StatusServiceUnavailable)
		return
	}
	httpx.Response(w, map[string]interface{}{"status": "ok", "db": "ok"}, http.StatusOK)
}
