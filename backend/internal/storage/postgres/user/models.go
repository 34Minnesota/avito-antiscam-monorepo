package users_repository

import (
	"time"

	"github.com/34Minnesota/avito-antiscam-monorepo/backend/internal/domain"
	"github.com/google/uuid"
)

type UserModel struct {
	ID           uuid.UUID
	Nickname     string
	Email        string
	PasswordHash string
	CreatedAt    time.Time
}

func userDomainFromModel(user UserModel) domain.User {
	return domain.NewUser(
		user.ID, user.Nickname, user.Email,
		user.PasswordHash, user.CreatedAt,
	)
}

// func userDomainsFromModels(users []UserModel) []domain.User {
// 	userDomains := make([]domain.User, len(users))

// 	for i, user := range users {
// 		userDomains[i] = userDomainFromModel(user)
// 	}

// 	return userDomains
// }
