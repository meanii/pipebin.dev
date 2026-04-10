package services

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	gonanoid "github.com/matoous/go-nanoid/v2"
	"github.com/meanii/pipebin.dev/libs/models"
	"github.com/meanii/pipebin.dev/services/api/internal/config"
	"github.com/meanii/pipebin.dev/services/api/repository"
)

// Compile-time check: PastesService must implement Service.
var _ Service = (*PastesService)(nil)

type PastesService struct {
	pastesRepository repository.Repository
}

func NewPastesService(repo repository.Repository) *PastesService {
	return &PastesService{pastesRepository: repo}
}

func (s *PastesService) CreatePaste(ctx context.Context, input models.CreatePasteInput) (string, error) {
	size := len(input.Content)
	if size >= config.GlobalConfig.MAX_PASTE_SIZE_IN_BYTES {
		return "", fmt.Errorf("content size must be less than %d bytes", config.GlobalConfig.MAX_PASTE_SIZE_IN_BYTES)
	}

	publicID, err := gonanoid.New(config.GlobalConfig.MAX_NANO_ID_LENGTH)
	if err != nil {
		return "", err
	}

	paste := &models.Paste{
		PublicID:         publicID,
		Title:            input.Title,
		Content:          input.Content,
		Size:             size,
		Language:         input.Language,
		IPHash:           input.IPHash,
		UserAgent:        input.UserAgent,
		ExpiresAt:        input.ExpiresAt,
		BurnAfterReading: input.BurnAfterReading,
	}

	if err = s.pastesRepository.Create(ctx, paste); err != nil {
		return "", err
	}
	return publicID, nil
}

func (s *PastesService) GetPasteByPublicID(ctx context.Context, publicID string) (*models.Paste, error) {
	if len(publicID) != config.GlobalConfig.MAX_NANO_ID_LENGTH {
		return nil, errors.New("invalid publicID")
	}

	paste, err := s.pastesRepository.GetByPublicID(ctx, publicID)
	if err != nil {
		return nil, err
	}

	// Burn after reading: delete the paste immediately so it can only be viewed once.
	if paste.BurnAfterReading {
		if delErr := s.pastesRepository.DeleteByPublicID(ctx, publicID); delErr != nil {
			slog.WarnContext(ctx, "burn after reading: failed to delete paste",
				slog.String("public_id", publicID),
				slog.String("error", delErr.Error()),
			)
		} else {
			slog.InfoContext(ctx, "burn after reading: paste deleted after first view",
				slog.String("public_id", publicID),
			)
		}
	}

	return paste, nil
}
