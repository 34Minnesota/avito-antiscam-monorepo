package domain

import (
	"time"

	"github.com/google/uuid"
)

type AttemptStatus string

const (
	StatusInProgress AttemptStatus = "in_progress"
	StatusCompleted  AttemptStatus = "completed"
)

func (s AttemptStatus) IsValid() bool {
	switch s {
	case StatusCompleted, StatusInProgress:
		return true
	default:
		return false
	}
}

type Attempt struct {
	ID                uuid.UUID
	UserID            uuid.UUID
	ScenarioVersionID uuid.UUID
	CurrentNodeID     uuid.UUID
	Status            AttemptStatus
	ScorePoints       int
	MaxScorePoints    int
	Revision          int
	StartedAt         *time.Time
	CompletedAt       *time.Time
}

func NewAttempt(
	id, userID, scenarioVersionID, currentNodeID uuid.UUID,
	status AttemptStatus,
	scorePoints, maxScorePoints, revision int,
	startedAt, completedAt *time.Time,
) Attempt {
	return Attempt{
		ID:                id,
		UserID:            userID,
		ScenarioVersionID: scenarioVersionID,
		CurrentNodeID:     currentNodeID,
		Status:            status,
		ScorePoints:       scorePoints,
		MaxScorePoints:    maxScorePoints,
		Revision:          revision,
		StartedAt:         startedAt,
		CompletedAt:       completedAt,
	}
}
