package services

import (
	"context"

	"github.com/meanii/pipebin.dev/libs/models"
)

// Service is the contract the handler layer depends on.
type Service interface {
	CreatePaste(ctx context.Context, input models.CreatePasteInput) (string, error)
	GetPasteByPublicID(ctx context.Context, publicID string) (*models.Paste, error)
}
