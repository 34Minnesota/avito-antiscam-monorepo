package users_service

import (
	"context"
	"fmt"

	"github.com/34Minnesota/avito-antiscam-monorepo/backend/internal/domain"
	"github.com/google/uuid"
)

func (s *UsersService) GetUser(
	ctx context.Context,
	userID uuid.UUID,
) (domain.User, error) {
	user, err := s.usersRepository.GetUser(ctx, userID)
	if err != nil {
		return domain.User{}, fmt.Errorf("get user from repository: %w", err)
	}

	return user, nil
}
