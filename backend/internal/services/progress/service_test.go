package progress_service_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/34Minnesota/avito-antiscam-monorepo/backend/internal/domain"
	domainErrors "github.com/34Minnesota/avito-antiscam-monorepo/backend/internal/domain/errors"
	progress_service "github.com/34Minnesota/avito-antiscam-monorepo/backend/internal/services/progress"
)

type repositoryStub struct {
	snapshot domain.ProgressSnapshot
	err      error
	limit    int
}

func (r *repositoryStub) Load(_ context.Context, _ domain.UserID, limit int) (domain.ProgressSnapshot, error) {
	r.limit = limit
	return r.snapshot, r.err
}

func TestGetAggregatesCurrentProgress(t *testing.T) {
	t.Parallel()
	passedScore := mustScore(t, 80, 100)
	failedScore := mustScore(t, 60, 100)
	now := time.Now().UTC()
	repo := &repositoryStub{snapshot: domain.ProgressSnapshot{Scenarios: []domain.ScenarioProgress{
		{ID: uuid.New(), Slug: "buyer-one", Title: "Buyer", Role: domain.RoleBuyer,
			AttemptsCount: 3, Completed: true, Passed: true, BestScore: &passedScore,
			RecentAttempts: []domain.AttemptResult{{ID: uuid.New(), Score: passedScore, Outcome: domain.OutcomeSafe, CompletedAt: now}}},
		{ID: uuid.New(), Slug: "seller-one", Title: "Seller", Role: domain.RoleSeller,
			AttemptsCount: 2, Completed: true, Passed: false, BestScore: &failedScore},
	}}}
	result, err := progress_service.NewService(repo).Get(context.Background(), mustUserID(t))
	if err != nil {
		t.Fatal(err)
	}
	if repo.limit != progress_service.HistoryLimit {
		t.Fatalf("limit = %d", repo.limit)
	}
	if result.TotalScenarios != 2 || result.CompletedScenarios != 2 || result.PassedScenarios != 1 {
		t.Fatalf("unexpected totals: %+v", result)
	}
	if result.CompletionPercent != 100 || result.PassedPercent != 50 {
		t.Fatalf("unexpected percentages: %+v", result)
	}
	if len(result.Roles) != 2 || result.Roles[0].Role != domain.RoleBuyer || result.Roles[1].Role != domain.RoleSeller {
		t.Fatalf("unexpected roles: %+v", result.Roles)
	}
}

func TestGetEmptyCatalogReturnsZeroPercentages(t *testing.T) {
	t.Parallel()
	result, err := progress_service.NewService(&repositoryStub{}).Get(context.Background(), mustUserID(t))
	if err != nil {
		t.Fatal(err)
	}
	if result.TotalScenarios != 0 || result.CompletionPercent != 0 || result.PassedPercent != 0 || len(result.Roles) != 2 {
		t.Fatalf("unexpected empty progress: %+v", result)
	}
}

