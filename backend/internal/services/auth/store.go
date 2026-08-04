package auth

import (
	"sync"
	"time"

	"github.com/google/uuid"

	"github.com/34Minnesota/avito-antiscam-monorepo/backend/internal/domain"
)

// Проверка, что SessionStore реализует Repository.
var _ Repository = (*SessionStore)(nil)

type SessionStore struct {
	mu sync.RWMutex

	sessions map[uuid.UUID]domain.Session
}

func NewSessionStore() *SessionStore {
	return &SessionStore{
		sessions: make(map[uuid.UUID]domain.Session),
	}
}

func (s *SessionStore) Create() domain.Session {

	now := time.Now()

	session := domain.Session{
		ID:         uuid.New(),
		CreatedAt:  now,
		LastSeenAt: now,
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	s.sessions[session.ID] = session

	return session
}

func (s *SessionStore) Get(id uuid.UUID) (domain.Session, bool) {

	s.mu.RLock()
	defer s.mu.RUnlock()

	session, ok := s.sessions[id]

	return session, ok
}

func (s *SessionStore) Touch(id uuid.UUID) bool {

	s.mu.Lock()
	defer s.mu.Unlock()

	session, ok := s.sessions[id]
	if !ok {
		return false
	}

	session.LastSeenAt = time.Now()

	s.sessions[id] = session

	return true
}