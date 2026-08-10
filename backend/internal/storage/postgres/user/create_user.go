package users_repository

import (
	"context"
	"errors"
	"fmt"

	domainErrors "github.com/34Minnesota/avito-antiscam-monorepo/backend/internal/domain/errors"
	"github.com/34Minnesota/avito-antiscam-monorepo/backend/internal/domain/models"
	"github.com/jackc/pgx/v5/pgconn"
)

func (r *UsersRepository) CreateUser(ctx context.Context, user models.User) error {
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
		return fmt.Errorf("%w: %s", domainErrors.ErrConflict, pgErr.ConstraintName)
	}

	return err
}
