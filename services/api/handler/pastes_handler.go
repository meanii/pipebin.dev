package handler

import (
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/alecthomas/chroma/v2/lexers"
	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/meanii/pipebin.dev/libs/hash"
	"github.com/meanii/pipebin.dev/libs/models"
	"github.com/meanii/pipebin.dev/services/api/internal/config"
	"github.com/meanii/pipebin.dev/services/api/internal/httpx"
	"github.com/meanii/pipebin.dev/services/api/internal/services"
)

type PasteHandler struct {
	pasteService services.Service
}

func NewPasteHandler(pasteService services.Service) *PasteHandler {
	return &PasteHandler{pasteService: pasteService}
}

type createPasteRequest struct {
	Title            string `json:"title"`
	Content          string `json:"content"`
	Language         string `json:"language"`
	ExpiresAt        string `json:"expires_at"`
	BurnAfterReading bool   `json:"burn"` // true = delete on first GET
}

func (req createPasteRequest) Validate() error {
	return validation.ValidateStruct(&req,
		validation.Field(&req.Title, validation.Required, validation.Length(1, 255)),
		validation.Field(&req.Content, validation.Required),
		validation.Field(&req.Language, validation.Required, validation.Length(1, 50)),
	)
}

// detectLanguage uses Chroma's analyser to guess the language from content.
// Returns "" when confidence is too low (plain text / unknown).
func detectLanguage(content string) string {
	lexer := lexers.Analyse(content)
	if lexer == nil {
		return ""
	}
	cfg := lexer.Config()
	if cfg == nil {
		return ""
	}
	name := strings.ToLower(cfg.Name)
	// Chroma returns "plaintext" / "text only" when nothing matched — treat as unknown.
	if name == "plaintext" || name == "text only" || name == "text" {
		return ""
	}
	// Prefer the first alias (shorter, lowercase) if available.
	if len(cfg.Aliases) > 0 {
		return cfg.Aliases[0]
	}
	return name
}

// parseBool returns true for "1", "true", "yes" (case-insensitive).
func parseBool(s string) bool {
	s = strings.ToLower(strings.TrimSpace(s))
	return s == "1" || s == "true" || s == "yes"
}

// parseCreateBody reads the request body once and returns a createPasteRequest
// regardless of how the content was sent:
//
//   - application/json           — standard JSON body
//   - application/x-www-form-urlencoded — HTML form; falls back to raw mode if
//     the "content" field is absent (e.g. curl -d @-)
//   - anything else / no Content-Type — raw body mode (curl -T -)
//     metadata via query params: ?t=title  ?lang=go  ?e=24h  ?once=1
func parseCreateBody(r *http.Request) (createPasteRequest, error) {
	ct := r.Header.Get("Content-Type")

	body, err := io.ReadAll(r.Body)
	if err != nil {
		return createPasteRequest{}, err
	}

	if strings.HasPrefix(ct, "application/json") {
		var req createPasteRequest
		if err = json.Unmarshal(body, &req); err != nil {
			return createPasteRequest{}, err
		}
		applyDefaults(&req, false)
		return req, nil
	}

	if strings.HasPrefix(ct, "application/x-www-form-urlencoded") {
		vals, err := url.ParseQuery(string(body))
		if err != nil {
			return createPasteRequest{}, err
		}
		if content := vals.Get("content"); content != "" {
			// Genuine HTML form submission.
			req := createPasteRequest{
				Title:            vals.Get("title"),
				Content:          content,
				Language:         vals.Get("language"),
				ExpiresAt:        vals.Get("expires_at"),
				BurnAfterReading: parseBool(vals.Get("burn")),
			}
			applyDefaults(&req, false)
			return req, nil
		}
		// No "content" field — curl -d @- sends this Content-Type with a raw body.
	}

	// Raw body mode: cat file | curl -sT - https://pipebin.dev/
	q := r.URL.Query()
	expires := q.Get("expires")
	if expires == "" {
		expires = q.Get("e")
	}
	burn := parseBool(q.Get("once")) || parseBool(q.Get("burn"))

	req := createPasteRequest{
		Title:            q.Get("t"),
		Content:          string(body),
		Language:         q.Get("lang"),
		ExpiresAt:        expires,
		BurnAfterReading: burn,
	}
	// Auto-detect language when not provided in raw mode.
	applyDefaults(&req, true)
	return req, nil
}

