package users_http_transport

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/34Minnesota/avito-antiscam-monorepo/backend/internal/domain"
	coreErrors "github.com/34Minnesota/avito-antiscam-monorepo/backend/internal/errors"
)

type serviceStub struct {
	user  domain.User
	err   error
	calls int
}

func (s *serviceStub) CreateUser(_ context.Context, _ CreateUserInput) (domain.User, error) {
	s.calls++
	return s.user, s.err
}

func performCreateUserRequest(t *testing.T, service UsersService, body string) *httptest.ResponseRecorder {
	t.Helper()
	gin.SetMode(gin.TestMode)
	router := gin.New()
	RegisterUsersRoutes(router, NewUsersHandler(service))
	ctx := context.Background()
	request := httptest.NewRequestWithContext(
		ctx,
		http.MethodPost,
		"/v1/users",
		strings.NewReader(body),
	)
	request.Header.Set("Content-Type", "application/json")
	response := httptest.NewRecorder()
	router.ServeHTTP(response, request)
	return response
}

func TestCreateUserReturnsCreatedUserWithoutPasswordHash(t *testing.T) {
	t.Parallel()

	service := &serviceStub{user: domain.User{
		ID:        uuid.New(),
		Nickname:  "alice",
		Email:     "alice@example.com",
		CreatedAt: time.Now().UTC(),
	}}
	response := performCreateUserRequest(t, service, `{"nickname":"alice","email":"alice@example.com","password":"password123"}`)

	if response.Code != http.StatusCreated {
		t.Fatalf("status = %d, body = %s", response.Code, response.Body.String())
	}
	var body map[string]any
	if err := json.Unmarshal(response.Body.Bytes(), &body); err != nil {
		t.Fatal(err)
	}
	if _, exists := body["password_hash"]; exists {
		t.Fatal("response contains password_hash")
	}
	if service.calls != 1 {
		t.Fatalf("service calls = %d, want 1", service.calls)
	}
}

func TestCreateUserReturnsBadRequestForMalformedOrInvalidJSON(t *testing.T) {
	t.Parallel()

	for _, body := range []string{
		`{"nickname":`,
		`{"nickname":"ab","email":"alice@example.com","password":"password123"}`,
	} {
		service := &serviceStub{}
		response := performCreateUserRequest(t, service, body)
		if response.Code != http.StatusBadRequest {
			t.Fatalf("body %q: status = %d", body, response.Code)
		}
		if service.calls != 0 {
			t.Fatalf("body %q: service calls = %d", body, service.calls)
		}
	}
}

func TestCreateUserMapsServiceErrors(t *testing.T) {
	t.Parallel()

	for _, test := range []struct {
		name   string
		err    error
		status int
	}{
		{name: "invalid argument", err: coreErrors.ErrInvalidArgument, status: http.StatusBadRequest},
		{name: "conflict", err: coreErrors.ErrConflict, status: http.StatusConflict},
		{name: "internal", err: errors.New("boom"), status: http.StatusInternalServerError},
	} {
		t.Run(test.name, func(t *testing.T) {
			response := performCreateUserRequest(t, &serviceStub{err: test.err}, `{"nickname":"alice","email":"alice@example.com","password":"password123"}`)
			if response.Code != test.status {
				t.Fatalf("status = %d, want %d", response.Code, test.status)
			}
		})
	}
}
