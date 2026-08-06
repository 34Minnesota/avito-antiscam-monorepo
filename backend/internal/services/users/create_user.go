package users_service

import (
	"context"
	"errors"
	"fmt"

	"github.com/34Minnesota/avito-antiscam-monorepo/backend/internal/domain"
	coreErrors "github.com/34Minnesota/avito-antiscam-monorepo/backend/internal/errors"
)

type CreateUserInput struct {
	Nickname string
	Email    string
	Password string
}

func (s *UsersService) CreateUser(
	ctx context.Context,
	userInput CreateUserInput,
) (domain.User, error) {
	if err := validateCreateUserInput(userInput); err != nil {
		return domain.User{}, err
	}

	userInput = normalizeCreateUserInput(userInput)

	passwordHash, err := s.passwordHasher.Hash(userInput.Password)
	if err != nil {
		return domain.User{}, fmt.Errorf("hash password: %w", err)
	}

	user := domain.NewUser(
		s.idGenerator.New(),
		userInput.Nickname,
		userInput.Email,
		passwordHash,
		s.clock.Now().UTC(),
	)
	if err := user.Validate(); err != nil {
		return domain.User{}, err
	}

	if err := s.usersRepository.CreateUser(ctx, user); err != nil {
		if errors.Is(err, coreErrors.ErrConflict) {
			return domain.User{}, fmt.Errorf("create user: %w", coreErrors.ErrConflict)
		}
		return domain.User{}, fmt.Errorf("create user: %w", err)
	}

	return user, nil
}
