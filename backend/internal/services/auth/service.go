package auth

import (
	"context"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"

	"github.com/34Minnesota/avito-antiscam-monorepo/backend/internal/domain"
	applogger "github.com/34Minnesota/avito-antiscam-monorepo/backend/internal/logger"
)

// Service отвечает за бизнес-логику авторизации
// и работы с пользовательскими сессиями.
type Service struct {
	repository Repository
	logger     *applogger.Logger
}

// NewService создает сервис авторизации.
func NewService(
	repository Repository,
	logger *applogger.Logger,
) *Service {
	return &Service{
		repository: repository,
		logger:     logger,
	}
}

// CreateSession создает новую сессию пользователя.
func (s *Service) CreateSession(
	ctx context.Context,
	userID uuid.UUID,
) (domain.Session, error) {

	session, err := s.repository.CreateSession(ctx, userID)
	if err != nil {
		if s.logger != nil {
			s.logger.Error(
				"create session failed",
				zap.String("user_id", userID.String()),
				zap.Error(err),
			)
		}
		return domain.Session{}, err
	}

	return session, nil
}

// ValidateSession проверяет существование сессии
// и обновляет время последней активности.
func (s *Service) ValidateSession(
	ctx context.Context,
	sessionID uuid.UUID,
) (domain.Session, error) {

	session, err := s.repository.GetSession(ctx, sessionID)
	if err != nil {
		if s.logger != nil {
			s.logger.Warn(
				"validate session failed",
				zap.String("session_id", sessionID.String()),
				zap.Error(err),
			)
		}
		return domain.Session{}, err
	}

	if err := s.repository.UpdateLastSeen(ctx, sessionID); err != nil {
		if s.logger != nil {
			s.logger.Error(
				"update last seen failed",
				zap.String("session_id", sessionID.String()),
				zap.Error(err),
			)
		}
		return domain.Session{}, err
	}

	session.LastSeenAt = time.Now().UTC()

	return session, nil
}

// Logout завершает пользовательскую сессию.
func (s *Service) Logout(
	ctx context.Context,
	sessionID uuid.UUID,
) error {

	err := s.repository.DeleteSession(ctx, sessionID)
	if err != nil {
		if s.logger != nil {
			s.logger.Error(
				"logout failed",
				zap.String("session_id", sessionID.String()),
				zap.Error(err),
			)
		}
	}

	return err
}
