package auth

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/34Minnesota/avito-antiscam-monorepo/backend/internal/domain"
)

type MockRepository struct {
	session domain.Session
	err     error
}

func (m *MockRepository) Create(
	ctx context.Context,
) (domain.Session, error) {
	return m.session, m.err
}

func (m *MockRepository) Get(
	ctx context.Context,
	id uuid.UUID,
) (domain.Session, error) {
	return m.session, m.err
}

func (m *MockRepository) Touch(
	ctx context.Context,
	id uuid.UUID,
) error {
	return m.err
}

func TestCreateSession(t *testing.T) {

	expected := domain.Session{
		ID:         uuid.New(),
		CreatedAt:  time.Now(),
		LastSeenAt: time.Now(),
	}

	repository := &MockRepository{
		session: expected,
	}

	service := NewService(repository)

	session, err := service.CreateSession(context.Background())
	if err != nil {
		t.Fatal(err)
	}

	if session.ID != expected.ID {
		t.Fatal("unexpected session id")
	}
}

func TestValidateSessionSuccess(t *testing.T) {

	expected := domain.Session{
		ID:         uuid.New(),
		CreatedAt:  time.Now(),
		LastSeenAt: time.Now(),
	}

	repository := &MockRepository{
		session: expected,
	}

	service := NewService(repository)

	session, err := service.ValidateSession(
		context.Background(),
		expected.ID,
	)
	if err != nil {
		t.Fatal(err)
	}

	if session.ID != expected.ID {
		t.Fatal("unexpected session id")
	}
}

func TestValidateSessionFail(t *testing.T) {

	repository := &MockRepository{
		err: errors.New("session not found"),
	}

	service := NewService(repository)

	_, err := service.ValidateSession(
		context.Background(),
		uuid.New(),
	)

	if err == nil {
		t.Fatal("expected error")
	}
}