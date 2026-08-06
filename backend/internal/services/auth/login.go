package auth

import (
	"context"

	"github.com/34Minnesota/avito-antiscam-monorepo/backend/internal/domain"
)

// Login авторизует пользователя по email и паролю
// и создает новую пользовательскую сессию.
func (s *Service) Login(
	ctx context.Context,
	email string,
	password string,
) (domain.Session, error) {

	user, err := s.users.GetUserByEmail(ctx, email)
	if err != nil {
		return domain.Session{}, ErrInvalidCredentials
	}

	if err := s.verifier.Compare(
		user.PasswordHash,
		password,
	); err != nil {
		return domain.Session{}, ErrInvalidCredentials
	}

	session, err := s.repository.CreateSession(
		ctx,
		user.ID,
	)
	if err != nil {
		return domain.Session{}, err
	}

	return session, nil
}
