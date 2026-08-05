package identity

import (
	"context"

	"github.com/34Minnesota/avito-antiscam-monorepo/backend/internal/domain"
)

type userIDKey struct{}

// TODO(owner: auth): провалидировать JWT и вызвать это в ручке
func WithUserID(ctx context.Context, userID domain.UserID) context.Context {
	return context.WithValue(ctx, userIDKey{}, userID)
}

func UserID(ctx context.Context) (domain.UserID, bool) {
	userID, ok := ctx.Value(userIDKey{}).(domain.UserID)

	return userID, ok && !userID.IsZero()
}
