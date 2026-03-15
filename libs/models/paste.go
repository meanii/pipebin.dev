package models

import (
	"time"

	"github.com/google/uuid"
)

type Paste struct {
	ID        uuid.UUID
	PublicID  string
	Title     string
	Content   string
	Size      int
	Language  string
	IPHash    string
	UserAgent string
	ExpiresAt *time.Time
	CreatedAt time.Time
}
