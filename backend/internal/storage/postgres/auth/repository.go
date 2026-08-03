package auth

import (
	"context"

	"avito-antiscam/backend/internal/domain"

	"github.com/google/uuid"
)

// Repository отвечает только
// за работу с базой данных.
type Repository interface {

	CreateGuest(
		ctx context.Context,
		user *domain.User,
	) error

	GetByID(
		ctx context.Context,
		id uuid.UUID,
	) (*domain.User, error)
}