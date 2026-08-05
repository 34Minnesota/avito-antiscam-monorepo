package users_repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/34Minnesota/avito-antiscam-monorepo/backend/internal/domain"
	core_errors "github.com/34Minnesota/avito-antiscam-monorepo/backend/internal/errors"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

func (r *UsersRepository) GetUser(
	ctx context.Context,
	userID uuid.UUID,
) (domain.User, error) {
	ctx, cancel := context.WithTimeout(ctx, r.pool.OpTimeout())
	defer cancel()

	query := `
	SELECT id, nickname, email, password_hash, created_at 
	FROM users 
	WHERE id=$1
	`

	row := r.pool.QueryRow(ctx, query, userID)

	var userModel UserModel
	err := row.Scan(
		&userModel.ID,
		&userModel.Nickname,
		&userModel.Email,
		&userModel.PasswordHash,
		&userModel.CreatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.User{}, fmt.Errorf(
				"user with id %d not found: %w",
				userID, core_errors.ErrNotFound,
			)
		}

		return domain.User{}, fmt.Errorf("scan user: %w", err)
	}

	userDomain := userDomainFromModel(userModel)

	return userDomain, nil
}
