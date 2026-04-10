package repository

import (
	"context"

	"github.com/meanii/pipebin.dev/libs/models"
)

// Repository is the contract the service layer depends on.
// Keeping it here (next to the real implementation) avoids import cycles
// while still making the service testable via mocks.
type Repository interface {
	Create(ctx context.Context, paste *models.Paste) error
	GetByPublicID(ctx context.Context, publicID string) (*models.Paste, error)
	DeleteByPublicID(ctx context.Context, publicID string) error
	DeleteExpired(ctx context.Context) (int64, error)
}
