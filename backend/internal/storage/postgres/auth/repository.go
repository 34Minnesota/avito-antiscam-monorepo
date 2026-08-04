package auth

import (
	"context"
	"time"

	"github.com/google/uuid"

	"github.com/34Minnesota/avito-antiscam-monorepo/backend/internal/domain"
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

func (r *AuthRepository) Create(
	ctx context.Context,
) (domain.Session, error) {
	ctx, cancel := context.WithTimeout(ctx, r.pool.OpTimeout())
	defer cancel()

	now := time.Now().UTC()
	session := domain.Session{
		ID:         uuid.New(),
		CreatedAt:  now,
		LastSeenAt: now,
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
	ctx, cancel := context.WithTimeout(ctx, r.pool.OpTimeout())
	defer cancel()

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
		return domain.Session{}, err
	}

	return session, nil
}

func (r *AuthRepository) Touch(
	ctx context.Context,
	id uuid.UUID,
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
		id,
	)

	return err
}
