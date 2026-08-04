package auth

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/34Minnesota/avito-antiscam-monorepo/backend/internal/domain"
)

type AuthRepository struct {
	pool *pgxpool.Pool
}

func NewRepository(pool *pgxpool.Pool) *AuthRepository {
	return &AuthRepository{
		pool: pool,
	}
}

func (r *AuthRepository) Create(
	ctx context.Context,
) (domain.Session, error) {

	session := domain.Session{
		ID:         uuid.New(),
		CreatedAt:  time.Now().UTC(),
		LastSeenAt: time.Now().UTC(),
	}

	const query = `
		INSERT INTO sessions (
			id,
			created_at,
			last_seen_at
		)
		VALUES ($1, $2, $3)
	`

	_, err := r.pool.Exec(
		ctx,
		query,
		session.ID,
		session.CreatedAt,
		session.LastSeenAt,
	)
	if err != nil {
		return domain.Session{}, err
	}

	return session, nil
}

func (r *AuthRepository) Get(
	ctx context.Context,
	id uuid.UUID,
) (domain.Session, error) {

	const query = `
		SELECT
			id,
			created_at,
			last_seen_at
		FROM sessions
		WHERE id = $1
	`

	var session domain.Session

	err := r.pool.QueryRow(
		ctx,
		query,
		id,
	).Scan(
		&session.ID,
		&session.CreatedAt,
		&session.LastSeenAt,
	)

	if err != nil {
		if err == pgx.ErrNoRows {
			return domain.Session{}, err
		}

		return domain.Session{}, err
	}

	return session, nil
}

func (r *AuthRepository) Touch(
	ctx context.Context,
	id uuid.UUID,
) error {

	const query = `
		UPDATE sessions
		SET last_seen_at = $1
		WHERE id = $2
	`

	_, err := r.pool.Exec(
		ctx,
		query,
		time.Now().UTC(),
		id,
	)

	return err
}