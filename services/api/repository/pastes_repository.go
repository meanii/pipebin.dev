package repository

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/meanii/pipebin.dev/libs/models"
)

type PastesRepository struct {
	db *pgxpool.Pool
}

func NewPasteRespository(db *pgxpool.Pool) *PastesRepository {
	return &PastesRepository{db: db}
}

// Compile-time check: PastesRepository must implement Repository.
var _ Repository = (*PastesRepository)(nil)

func (r *PastesRepository) Create(ctx context.Context, paste *models.Paste) error {
	const q = `
		INSERT INTO pastes
			(public_id, title, content, size, language, ip_hash, user_agent, expires_at, burn_after_reading)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		RETURNING id, created_at`

	return r.db.QueryRow(ctx, q,
		paste.PublicID,
		paste.Title,
		paste.Content,
		paste.Size,
		paste.Language,
		paste.IPHash,
		paste.UserAgent,
		paste.ExpiresAt,
		paste.BurnAfterReading,
	).Scan(&paste.ID, &paste.CreatedAt)
}

func (r *PastesRepository) GetByPublicID(ctx context.Context, publicID string) (*models.Paste, error) {
	if publicID == "" {
		return nil, errors.New("provide valid publicID")
	}

	const q = `
		SELECT id, public_id, title, content, size, language, ip_hash, user_agent,
		       expires_at, burn_after_reading, created_at
		FROM pastes
		WHERE public_id = $1
		LIMIT 1`

	var p models.Paste
	err := r.db.QueryRow(ctx, q, publicID).Scan(
		&p.ID,
		&p.PublicID,
		&p.Title,
		&p.Content,
		&p.Size,
		&p.Language,
		&p.IPHash,
		&p.UserAgent,
		&p.ExpiresAt,
		&p.BurnAfterReading,
		&p.CreatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, errors.New("paste not found")
	}
	if err != nil {
		return nil, err
	}
	return &p, nil
}

func (r *PastesRepository) DeleteByPublicID(ctx context.Context, publicID string) error {
	const q = `DELETE FROM pastes WHERE public_id = $1`
	_, err := r.db.Exec(ctx, q, publicID)
	return err
}

func (r *PastesRepository) DeleteExpired(ctx context.Context) (int64, error) {
	const q = `DELETE FROM pastes WHERE expires_at IS NOT NULL AND expires_at < NOW()`
	tag, err := r.db.Exec(ctx, q)
	if err != nil {
		return 0, err
	}
	return tag.RowsAffected(), nil
}
