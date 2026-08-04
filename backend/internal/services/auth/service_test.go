package auth

import (
	"testing"

	"github.com/google/uuid"
)

func TestCreateSession(t *testing.T) {

	store := NewSessionStore()

	service := NewService(store)

	session := service.CreateSession()

	if session.ID == uuid.Nil {
		t.Fatal("session id should not be nil")
	}

	if _, ok := store.Get(session.ID); !ok {
		t.Fatal("session should be stored")
	}
}

func TestValidateSessionSuccess(t *testing.T) {

	store := NewSessionStore()

	service := NewService(store)

	session := service.CreateSession()

	result, ok := service.ValidateSession(session.ID)

	if !ok {
		t.Fatal("expected valid session")
	}

	if result.ID != session.ID {
		t.Fatal("wrong session returned")
	}
}

func TestValidateSessionFail(t *testing.T) {

	store := NewSessionStore()

	service := NewService(store)

	_, ok := service.ValidateSession(uuid.New())

	if ok {
		t.Fatal("expected validation to fail")
	}
}