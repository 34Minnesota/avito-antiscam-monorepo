package domain

import (
	"time"

	"github.com/google/uuid"
)

type ScenarioVersion struct {
	ID             uuid.UUID
	ScenarioID     uuid.UUID
	Version        int
	StartNodeID    uuid.UUID
	MaxScorePoints int
	PublishedAt    time.Time
}

func NewScenarioVersion(
	id, scenarioID uuid.UUID,
	version int,
	startNodeID uuid.UUID,
	maxScorePoints int,
	publishedAt time.Time,
) ScenarioVersion {
	return ScenarioVersion{
		ID:             id,
		ScenarioID:     scenarioID,
		Version:        version,
		StartNodeID:    startNodeID,
		MaxScorePoints: maxScorePoints,
		PublishedAt:    publishedAt,
	}
}
