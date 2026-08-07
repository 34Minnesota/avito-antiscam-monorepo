package auth

import (
	"context"

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
