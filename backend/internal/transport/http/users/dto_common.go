package users_http_transport

import (
	"time"

	"github.com/google/uuid"
)

type UserDTOResponse struct {
	ID        uuid.UUID `json:"id"`
	Nickname  string    `json:"nickname"`
	Email     string    `json:"email"`
	CreatedAt time.Time `json:"created_at"`
}
