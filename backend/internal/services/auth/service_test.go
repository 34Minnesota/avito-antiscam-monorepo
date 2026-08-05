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

func (m *MockRepository) CreateSession(
	ctx context.Context,
	userID uuid.UUID,
) (domain.Session, error) {
	return m.session, m.err
}

func (m *MockRepository) GetSession(
	ctx context.Context,
	sessionID uuid.UUID,
) (domain.Session, error) {
	return m.session, m.err
}

func (m *MockRepository) UpdateLastSeen(
	ctx context.Context,
	sessionID uuid.UUID,
) error {
	return m.err
}

func (m *MockRepository) DeleteSession(
	ctx context.Context,
	sessionID uuid.UUID,
) error {
	return m.err
}

func TestCreateSession(t *testing.T) {
	expected := domain.Session{
		ID:         uuid.New(),
		UserID:     uuid.New(),
		CreatedAt:  time.Now().UTC(),
		LastSeenAt: time.Now().UTC(),
		ExpiresAt:  time.Now().UTC().Add(30 * 24 * time.Hour),
	}

	repository := &MockRepository{
		session: expected,
	}

	service := NewService(repository, nil)

	session, err := service.CreateSession(
		context.Background(),
		expected.UserID,
	)
	if err != nil {
		t.Fatal(err)
	}

	if session.ID != expected.ID {
		t.Fatal("unexpected session id")
	}

	if session.UserID != expected.UserID {
		t.Fatal("unexpected user id")
	}
}

func TestValidateSessionSuccess(t *testing.T) {
	expected := domain.Session{
		ID:         uuid.New(),
		UserID:     uuid.New(),
		CreatedAt:  time.Now().UTC(),
		LastSeenAt: time.Now().UTC(),
		ExpiresAt:  time.Now().UTC().Add(30 * 24 * time.Hour),
	}

	repository := &MockRepository{
		session: expected,
	}

	service := NewService(repository, nil)

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

	if session.UserID != expected.UserID {
		t.Fatal("unexpected user id")
	}
}

func TestValidateSessionFail(t *testing.T) {
	repository := &MockRepository{
		err: errors.New("session not found"),
	}

	service := NewService(repository, nil)

	_, err := service.ValidateSession(
		context.Background(),
		uuid.New(),
	)

	if err == nil {
		t.Fatal("expected error")
	}
}

func TestLogout(t *testing.T) {
	repository := &MockRepository{}

	service := NewService(repository, nil)

	err := service.Logout(
		context.Background(),
		uuid.New(),
	)

	if err != nil {
		t.Fatal(err)
	}
}
