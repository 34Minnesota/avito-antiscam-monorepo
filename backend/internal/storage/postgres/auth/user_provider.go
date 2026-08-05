package auth

import (
	"context"

	"github.com/google/uuid"

	"github.com/34Minnesota/avito-antiscam-monorepo/backend/internal/domain"
)

// UserProvider описывает минимальный контракт,
// необходимый сервису авторизации для работы
// с пользователями.
//
// Реализацию этого интерфейса предоставляет
// сервис users.
type UserProvider interface {
	GetByID(
		ctx context.Context,
		id uuid.UUID,
	) (domain.User, error)

	GetByEmail(
		ctx context.Context,
		email string,
	) (domain.User, error)
}
