package domain

import (
	"time"

	"github.com/google/uuid"
)

type ScenarioRole = Role

type Version struct {
	ID          uuid.UUID
	Number      int
	MaxPoints   int
	PassPercent int
	PublishedAt time.Time
}

type AttemptResult struct {
	ID          uuid.UUID
	Version     Version
	Score       Score
	Passed      bool
	CompletedAt time.Time
}

type VersionProgress struct {
	Version         Version
	Completed       bool
	Passed          bool
	AttemptsCount   int
	BestScore       *Score
	ActiveAttemptID *uuid.UUID
}

type ScenarioProgress struct {
	ID             uuid.UUID
	Slug           string
	Title          string
	Role           ScenarioRole
	Current        VersionProgress
	History        []VersionProgress
	RecentAttempts []AttemptResult
}

type RoleProgress struct {
	Role               ScenarioRole
	TotalScenarios     int
	CompletedScenarios int
	PassedScenarios    int
	CompletionPercent  int
	PassedPercent      int
	Scenarios          []ScenarioProgress
}

type OutdatedActiveAttempt struct {
	AttemptID      uuid.UUID
	ScenarioSlug   string
	ScenarioTitle  string
	Version        Version
	CurrentVersion Version
}

type ProgressSnapshot struct {
	Scenarios              []ScenarioProgress
	OutdatedActiveAttempts []OutdatedActiveAttempt
}

type OverallProgress struct {
	TotalScenarios         int
	CompletedScenarios     int
	PassedScenarios        int
	CompletionPercent      int
	PassedPercent          int
	Roles                  []RoleProgress
	OutdatedActiveAttempts []OutdatedActiveAttempt
}
