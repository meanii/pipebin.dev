package handler

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/meanii/pipebin.dev/libs/models"
	"github.com/meanii/pipebin.dev/services/api/internal/httpx"
	"github.com/meanii/pipebin.dev/services/api/internal/services"
	"go.uber.org/zap"
)

type PasteHandler struct {
	pasteService services.PastesService
}

func NewPasteHandler(pasteService services.PastesService) *PasteHandler {
	return &PasteHandler{
		pasteService: pasteService,
	}
}

func (h *PasteHandler) CreatePaste(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Title     string     `json:"title"`
		Content   string     `json:"content"`
		Langauge  string     `json:"language"`
		ExpiresAt *time.Time `json:"expires_at"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	userIP := httpx.GetClientIP(r)
	zap.S().Infof("userIp: ", userIP)

	publicID, err := h.pasteService.CreatePaste(r.Context(), models.CreatePasteInput{
		Title:     req.Title,
		Content:   req.Content,
		Language:  req.Langauge,
		ExpiresAt: req.ExpiresAt,
		IPHash:    userIP,
		UserAgent: r.Header.Get("User-Agent"),
	})
	if err != nil {
		http.Error(w, "failed to create bin", http.StatusBadGateway)
		return
	}

	pipebinUrl := fmt.Sprintf("http://localhost:8002/p/%s", publicID)
	w.WriteHeader(http.StatusCreated)
	w.Write([]byte(pipebinUrl))
}
