package identity

import (
	"context"

	"github.com/34Minnesota/avito-antiscam-monorepo/backend/internal/domain/models"
)

type userIDKey struct{}

func WithUserID(ctx context.Context, userID models.UserID) context.Context {
	return context.WithValue(ctx, userIDKey{}, userID)
}

func UserID(ctx context.Context) (models.UserID, bool) {
	userID, ok := ctx.Value(userIDKey{}).(models.UserID)

	return userID, ok && !userID.IsZero()
}
