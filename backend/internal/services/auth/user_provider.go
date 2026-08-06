package auth

import (
	"context"

	"github.com/google/uuid"

	"github.com/34Minnesota/avito-antiscam-monorepo/backend/internal/domain"
)

// UserProvider описывает минимальный контракт,
// необходимый сервису авторизации.
type UserProvider interface {
	GetUser(
		ctx context.Context,
		userID uuid.UUID,
	) (domain.User, error)

	GetUserByEmail(
		ctx context.Context,
		email string,
	) (domain.User, error)
}
