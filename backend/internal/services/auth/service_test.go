package auth

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"

	"github.com/34Minnesota/avito-antiscam-monorepo/backend/internal/domain"
)

type mockUserProvider struct {
	user domain.User
	err  error
}

func (m *mockUserProvider) GetUser(
	ctx context.Context,
	userID uuid.UUID,
) (domain.User, error) {
	return m.user, m.err
}

func (m *mockUserProvider) GetUserByEmail(
	ctx context.Context,
	email string,
) (domain.User, error) {
	return m.user, m.err
}

type mockPasswordVerifier struct {
	err error
}

func (m *mockPasswordVerifier) Compare(
	hash string,
	password string,
) error {
	return m.err
}

func TestLoginSuccess(t *testing.T) {
	user := domain.User{
		ID:           uuid.New(),
		Email:        "user@test.ru",
		PasswordHash: "hash",
	}

	session := domain.Session{
		ID:     uuid.New(),
		UserID: user.ID,
	}

	repository := &MockRepository{
		session: session,
	}

	users := &mockUserProvider{
		user: user,
	}

	verifier := &mockPasswordVerifier{}

	service := NewService(
		repository,
		users,
		verifier,
		nil,
	)

	result, err := service.Login(
		context.Background(),
		user.Email,
		"password",
	)
	if err != nil {
		t.Fatal(err)
	}

	if result.ID != session.ID {
		t.Fatal("unexpected session id")
	}
}

func TestLoginUserNotFound(t *testing.T) {
	repository := &MockRepository{}

	users := &mockUserProvider{
		err: errors.New("not found"),
	}

	verifier := &mockPasswordVerifier{}

	service := NewService(
		repository,
		users,
		verifier,
		nil,
	)

	_, err := service.Login(
		context.Background(),
		"user@test.ru",
		"password",
	)

	if !errors.Is(err, ErrInvalidCredentials) {
		t.Fatal("expected invalid credentials")
	}
}

func TestLoginWrongPassword(t *testing.T) {
	user := domain.User{
		ID:           uuid.New(),
		Email:        "user@test.ru",
		PasswordHash: "hash",
	}

	repository := &MockRepository{}

	users := &mockUserProvider{
		user: user,
	}

	verifier := &mockPasswordVerifier{
		err: errors.New("wrong password"),
	}

	service := NewService(
		repository,
		users,
		verifier,
		nil,
	)

	_, err := service.Login(
		context.Background(),
		user.Email,
		"password",
	)

	if !errors.Is(err, ErrInvalidCredentials) {
		t.Fatal("expected invalid credentials")
	}
}
