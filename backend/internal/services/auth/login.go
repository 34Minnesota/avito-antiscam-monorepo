package auth

import (
	"context"
	"errors"

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

// Authenticate является алиасом Login.
// Оставлен для совместимости, если понадобится
// использовать другое название в transport.
func (s *Service) Authenticate(
	ctx context.Context,
	email string,
	password string,
) (domain.Session, error) {

	session, err := s.Login(ctx, email, password)
	if err != nil {
		if errors.Is(err, ErrInvalidCredentials) {
			return domain.Session{}, ErrInvalidCredentials
		}
		return domain.Session{}, err
	}

	return session, nil
}
