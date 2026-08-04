package auth

import (
	"github.com/google/uuid"

	"github.com/34Minnesota/avito-antiscam-monorepo/backend/internal/domain"
)

type Service struct {
	repository Repository
}

func NewService(repository Repository) *Service {
	return &Service{
		repository: repository,
	}
}

// Создать новую анонимную сессию.
func (s *Service) CreateSession() domain.Session {
	return s.repository.Create()
}

// Проверить существование сессии и обновить время активности.
func (s *Service) ValidateSession(id uuid.UUID) (domain.Session, bool) {

	ok := s.repository.Touch(id)
	if !ok {
		return domain.Session{}, false
	}

	return s.repository.Get(id)
}