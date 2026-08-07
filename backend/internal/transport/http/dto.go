package httptransport

import (
	"github.com/google/uuid"
)

type UpdateMeRequest struct {
	Nickname string `json:"nickname" binding:"required,min=1,max=32"`
}

type StartAttemptRequest struct {
	ScenarioID uuid.UUID `json:"scenario_id" binding:"required"`
}

type ChoiceRequest struct {
	SceneID          string `json:"scene_id" binding:"required"`
	OptionID         string `json:"option_id" binding:"required"`
	ExpectedRevision *int   `json:"expected_revision" binding:"required,gte=0"`
}
