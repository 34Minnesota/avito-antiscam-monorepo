package auth

import (
	"context"

	"github.com/google/uuid"

	"github.com/34Minnesota/avito-antiscam-monorepo/backend/internal/domain/models"
)

// UserProvider описывает минимальный контракт,
// необходимый сервису авторизации.
type UserProvider interface {
	GetUser(
		ctx context.Context,
		userID uuid.UUID,
	) (models.User, error)

	GetUserByEmail(
		ctx context.Context,
		email string,
	) (models.User, error)
}
