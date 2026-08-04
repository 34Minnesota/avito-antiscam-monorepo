// Package identity хранит идентификатор пользователя в контексте запроса.
//
// Ключ — приватный тип, а не строка: чужой пакет не может его подделать
// или случайно перезаписать одноимённым ключом.
package identity

import (
	"context"

	"github.com/34Minnesota/avito-antiscam-monorepo/backend/internal/domain"
)

type userIDKey struct{}

func WithUserID(ctx context.Context, userID domain.UserID) context.Context {
	return context.WithValue(ctx, userIDKey{}, userID)
}

func UserID(ctx context.Context) (domain.UserID, bool) {
	userID, ok := ctx.Value(userIDKey{}).(domain.UserID)
	return userID, ok && !userID.IsZero()
}
