package domain

import "github.com/google/uuid"

// UserID нужен чтобы не передавать обычный uuid.UUID без контекста в scoring/progress
// TODO(owner: auth): надо класть userID в контекст http запросов
type UserID uuid.UUID

func NewUserID(id uuid.UUID) (UserID, error) {
	if id == uuid.Nil {
		return UserID{}, ErrInvalidUserID
	}

	return UserID(id), nil
}

func (id UserID) UUID() uuid.UUID {
	return uuid.UUID(id)
}

func (id UserID) IsZero() bool {
	return id.UUID() == uuid.Nil
}
