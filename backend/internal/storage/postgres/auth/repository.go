package auth

import (
	"context"
	"time"

	"github.com/google/uuid"

	"github.com/34Minnesota/avito-antiscam-monorepo/backend/internal/domain/models"
	postgrespool "github.com/34Minnesota/avito-antiscam-monorepo/backend/internal/storage/postgres/pool"
)

type AuthRepository struct {
	pool *postgrespool.Pool
}

func NewRepository(pool *postgrespool.Pool) *AuthRepository {
	return &AuthRepository{
		pool: pool,
	}
}

func (r *AuthRepository) CreateSession(
	ctx context.Context,
	userID uuid.UUID,
) (models.Session, error) {

	ctx, cancel := context.WithTimeout(ctx, r.pool.OpTimeout())
	defer cancel()

	now := time.Now().UTC()

	session := models.Session{
		ID:         uuid.New(),
		UserID:     userID,
		CreatedAt:  now,
		LastSeenAt: now,
		ExpiresAt:  now.Add(30 * 24 * time.Hour),
	}

	const query = `
		INSERT INTO sessions (
			id,
			user_id,
			created_at,
			last_seen_at,
			expires_at
		)
		VALUES ($1, $2, $3, $4, $5)
	`

	_, err := r.pool.Exec(
		ctx,
		query,
		session.ID,
		session.UserID,
		session.CreatedAt,
		session.LastSeenAt,
		session.ExpiresAt,
	)
	if err != nil {
		return models.Session{}, err
	}

	return session, nil
}

func (r *AuthRepository) GetSession(
	ctx context.Context,
	sessionID uuid.UUID,
) (models.Session, error) {

	ctx, cancel := context.WithTimeout(ctx, r.pool.OpTimeout())
	defer cancel()

	const query = `
		SELECT
			id,
			user_id,
			created_at,
			last_seen_at,
			expires_at
		FROM sessions
		WHERE id = $1
	`

	var session models.Session

	err := r.pool.QueryRow(
		ctx,
		query,
		sessionID,
	).Scan(
		&session.ID,
		&session.UserID,
		&session.CreatedAt,
		&session.LastSeenAt,
		&session.ExpiresAt,
	)

	if err != nil {
		return models.Session{}, err
	}

	return session, nil
}

func (r *AuthRepository) UpdateLastSeen(
	ctx context.Context,
	sessionID uuid.UUID,
) error {

	ctx, cancel := context.WithTimeout(ctx, r.pool.OpTimeout())
	defer cancel()

	const query = `
		UPDATE sessions
		SET last_seen_at = $1
		WHERE id = $2
	`

	_, err := r.pool.Exec(
		ctx,
		query,
		time.Now().UTC(),
		sessionID,
	)

	return err
}

func (r *AuthRepository) DeleteSession(
	ctx context.Context,
	sessionID uuid.UUID,
) error {

	ctx, cancel := context.WithTimeout(ctx, r.pool.OpTimeout())
	defer cancel()

	const query = `
		DELETE FROM sessions
		WHERE id = $1
	`

	_, err := r.pool.Exec(
		ctx,
		query,
		sessionID,
	)

	return err
}