func TestGetCalculatesScenarioDynamics(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name          string
		attemptsCount int
		initial       *domain.Score
		latest        *domain.Score
		wantDelta     *int
		wantTrend     *domain.ProgressTrend
	}{
		{
			name: "no completed attempts",
		},
		{
			name:          "one attempt",
			attemptsCount: 1,
			initial:       scorePointer(mustScore(t, 42, 100)),
			latest:        scorePointer(mustScore(t, 42, 100)),
			wantDelta:     intPointer(0),
			wantTrend:     trendPointer(domain.ProgressTrendStable),
		},
		{
			name:          "improving",
			attemptsCount: 2,
			initial:       scorePointer(mustScore(t, 42, 100)),
			latest:        scorePointer(mustScore(t, 78, 100)),
			wantDelta:     intPointer(36),
			wantTrend:     trendPointer(domain.ProgressTrendImproving),
		},
		{
			name:          "declining",
			attemptsCount: 2,
			initial:       scorePointer(mustScore(t, 78, 100)),
			latest:        scorePointer(mustScore(t, 42, 100)),
			wantDelta:     intPointer(-36),
			wantTrend:     trendPointer(domain.ProgressTrendDeclining),
		},
		{
			name:          "stable",
			attemptsCount: 2,
			initial:       scorePointer(mustScore(t, 60, 100)),
			latest:        scorePointer(mustScore(t, 60, 100)),
			wantDelta:     intPointer(0),
			wantTrend:     trendPointer(domain.ProgressTrendStable),
		},
	}

	for _, tt := range tests {

		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			repo := &repositoryStub{snapshot: domain.ProgressSnapshot{Scenarios: []domain.ScenarioProgress{{
				ID: uuid.New(), Slug: "scenario", Title: "Scenario", Role: domain.RoleBuyer,
				AttemptsCount: tt.attemptsCount,
				InitialScore:  tt.initial, LatestScore: tt.latest,
			}}}}

			result, err := progress_service.NewService(repo).Get(context.Background(), mustUserID(t))
			if err != nil {
				t.Fatal(err)
			}
			scenario := result.Roles[0].Scenarios[0]
			if !equalIntPointers(scenario.ImprovementPercentPoints, tt.wantDelta) {
				t.Fatalf("improvement = %v, want %v", scenario.ImprovementPercentPoints, tt.wantDelta)
			}
			if !equalTrendPointers(scenario.Trend, tt.wantTrend) {
				t.Fatalf("trend = %v, want %v", scenario.Trend, tt.wantTrend)
			}
		})
	}
}

func TestGetRejectsInvalidSnapshotAndMissingDependency(t *testing.T) {
	t.Parallel()
	_, err := progress_service.NewService(nil).Get(context.Background(), mustUserID(t))
	if !errors.Is(err, domainErrors.ErrDependencyUnavailable) {
		t.Fatalf("got %v", err)
	}
	invalid := domain.ScenarioProgress{ID: uuid.New(), Slug: "broken", Title: "Broken", Role: domain.RoleBuyer, Passed: true}
	_, err = progress_service.NewService(&repositoryStub{snapshot: domain.ProgressSnapshot{Scenarios: []domain.ScenarioProgress{invalid}}}).Get(context.Background(), mustUserID(t))
	if !errors.Is(err, domainErrors.ErrDataInconsistent) {
		t.Fatalf("got %v", err)
	}
}

func TestGetRejectsOversizedOrUnorderedHistory(t *testing.T) {
	t.Parallel()
	now := time.Now().UTC()
	score := mustScore(t, 80, 100)
	attempts := make([]domain.AttemptResult, progress_service.HistoryLimit+1)
	for index := range attempts {
		attempts[index] = domain.AttemptResult{ID: uuid.New(), Score: score, Outcome: domain.OutcomeSafe, CompletedAt: now.Add(-time.Duration(index) * time.Minute)}
	}
	snapshot := domain.ProgressSnapshot{Scenarios: []domain.ScenarioProgress{{
		ID: uuid.New(), Slug: "scenario", Title: "Scenario", Role: domain.RoleBuyer,
		RecentAttempts: attempts,
	}}}
	_, err := progress_service.NewService(&repositoryStub{snapshot: snapshot}).Get(context.Background(), mustUserID(t))
	if !errors.Is(err, domainErrors.ErrDataInconsistent) {
		t.Fatalf("got %v", err)
	}
}

func mustScore(t *testing.T, points, max int) domain.Score {
	t.Helper()
	score, err := domain.NewScore(points, max)
	if err != nil {
		t.Fatal(err)
	}
	return score
}

func mustUserID(t *testing.T) domain.UserID {
	t.Helper()
	id, err := domain.NewUserID(uuid.New())
	if err != nil {
		t.Fatal(err)
	}
	return id
}

func scorePointer(score domain.Score) *domain.Score { return &score }

func intPointer(value int) *int { return &value }

func trendPointer(trend domain.ProgressTrend) *domain.ProgressTrend { return &trend }

func equalIntPointers(actual, expected *int) bool {
	return actual == nil && expected == nil || actual != nil && expected != nil && *actual == *expected
}

func equalTrendPointers(actual, expected *domain.ProgressTrend) bool {
	return actual == nil && expected == nil || actual != nil && expected != nil && *actual == *expected
}
