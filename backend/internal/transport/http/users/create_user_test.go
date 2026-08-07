package users_http_transport

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestPublicUserRoutesAreDisabled(t *testing.T) {
	t.Parallel()

	gin.SetMode(gin.TestMode)

	router := gin.New()
	RegisterUsersRoutes(router, nil)

	tests := []struct {
		name   string
		method string
		path   string
	}{
		{
			name:   "create user",
			method: http.MethodPost,
			path:   "/v1/users",
		},
		{
			name:   "get user",
			method: http.MethodGet,
			path:   "/v1/user/00000000-0000-0000-0000-000000000000",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()

			req := httptest.NewRequestWithContext(
				ctx,
				tt.method,
				tt.path,
				nil,
			)
			resp := httptest.NewRecorder()

			router.ServeHTTP(resp, req)

			if resp.Code != http.StatusNotFound {
				t.Fatalf("status = %d, want %d", resp.Code, http.StatusNotFound)
			}
		})
	}
}
