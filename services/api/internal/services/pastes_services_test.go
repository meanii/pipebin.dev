package services

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/meanii/pipebin.dev/libs/models"
	"github.com/meanii/pipebin.dev/services/api/internal/config"
	"github.com/meanii/pipebin.dev/services/api/repository"
)

// mockRepo is an in-memory implementation of repository.Repository for testing.
type mockRepo struct {
	pastes    map[string]*models.Paste
	createErr error
	getErr    error
}

var _ repository.Repository = (*mockRepo)(nil)

func newMockRepo() *mockRepo {
	return &mockRepo{pastes: make(map[string]*models.Paste)}
}

func (m *mockRepo) Create(_ context.Context, paste *models.Paste) error {
	if m.createErr != nil {
		return m.createErr
	}
	paste.ID = uuid.New()
	paste.CreatedAt = time.Now()
	m.pastes[paste.PublicID] = paste
	return nil
}

func (m *mockRepo) GetByPublicID(_ context.Context, publicID string) (*models.Paste, error) {
	if m.getErr != nil {
		return nil, m.getErr
	}
	p, ok := m.pastes[publicID]
	if !ok {
		return nil, errors.New("paste not found")
	}
	return p, nil
}

func (m *mockRepo) DeleteByPublicID(_ context.Context, publicID string) error {
	delete(m.pastes, publicID)
	return nil
}

func (m *mockRepo) DeleteExpired(_ context.Context) (int64, error) {
	return 0, nil
}

// ensureConfig sets up GlobalConfig with test defaults if not already set.
func ensureConfig() {
	if config.GlobalConfig == nil {
		config.GlobalConfig = &config.Config{
			MAX_PASTE_SIZE_IN_BYTES: 10 << 20,
			MAX_NANO_ID_LENGTH:      24,
			FRONTEND_URL:            "http://localhost:8002",
		}
	}
}

func TestCreatePaste_HappyPath(t *testing.T) {
	ensureConfig()
	svc := NewPastesService(newMockRepo())

	publicID, err := svc.CreatePaste(context.Background(), models.CreatePasteInput{
		Title:    "test",
		Content:  "hello world",
		Language: "text",
		IPHash:   "abc123",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(publicID) != config.GlobalConfig.MAX_NANO_ID_LENGTH {
		t.Errorf("publicID length: got %d, want %d", len(publicID), config.GlobalConfig.MAX_NANO_ID_LENGTH)
	}
}

func TestCreatePaste_ContentTooLarge(t *testing.T) {
	ensureConfig()
	svc := NewPastesService(newMockRepo())

	big := make([]byte, config.GlobalConfig.MAX_PASTE_SIZE_IN_BYTES+1)
	_, err := svc.CreatePaste(context.Background(), models.CreatePasteInput{
		Title:    "big",
		Content:  string(big),
		Language: "text",
		IPHash:   "abc123",
	})
	if err == nil {
		t.Fatal("expected error for oversized content, got nil")
	}
}

func TestCreatePaste_RepoError(t *testing.T) {
	ensureConfig()
	repo := newMockRepo()
	repo.createErr = errors.New("db error")
	svc := NewPastesService(repo)

	_, err := svc.CreatePaste(context.Background(), models.CreatePasteInput{
		Title:    "test",
		Content:  "content",
		Language: "text",
		IPHash:   "abc",
	})
	if err == nil {
		t.Fatal("expected error from repo, got nil")
	}
}

func TestGetPasteByPublicID_InvalidLength(t *testing.T) {
	ensureConfig()
	svc := NewPastesService(newMockRepo())

	_, err := svc.GetPasteByPublicID(context.Background(), "short")
	if err == nil {
		t.Fatal("expected error for invalid publicID length, got nil")
	}
}

func TestGetPasteByPublicID_NotFound(t *testing.T) {
	ensureConfig()
	svc := NewPastesService(newMockRepo())

	id := string(make([]byte, config.GlobalConfig.MAX_NANO_ID_LENGTH))
	_, err := svc.GetPasteByPublicID(context.Background(), id)
	if err == nil {
		t.Fatal("expected not-found error, got nil")
	}
}

func TestGetPasteByPublicID_HappyPath(t *testing.T) {
	ensureConfig()
	repo := newMockRepo()
	svc := NewPastesService(repo)

	// Create via service so it lands in the mock.
	publicID, err := svc.CreatePaste(context.Background(), models.CreatePasteInput{
		Title:    "hello",
		Content:  "world",
		Language: "text",
		IPHash:   "abc",
	})
	if err != nil {
		t.Fatalf("create: %v", err)
	}

	paste, err := svc.GetPasteByPublicID(context.Background(), publicID)
	if err != nil {
		t.Fatalf("get: %v", err)
	}
	if paste.Title != "hello" {
		t.Errorf("title: got %q, want %q", paste.Title, "hello")
	}
}
