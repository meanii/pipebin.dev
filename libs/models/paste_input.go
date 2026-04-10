package models

import "time"

type CreatePasteInput struct {
	Title            string
	Content          string
	Language         string
	ExpiresAt        *time.Time
	IPHash           string
	UserAgent        string
	BurnAfterReading bool
}
