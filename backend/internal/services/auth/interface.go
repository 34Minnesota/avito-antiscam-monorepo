package auth

import (
	"context"

	"avito-antiscam/backend/internal/domain"
)

// Service описывает всё,
// что умеет модуль авторизации.
type Service interface {

	// Создает гостевого пользователя
	// и возвращает JWT.
	CreateGuest(ctx context.Context) (string, error)

	// Получает пользователя
	// по JWT.
	GetUserFromToken(ctx context.Context, token string) (*domain.User, error)
}