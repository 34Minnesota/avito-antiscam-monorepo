package domain

import (
	"time"

	"github.com/google/uuid"
)

type Session struct {
	ID         uuid.UUID
	CreatedAt  time.Time
	LastSeenAt time.Time
}
