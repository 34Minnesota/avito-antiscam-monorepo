package domain

import (
	"time"

	"github.com/google/uuid"
)

type AttemptResult struct {
	ID          uuid.UUID
	Score       Score
	Outcome     Outcome
	CompletedAt time.Time
}

type ScenarioProgress struct {
	ID                       uuid.UUID
	Slug                     string
	Title                    string
	Role                     Role
	Completed                bool
	Passed                   bool
	AttemptsCount            int
	BestScore                *Score
	ActiveAttemptID          *uuid.UUID
	RecentAttempts           []AttemptResult
	InitialScore             *Score
	LatestScore              *Score
	ImprovementPercentPoints *int
	Trend                    *ProgressTrend
}

type ProgressTrend string

const (
	ProgressTrendImproving ProgressTrend = "improving"
	ProgressTrendStable    ProgressTrend = "stable"
	ProgressTrendDeclining ProgressTrend = "declining"
)

type RoleProgress struct {
	Role               Role
	TotalScenarios     int
	CompletedScenarios int
	PassedScenarios    int
	CompletionPercent  int
	PassedPercent      int
	Scenarios          []ScenarioProgress
}

type ProgressSnapshot struct {
	Scenarios []ScenarioProgress
}

type OverallProgress struct {
	TotalScenarios     int
	CompletedScenarios int
	PassedScenarios    int
	CompletionPercent  int
	PassedPercent      int
	Roles              []RoleProgress
}
