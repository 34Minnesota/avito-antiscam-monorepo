package users_repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/34Minnesota/avito-antiscam-monorepo/backend/internal/domain"
	core_errors "github.com/34Minnesota/avito-antiscam-monorepo/backend/internal/errors"
	"github.com/jackc/pgx/v5/pgconn"
)

func (r *UsersRepository) CreateUser(ctx context.Context, user domain.User) error {
	ctx, cancel := context.WithTimeout(ctx, r.pool.OpTimeout())
	defer cancel()

	const query = `
		INSERT INTO users (id, nickname, email, password_hash, created_at)
		VALUES ($1, $2, $3, $4, $5)
	`

	_, err := r.pool.Exec(ctx, query,
		user.ID,
		user.Nickname,
		user.Email,
		user.PasswordHash,
		user.CreatedAt,
	)
	if err == nil {
		return nil
	}

	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) && pgErr.Code == "23505" {
		return fmt.Errorf("%w: %s", core_errors.ErrConflict, pgErr.ConstraintName)
	}

	return err
}
