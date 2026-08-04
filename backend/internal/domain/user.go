package domain

import (
	"time"

	domainErrors "github.com/34Minnesota/avito-antiscam-monorepo/backend/internal/domain/errors"
	"github.com/google/uuid"
)

type User struct {
	ID        uuid.UUID
	Nickname  string
	CreatedAt time.Time
}

// UserID нужен чтобы не передавать обычный uuid.UUID без контекста в scoring/progress
// TODO(owner: auth): надо класть userID в контекст http запросов
type UserID uuid.UUID

func NewUserID(id uuid.UUID) (UserID, error) {
	if id == uuid.Nil {
		return UserID{}, domainErrors.ErrInvalidUserID
	}

	return UserID(id), nil
}

func (id UserID) UUID() uuid.UUID {
	return uuid.UUID(id)
}

func (id UserID) IsZero() bool {
	return id.UUID() == uuid.Nil
}
