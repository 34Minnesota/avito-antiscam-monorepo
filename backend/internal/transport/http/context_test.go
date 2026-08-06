package httptransport

import (
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/34Minnesota/avito-antiscam-monorepo/backend/internal/domain"
)

func TestCurrentUserUsesSessionUserID(t *testing.T) {
	userID := uuid.New()
	ctx, _ := gin.CreateTestContext(httptest.NewRecorder())
	ctx.Set(sessionContextKey, domain.Session{ID: uuid.New(), UserID: userID})

	got, ok := CurrentUser(ctx)
	if !ok {
		t.Fatal("expected current user")
	}
	if got.UUID() != userID {
		t.Fatalf("user id = %s, want %s", got.UUID(), userID)
	}
}

func TestCurrentUserRejectsMissingOrInvalidSession(t *testing.T) {
	tests := map[string]func(*gin.Context){
		"missing": func(*gin.Context) {},
		"wrong type": func(ctx *gin.Context) {
			ctx.Set(sessionContextKey, "not a session")
		},
		"empty session": func(ctx *gin.Context) {
			ctx.Set(sessionContextKey, domain.Session{})
		},
		"session without user": func(ctx *gin.Context) {
			ctx.Set(sessionContextKey, domain.Session{ID: uuid.New()})
		},
	}

	for name, arrange := range tests {
		t.Run(name, func(t *testing.T) {
			ctx, _ := gin.CreateTestContext(httptest.NewRecorder())
			arrange(ctx)

			if _, ok := CurrentUser(ctx); ok {
				t.Fatal("expected missing current user")
			}
		})
	}
}
