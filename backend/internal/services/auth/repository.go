package auth

import (
	"github.com/google/uuid"

	"github.com/34Minnesota/avito-antiscam-monorepo/backend/internal/domain"
)

type Repository interface {

	Create() domain.Session

	Get(id uuid.UUID) (domain.Session, bool)

	Touch(id uuid.UUID) bool
}