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
	// TODO(owner: training-engine): enforce positive step numbers, non-negative deltas
	// and client_action_id idempotency in the attempt engine transaction.
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
