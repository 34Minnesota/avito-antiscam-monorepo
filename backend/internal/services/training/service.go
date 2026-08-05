package training

import (
	"context"
	"time"

	"github.com/34Minnesota/avito-antiscam-monorepo/backend/internal/domain"
	"github.com/google/uuid"
)

type Repository interface {
	UpsertScenario(ctx context.Context, s domain.Scenario) error
	ListScenarios(ctx context.Context, role domain.Role) ([]domain.Scenario, error)
	ScenarioByID(ctx context.Context, id uuid.UUID) (domain.Scenario, error)
	ScenarioStats(ctx context.Context, userID uuid.UUID) (map[uuid.UUID]ScenarioStats, error)

	CreateAttempt(ctx context.Context, a domain.Attempt) error
	AttemptByID(ctx context.Context, id uuid.UUID) (domain.Attempt, error)
	ActiveAttempt(ctx context.Context, userID, scenarioID uuid.UUID) (domain.Attempt, error)
	SaveStep(ctx context.Context, step domain.AttemptStep, a domain.Attempt) error
	FinishAttempt(ctx context.Context, id uuid.UUID, score int, outcome domain.Outcome, at time.Time) error
	Steps(ctx context.Context, attemptID uuid.UUID) ([]domain.AttemptStep, error)
	BestScoreExcept(ctx context.Context, userID, scenarioID, exceptAttempt uuid.UUID) (*int, error)
}
