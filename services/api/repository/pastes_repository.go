package repository

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/meanii/pipebin.dev/libs/models"
	"gorm.io/gorm"
)

type PasteModel struct {
	ID        uuid.UUID  `gorm:"type:UUID;default:gen_random_uuid();column:ID"`
	PublicID  string     `gorm:"type:VARCHAR(50);not null;column:public_id"`
	Title     string     `gorm:"type:VARCHAR(50);not null;column:title"`
	Content   string     `gorm:"type:TEXT;column:content"`
	Language  string     `gorm:"type:VARCHAR(50);not null;column:language"`
	IPHash    string     `gorm:"type:INET;not null;column:ip_hash"`
	UserAgent string     `gorm:"type:VARCHAR(255);column:user_agent"`
	ExpiresAt *time.Time `gorm:"type:TIMESTAMPTZ;column:expires_at"`
	CreatedAt time.Time  `gorm:"autoCreateTime"`
}

type PastesRepository struct {
	db *gorm.DB
}

func NewPasteRespository(db *gorm.DB) *PastesRepository {
	return &PastesRepository{
		db: db,
	}
}

func (r *PastesRepository) Create(ctx context.Context, paste *models.Paste) error {
	m := &PasteModel{
		PublicID:  paste.PublicID,
		Title:     paste.Title,
		Content:   paste.Content,
		Language:  paste.Language,
		IPHash:    paste.IPHash,
		UserAgent: paste.UserAgent,
		ExpiresAt: paste.ExpiresAt,
	}
	err := r.db.WithContext(ctx).Create(m).Error
	return err
}

func (r *PastesRepository) GetByPublicID(ctx context.Context, publicID string) (*models.Paste, error) {
	var paste PasteModel
	if publicID == "" {
		return nil, errors.New("provide valid publicID")
	}
	err := r.db.WithContext(ctx).Where("PublicID = ?", publicID).First(&paste).Error
	return &models.Paste{
		ID:        paste.ID,
		Title:     paste.Title,
		Content:   paste.Content,
		Language:  paste.Language,
		IPHash:    paste.IPHash,
		UserAgent: paste.UserAgent,
		ExpiresAt: paste.ExpiresAt,
		CreatedAt: paste.CreatedAt,
	}, err
}
