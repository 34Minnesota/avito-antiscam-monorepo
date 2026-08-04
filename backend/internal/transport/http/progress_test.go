package httptransport

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/34Minnesota/avito-antiscam-monorepo/backend/generated/openapi"
	"github.com/34Minnesota/avito-antiscam-monorepo/backend/internal/domain"
	progressservice "github.com/34Minnesota/avito-antiscam-monorepo/backend/internal/services/progress"
	"github.com/34Minnesota/avito-antiscam-monorepo/backend/internal/transport/http/identity"
)

type progressStub struct {
	result domain.OverallProgress
	err    error
}

func (s progressStub) Get(context.Context, domain.UserID) (domain.OverallProgress, error) {
	return s.result, s.err
}

func TestProgressHandlerRequiresIdentity(t *testing.T) {
	t.Parallel()
	response := performRequest(t, progressStub{}, false)
	if response.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d", response.Code)
	}
}

func TestProgressHandlerMapsSuccess(t *testing.T) {
	t.Parallel()
	response := performRequest(t, progressStub{result: domain.OverallProgress{Roles: []domain.RoleProgress{}}}, true)
	if response.Code != http.StatusOK {
		t.Fatalf("status = %d, body = %s", response.Code, response.Body.String())
	}
	var body openapi.ProgressResponse
	if err := json.Unmarshal(response.Body.Bytes(), &body); err != nil {
		t.Fatal(err)
	}
	if body.TotalScenarios != 0 || body.Roles == nil || body.OutdatedActiveAttempts == nil {
		t.Fatalf("unexpected body: %+v", body)
	}
}

func TestProgressHandlerMapsErrors(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name   string
		err    error
		status int
	}{
		{name: "dependency", err: progressservice.ErrDependencyUnavailable, status: http.StatusServiceUnavailable},
		{name: "inconsistent", err: progressservice.ErrDataInconsistent, status: http.StatusInternalServerError},
		{name: "unknown", err: errors.New("boom"), status: http.StatusInternalServerError},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			response := performRequest(t, progressStub{err: tt.err}, true)
			if response.Code != tt.status {
				t.Fatalf("status = %d, want %d", response.Code, tt.status)
			}
		})
	}
}

func performRequest(t *testing.T, service ProgressGetter, authenticated bool) *httptest.ResponseRecorder {
	t.Helper()
	gin.SetMode(gin.TestMode)
	router := gin.New()
	RegisterProgressRoutes(router, NewProgressHandler(service))
	request := httptest.NewRequestWithContext(context.Background(), http.MethodGet, "/v1/progress", nil)
	if authenticated {
		userID, err := domain.NewUserID(uuid.New())
		if err != nil {
			t.Fatal(err)
		}
		request = request.WithContext(identity.WithUserID(request.Context(), userID))
	}
	response := httptest.NewRecorder()
	router.ServeHTTP(response, request)
	return response
}
