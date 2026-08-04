package auth

import (
	"context"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"

	"github.com/34Minnesota/avito-antiscam-monorepo/backend/internal/domain"
	applogger "github.com/34Minnesota/avito-antiscam-monorepo/backend/internal/logger"
)

// Service отвечает за бизнес-логику
// авторизации и работы с сессиями.
type Service struct {
	repository Repository
	logger     *applogger.Logger
}

// NewService создает сервис авторизации.
func NewService(repository Repository, logger *applogger.Logger) *Service {
	return &Service{
		repository: repository,
		logger:     logger,
	}
}

// CreateSession создает новую гостевую сессию.
func (s *Service) CreateSession(
	ctx context.Context,
) (domain.Session, error) {
	session, err := s.repository.Create(ctx)
	if err != nil && s.logger != nil {
		s.logger.Error("create session failed", zap.Error(err))
	}

	return session, err
}

// ValidateSession проверяет существование сессии
// и обновляет время последней активности.
func (s *Service) ValidateSession(
	ctx context.Context,
	id uuid.UUID,
) (domain.Session, error) {

	session, err := s.repository.Get(ctx, id)
	if err != nil {
		if s.logger != nil {
			s.logger.Warn("validate session failed", zap.String("session_id", id.String()), zap.Error(err))
		}
		return domain.Session{}, err
	}

	if err := s.repository.Touch(ctx, id); err != nil {
		if s.logger != nil {
			s.logger.Error("touch session failed", zap.String("session_id", id.String()), zap.Error(err))
		}
		return domain.Session{}, err
	}

	// Возвращаем уже актуальную сессию.
	session.LastSeenAt = time.Now().UTC()

	return session, nil
}
