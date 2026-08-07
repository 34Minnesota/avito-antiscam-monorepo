package users_service

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/34Minnesota/avito-antiscam-monorepo/backend/internal/domain"
	domainErrors "github.com/34Minnesota/avito-antiscam-monorepo/backend/internal/domain/errors"
)

type UsersService struct {
	usersRepository UsersRepository
	passwordHasher  PasswordHasher
	idGenerator     IDGenerator
	clock           Clock
}

type PasswordHasher interface {
	Hash(password string) (string, error)
}

type IDGenerator interface {
	New() uuid.UUID
}

type Clock interface {
	Now() time.Time
}

type UsersRepository interface {
	CreateUser(ctx context.Context, user domain.User) error
	GetUser(ctx context.Context, userID uuid.UUID) (domain.User, error)
}

func NewUsersService(
	usersRepository UsersRepository,
	passwordHasher PasswordHasher,
	idGenerator IDGenerator,
	clock Clock,
) *UsersService {
	return &UsersService{
		usersRepository: usersRepository,
		passwordHasher:  passwordHasher,
		idGenerator:     idGenerator,
		clock:           clock,
	}
}

func validateCreateUserInput(input CreateUserInput) error {
	if err := domain.ValidateNickname(input.Nickname); err != nil {
		return err
	}

	if err := domain.ValidateEmail(input.Email); err != nil {
		return err
	}

	if len([]rune(input.Password)) < domain.MinPasswordLength ||
		len([]byte(input.Password)) > domain.MaxPasswordBytes {
		return fmt.Errorf("invalid password length: %w", domainErrors.ErrInvalidArgument)
	}

	return nil
}

func normalizeCreateUserInput(input CreateUserInput) CreateUserInput {
	input.Email = strings.ToLower(strings.TrimSpace(input.Email))
	return input
}
