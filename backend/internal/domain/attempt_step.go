package domain

import (
	"time"

	"github.com/google/uuid"
)

type AttemptStep struct {
	ID             uuid.UUID
	AttemptID      uuid.UUID
	StepNo         int
	TransitionID   uuid.UUID
	ClientActionID uuid.UUID
	ScoreDelta     int
	CreatedAt      time.Time
}

func NewAttemptStep(
	id, attemptID uuid.UUID,
	stepNo int,
	transitionID, clientActionID uuid.UUID,
	scoreDelta int,
	createdAt time.Time,
) AttemptStep {
	return AttemptStep{
		ID:             id,
		AttemptID:      attemptID,
		StepNo:         stepNo,
		TransitionID:   transitionID,
		ClientActionID: clientActionID,
		ScoreDelta:     scoreDelta,
		CreatedAt:      createdAt,
	}
}
