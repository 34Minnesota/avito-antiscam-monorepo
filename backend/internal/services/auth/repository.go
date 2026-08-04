package auth

import (
	"context"

	"github.com/google/uuid"

	"github.com/34Minnesota/avito-antiscam-monorepo/backend/internal/domain"
)

// Repository описывает контракт,
// который должен реализовать
// любой источник хранения сессий.
type Repository interface {

	// Create создает новую сессию.
	Create(ctx context.Context) (domain.Session, error)

	// Get возвращает сессию по ID.
	Get(
		ctx context.Context,
		id uuid.UUID,
	) (domain.Session, error)

	// Touch обновляет время
	// последней активности.
	Touch(
		ctx context.Context,
		id uuid.UUID,
	) error
}