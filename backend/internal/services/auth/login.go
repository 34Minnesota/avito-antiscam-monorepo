package auth

import (
	"context"

	domainErrors "github.com/34Minnesota/avito-antiscam-monorepo/backend/internal/domain/errors"
	"github.com/34Minnesota/avito-antiscam-monorepo/backend/internal/domain/models"
)

// Login авторизует пользователя по email и паролю
// и создает новую пользовательскую сессию.
func (s *Service) Login(
	ctx context.Context,
	email string,
	password string,
) (models.Session, error) {

	user, err := s.users.GetUserByEmail(ctx, email)
	if err != nil {
		return models.Session{}, domainErrors.ErrInvalidCredentials
	}

	if err := s.verifier.Compare(
		user.PasswordHash,
		password,
	); err != nil {
		return models.Session{}, domainErrors.ErrInvalidCredentials
	}

	session, err := s.repository.CreateSession(
		ctx,
		user.ID,
	)
	if err != nil {
		return models.Session{}, err
	}

	return session, nil
}
