package auth

import (
	"context"

	"github.com/google/uuid"

	"github.com/34Minnesota/avito-antiscam-monorepo/backend/internal/domain/models"
)

// Repository описывает контракт хранения пользовательских сессий.
type Repository interface {
	// CreateSession создает новую сессию пользователя.
	CreateSession(
		ctx context.Context,
		userID uuid.UUID,
	) (models.Session, error)

	// GetSession возвращает сессию по ID.
	GetSession(
		ctx context.Context,
		sessionID uuid.UUID,
	) (models.Session, error)

	// UpdateLastSeen обновляет время последней активности.
	UpdateLastSeen(
		ctx context.Context,
		sessionID uuid.UUID,
	) error

	// DeleteSession удаляет сессию (logout).
	DeleteSession(
		ctx context.Context,
		sessionID uuid.UUID,
	) error
}