// applyDefaults fills in Title and Language when absent.
// autoDetect enables Chroma language detection (only for raw pipe mode).
func applyDefaults(req *createPasteRequest, autoDetect bool) {
	if req.Title == "" {
		req.Title = "paste"
	}
	if req.Language == "" {
		if autoDetect {
			if detected := detectLanguage(req.Content); detected != "" {
				req.Language = detected
				return
			}
		}
		req.Language = "text"
	}
}

func (h *PasteHandler) CreatePaste(w http.ResponseWriter, r *http.Request) {
	req, err := parseCreateBody(r)
	if err != nil {
		slog.WarnContext(r.Context(), "create paste: failed to read body", slog.String("error", err.Error()))
		httpx.EResponse(w, "failed to read request body", http.StatusBadRequest)
		return
	}

	if err := req.Validate(); err != nil {
		slog.WarnContext(r.Context(), "create paste: validation failed", slog.String("error", err.Error()))
		httpx.EResponse(w, err.Error(), http.StatusBadRequest)
		return
	}

	userIPHash := hash.GetSHA256Hash(httpx.GetClientIP(r))

	input := models.CreatePasteInput{
		Title:            req.Title,
		Content:          req.Content,
		Language:         req.Language,
		IPHash:           userIPHash,
		UserAgent:        r.Header.Get("User-Agent"),
		BurnAfterReading: req.BurnAfterReading,
	}

	if req.ExpiresAt != "" {
		duration, err := time.ParseDuration(req.ExpiresAt)
		if err != nil {
			slog.WarnContext(r.Context(), "create paste: invalid expires_at", slog.String("value", req.ExpiresAt))
			httpx.EResponse(w, "invalid expires_at: "+err.Error(), http.StatusBadRequest)
			return
		}
		t := time.Now().Add(duration)
		input.ExpiresAt = &t
	}

	publicID, err := h.pasteService.CreatePaste(r.Context(), input)
	if err != nil {
		slog.ErrorContext(r.Context(), "create paste: service error", slog.String("error", err.Error()))
		httpx.EResponse(w, err.Error(), http.StatusBadRequest)
		return
	}

	slog.InfoContext(r.Context(), "paste created",
		slog.String("public_id", publicID),
		slog.String("language", req.Language),
		slog.Int("size", len(req.Content)),
		slog.Bool("burn", req.BurnAfterReading),
	)

	pipebinURL := fmt.Sprintf("%s/p/%s", config.GlobalConfig.FRONTEND_URL, publicID)

	// curl clients get plain-text output so piping works out of the box.
	if strings.HasPrefix(r.UserAgent(), "curl/") {
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.WriteHeader(http.StatusCreated)
		fmt.Fprintf(w, "%s\n", pipebinURL)
		if req.BurnAfterReading {
			fmt.Fprintf(w, "⚠  burn after reading — this paste will be deleted on first view\n")
		}
		return
	}

	httpx.Response(w, map[string]interface{}{
		"url":  pipebinURL,
		"burn": req.BurnAfterReading,
	}, http.StatusCreated)
}

func (h *PasteHandler) GetPasteByPublicID(w http.ResponseWriter, r *http.Request) {
	publicID := r.PathValue("public_id")

	paste, err := h.pasteService.GetPasteByPublicID(r.Context(), publicID)
	if err != nil {
		slog.WarnContext(r.Context(), "get paste: not found",
			slog.String("public_id", publicID),
			slog.String("error", err.Error()),
		)
		httpx.EResponse(w, err.Error(), http.StatusNotFound)
		return
	}

	if paste.ExpiresAt != nil && time.Now().After(*paste.ExpiresAt) {
		slog.InfoContext(r.Context(), "get paste: expired", slog.String("public_id", publicID))
		httpx.EResponse(w, "bin has been expired", http.StatusGone)
		return
	}

	slog.DebugContext(r.Context(), "get paste: ok",
		slog.String("public_id", publicID),
		slog.Bool("burn", paste.BurnAfterReading),
	)

	httpx.Response(w, map[string]interface{}{
		"id":                 paste.ID,
		"public_id":          paste.PublicID,
		"title":              paste.Title,
		"content":            paste.Content,
		"language":           paste.Language,
		"created_at":         paste.CreatedAt,
		"expires_at":         paste.ExpiresAt,
		"burn_after_reading": paste.BurnAfterReading,
	}, http.StatusOK)
}
