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

	"github.com/34Minnesota/avito-antiscam-monorepo/backend/internal/domain"
	domainErrors "github.com/34Minnesota/avito-antiscam-monorepo/backend/internal/domain/errors"
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
	var body progressResponse
	if err := json.Unmarshal(response.Body.Bytes(), &body); err != nil {
		t.Fatal(err)
	}
	if body.TotalScenarios != 0 || body.Roles == nil {
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
		{name: "dependency", err: domainErrors.ErrDependencyUnavailable, status: http.StatusServiceUnavailable},
		{name: "inconsistent", err: domainErrors.ErrDataInconsistent, status: http.StatusInternalServerError},
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
	if authenticated {
		userID, err := domain.NewUserID(uuid.New())
		if err != nil {
			t.Fatal(err)
		}
		router.Use(func(c *gin.Context) {
			c.Set(sessionContextKey, domain.Session{ID: uuid.New(), UserID: userID.UUID()})
		})
	}
	RegisterProgressRoutes(router.Group("/v1"), NewProgressHandler(service))
	request := httptest.NewRequestWithContext(context.Background(), http.MethodGet, "/v1/progress", nil)
	response := httptest.NewRecorder()
	router.ServeHTTP(response, request)
	return response
}
