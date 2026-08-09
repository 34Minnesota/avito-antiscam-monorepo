package httptransport

import (
	"time"

	"github.com/34Minnesota/avito-antiscam-monorepo/backend/internal/services/training"
	"github.com/google/uuid"
)

type UpdateMeRequest struct {
	Nickname string `json:"nickname" binding:"required,min=1,max=32"`
}

type StartAttemptRequest struct {
	ScenarioID uuid.UUID `json:"scenario_id" binding:"required"`
}

type ScenariosResponse struct {
	Scenarios []training.ScenarioCard `json:"scenarios"`
}

type ChoiceRequest struct {
	SceneID          string `json:"scene_id" binding:"required"`
	OptionID         string `json:"option_id" binding:"required"`
	ExpectedRevision *int   `json:"expected_revision" binding:"required,gte=0"`
}
type RegisterRequest struct {
	Nickname string `json:"nickname" binding:"required,min=3,max=30"`
	Email    string `json:"email" binding:"required,email,max=254"`
	Password string `json:"password" binding:"required,min=8,max=72"`
}

type RegisterResponse struct {
	SessionID uuid.UUID `json:"session_id"`
}

type LoginRequest struct {
	Email    string `json:"email" binding:"required,email,max=254"`
	Password string `json:"password" binding:"required,min=8,max=72"`
}

type LoginResponse struct {
	SessionID uuid.UUID `json:"session_id"`
}

type UserResponse struct {
	ID        uuid.UUID `json:"id"`
	Nickname  string    `json:"nickname"`
	Email     string    `json:"email"`
	CreatedAt time.Time `json:"created_at"`
}
