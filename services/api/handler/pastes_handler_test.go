package handler

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/meanii/pipebin.dev/libs/models"
	"github.com/meanii/pipebin.dev/services/api/internal/config"
	"github.com/meanii/pipebin.dev/services/api/internal/services"
)

// mockService implements services.Service for handler tests.
type mockService struct {
	createID  string
	createErr error
	getPaste  *models.Paste
	getErr    error
}

var _ services.Service = (*mockService)(nil)

func (m *mockService) CreatePaste(_ context.Context, _ models.CreatePasteInput) (string, error) {
	return m.createID, m.createErr
}

func (m *mockService) GetPasteByPublicID(_ context.Context, _ string) (*models.Paste, error) {
	return m.getPaste, m.getErr
}

func ensureConfig() {
	if config.GlobalConfig == nil {
		config.GlobalConfig = &config.Config{
			MAX_PASTE_SIZE_IN_BYTES: 10 << 20,
			MAX_NANO_ID_LENGTH:      24,
			FRONTEND_URL:            "http://localhost:8002",
		}
	}
}

func TestCreatePaste_Success(t *testing.T) {
	ensureConfig()
	h := NewPasteHandler(&mockService{createID: "abc123xyz"})

	body := `{"title":"test","content":"hello","language":"text"}`
	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	h.CreatePaste(rec, req)

	if rec.Code != http.StatusCreated {
		t.Errorf("status: got %d, want %d", rec.Code, http.StatusCreated)
	}
	var resp map[string]interface{}
	json.NewDecoder(rec.Body).Decode(&resp)
	data := resp["data"].(map[string]interface{})
	if !strings.Contains(data["url"].(string), "abc123xyz") {
		t.Errorf("url %q does not contain public_id", data["url"])
	}
}

func TestCreatePaste_CurlUserAgent_PlainText(t *testing.T) {
	ensureConfig()
	h := NewPasteHandler(&mockService{createID: "curltest123"})

	body := `{"title":"curl","content":"data","language":"text"}`
	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", "curl/8.1.2")
	rec := httptest.NewRecorder()

	h.CreatePaste(rec, req)

	if rec.Code != http.StatusCreated {
		t.Errorf("status: got %d, want %d", rec.Code, http.StatusCreated)
	}
	ct := rec.Header().Get("Content-Type")
	if !strings.HasPrefix(ct, "text/plain") {
		t.Errorf("Content-Type: got %q, want text/plain", ct)
	}
	if !strings.Contains(rec.Body.String(), "curltest123") {
		t.Errorf("body %q missing public_id", rec.Body.String())
	}
}

func TestCreatePaste_MissingFields(t *testing.T) {
	ensureConfig()
	h := NewPasteHandler(&mockService{})

	// content is required — title and language have defaults, but content does not
	body := `{"title":"t","language":"go"}` // missing content
	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	h.CreatePaste(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("status: got %d, want %d", rec.Code, http.StatusBadRequest)
	}
}

func TestCreatePaste_ServiceError(t *testing.T) {
	ensureConfig()
	h := NewPasteHandler(&mockService{createErr: errors.New("db down")})

	body := `{"title":"t","content":"c","language":"text"}`
	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	h.CreatePaste(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("status: got %d, want %d", rec.Code, http.StatusBadRequest)
	}
}

func TestGetPaste_Success(t *testing.T) {
	ensureConfig()
	now := time.Now()
	h := NewPasteHandler(&mockService{getPaste: &models.Paste{
		PublicID:  "testid",
		Title:     "hello",
		Content:   "world",
		Language:  "text",
		CreatedAt: now,
	}})

	req := httptest.NewRequest(http.MethodGet, "/p/testid", nil)
	req.SetPathValue("public_id", "testid")
	rec := httptest.NewRecorder()

	h.GetPasteByPublicID(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("status: got %d, want %d", rec.Code, http.StatusOK)
	}
}

func TestGetPaste_NotFound(t *testing.T) {
	ensureConfig()
	h := NewPasteHandler(&mockService{getErr: errors.New("paste not found")})

	req := httptest.NewRequest(http.MethodGet, "/p/notexist", nil)
	req.SetPathValue("public_id", "notexist")
	rec := httptest.NewRecorder()

	h.GetPasteByPublicID(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Errorf("status: got %d, want %d", rec.Code, http.StatusNotFound)
	}
}

func TestGetPaste_Expired(t *testing.T) {
	ensureConfig()
	past := time.Now().Add(-1 * time.Hour)
	h := NewPasteHandler(&mockService{getPaste: &models.Paste{
		PublicID:  "expiredid",
		Title:     "expired",
		Content:   "gone",
		Language:  "text",
		ExpiresAt: &past,
	}})

	req := httptest.NewRequest(http.MethodGet, "/p/expiredid", nil)
	req.SetPathValue("public_id", "expiredid")
	rec := httptest.NewRecorder()

	h.GetPasteByPublicID(rec, req)

	if rec.Code != http.StatusGone {
		t.Errorf("status: got %d, want %d", rec.Code, http.StatusGone)
	}
}
