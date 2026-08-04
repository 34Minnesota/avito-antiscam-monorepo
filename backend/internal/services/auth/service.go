package auth

import (
	"context"
	"time"

	"github.com/google/uuid"

	"github.com/34Minnesota/avito-antiscam-monorepo/backend/internal/domain"
)

// Service отвечает за бизнес-логику
// авторизации и работы с сессиями.
type Service struct {
	repository Repository
}

// NewService создает сервис авторизации.
func NewService(repository Repository) *Service {
	return &Service{
		repository: repository,
	}
}

// CreateSession создает новую гостевую сессию.
func (s *Service) CreateSession(
	ctx context.Context,
) (domain.Session, error) {
	return s.repository.Create(ctx)
}

// ValidateSession проверяет существование сессии
// и обновляет время последней активности.
func (s *Service) ValidateSession(
	ctx context.Context,
	id uuid.UUID,
) (domain.Session, error) {

	session, err := s.repository.Get(ctx, id)
	if err != nil {
		return domain.Session{}, err
	}

	if err := s.repository.Touch(ctx, id); err != nil {
		return domain.Session{}, err
	}

	// Возвращаем уже актуальную сессию.
	session.LastSeenAt = time.Now().UTC()

	return session, nil
}
