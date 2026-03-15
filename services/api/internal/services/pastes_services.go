package services

import (
	"context"
	"errors"
	"fmt"

	gonanoid "github.com/matoous/go-nanoid/v2"
	"github.com/meanii/pipebin.dev/libs/models"
	"github.com/meanii/pipebin.dev/services/api/internal/config"
	"github.com/meanii/pipebin.dev/services/api/repository"
)

// keeping 24 length of nano ID
const MAX_NANO_ID_LEN = 24

type PastesService struct {
	pastesRepository repository.PastesRepository
}

func NewPastesService(pastesRepo repository.PastesRepository) *PastesService {
	return &PastesService{
		pastesRepository: pastesRepo,
	}
}

func (s *PastesService) CreatePaste(ctx context.Context, input models.CreatePasteInput) (string, error) {

	size := len(input.Content)
	if size >= config.GlobalConfig.MAX_PASTE_SIZE_IN_BYTES {
		return "", fmt.Errorf("content size must be less than %d bytes.", config.GlobalConfig.MAX_PASTE_SIZE_IN_BYTES)
	}

	publicId, err := gonanoid.New(MAX_NANO_ID_LEN)
	if err != nil {
		return "", err
	}

	paste := &models.Paste{
		PublicID:  publicId,
		Title:     input.Title,
		Content:   input.Content,
		Size:      size,
		Language:  input.Language,
		IPHash:    input.IPHash,
		UserAgent: input.UserAgent,
	}
	if input.ExpiresAt != nil {
		paste.ExpiresAt = input.ExpiresAt
	}

	err = s.pastesRepository.Create(ctx, paste)
	return publicId, err
}

func (s *PastesService) GetPasteByPublicID(ctx context.Context, publicID string) (*models.Paste, error) {
	if len(publicID) != MAX_NANO_ID_LEN {
		return nil, errors.New("invalid publicID")
	}
	return s.pastesRepository.GetByPublicID(ctx, publicID)
}
