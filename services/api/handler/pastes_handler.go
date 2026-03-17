package handler

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/meanii/pipebin.dev/libs/hash"
	"github.com/meanii/pipebin.dev/libs/models"
	"github.com/meanii/pipebin.dev/services/api/internal/httpx"
	"github.com/meanii/pipebin.dev/services/api/internal/services"
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
		Title     string     `json:"title" validate:"required"`
		Content   string     `json:"content" validate:"required"`
		Langauge  string     `json:"language" validate:"required"`
		ExpiresAt *time.Time `json:"expires_at"`
	}
	var err error

	if err = json.NewDecoder(r.Body).Decode(&req); err != nil {
		httpx.EResponse(w, err.Error(), http.StatusBadRequest)
		return
	}

	err = validator.New(validator.WithRequiredStructEnabled()).Struct(&req)
	if err != nil {
		httpx.EResponse(w, err.Error(), http.StatusBadRequest)
		return
	}

	userIP := httpx.GetClientIP(r)
	userIPHash := hash.GetSHA256Hash(userIP)

	publicID, err := h.pasteService.CreatePaste(r.Context(), models.CreatePasteInput{
		Title:     req.Title,
		Content:   req.Content,
		Language:  req.Langauge,
		ExpiresAt: req.ExpiresAt,
		IPHash:    userIPHash,
		UserAgent: r.Header.Get("User-Agent"),
	})
	if err != nil {
		httpx.EResponse(w, err.Error(), http.StatusBadRequest)
		return
	}

	pipebinUrl := fmt.Sprintf("http://localhost:8002/p/%s", publicID)
	httpx.Response(w, map[string]interface{}{
		"url": pipebinUrl,
	}, http.StatusCreated)
}

func (h *PasteHandler) GetPasteByPublicID(w http.ResponseWriter, r *http.Request) {
	publicID := r.PathValue("public_id")
	paste, err := h.pasteService.GetPasteByPublicID(r.Context(), publicID)
	if err != nil {
		httpx.EResponse(w, err.Error(), http.StatusNotFound)
		return
	}
	httpx.Response(w, map[string]interface{}{
		"content": paste.Content,
	}, http.StatusOK)
}
