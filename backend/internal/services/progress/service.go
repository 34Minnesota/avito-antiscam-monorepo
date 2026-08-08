package progress_service

import (
	"context"
	"fmt"
	"sort"

	"github.com/google/uuid"

	"github.com/34Minnesota/avito-antiscam-monorepo/backend/internal/domain"
	domainErrors "github.com/34Minnesota/avito-antiscam-monorepo/backend/internal/domain/errors"
)

const HistoryLimit = 20

type Repository interface {
	Load(context.Context, domain.UserID, int) (domain.ProgressSnapshot, error)
}

type Service struct{ repository Repository }

func NewService(repository Repository) Service {
	return Service{repository: repository}
}

func (s Service) Get(ctx context.Context, userID domain.UserID) (domain.OverallProgress, error) {
	if userID.IsZero() {
		return domain.OverallProgress{}, domainErrors.ErrInvalidUserID
	}

	if s.repository == nil {
		return domain.OverallProgress{}, fmt.Errorf("%w: repository is not configured", domainErrors.ErrDependencyUnavailable)
	}

	snapshot, err := s.repository.Load(ctx, userID, HistoryLimit)
	if err != nil {
		return domain.OverallProgress{}, err
	}

	if err := validate(snapshot); err != nil {
		return domain.OverallProgress{}, fmt.Errorf("%w: %w", domainErrors.ErrDataInconsistent, err)
	}

	return aggregate(snapshot), nil
}

func aggregate(snapshot domain.ProgressSnapshot) domain.OverallProgress {
	roles := map[domain.Role]*domain.RoleProgress{
		domain.RoleBuyer:  {Role: domain.RoleBuyer},
		domain.RoleSeller: {Role: domain.RoleSeller},
	}

	overall := domain.OverallProgress{}

	for _, scenario := range snapshot.Scenarios {
		if scenario.InitialScore != nil && scenario.LatestScore != nil {
			initialPercent := scenario.InitialScore.Percent()
			latestPercent := scenario.LatestScore.Percent()

			improvement := latestPercent - initialPercent
			var trend domain.ProgressTrend

			if latestPercent > initialPercent {
				trend = domain.ProgressTrendImproving
			} else if latestPercent < initialPercent {
				trend = domain.ProgressTrendDeclining
			} else {
				trend = domain.ProgressTrendStable
			}

			scenario.ImprovementPercentPoints = &improvement
			scenario.Trend = &trend
		}

		role := roles[scenario.Role]

		role.Scenarios = append(role.Scenarios, scenario)

		role.TotalScenarios++
		overall.TotalScenarios++

		if scenario.Completed {
			role.CompletedScenarios++
			overall.CompletedScenarios++
		}

		if scenario.Passed {
			role.PassedScenarios++
			overall.PassedScenarios++
		}
	}

	overall.CompletionPercent = percentage(overall.CompletedScenarios, overall.TotalScenarios)
	overall.PassedPercent = percentage(overall.PassedScenarios, overall.TotalScenarios)

	for _, roleName := range []domain.Role{domain.RoleBuyer, domain.RoleSeller} {
		role := roles[roleName]
		role.CompletionPercent = percentage(role.CompletedScenarios, role.TotalScenarios)
		role.PassedPercent = percentage(role.PassedScenarios, role.TotalScenarios)
		sort.Slice(role.Scenarios, func(i, j int) bool { return role.Scenarios[i].Slug < role.Scenarios[j].Slug })
		overall.Roles = append(overall.Roles, *role)
	}

	return overall
}

func percentage(value, total int) int {
	if total == 0 {
		return 0
	}

	return int((int64(value)*100 + int64(total)/2) / int64(total))
}

func validate(snapshot domain.ProgressSnapshot) error {
	for _, scenario := range snapshot.Scenarios {
		if err := validateScenario(scenario); err != nil {
			return err
		}
	}

	return nil
}

func validateScenario(scenario domain.ScenarioProgress) error {
	if !isValidScenarioIdentity(scenario) {
		return domainErrors.ErrInvalidProgressData
	}

	if !isValidScenarioState(scenario) {
		return domainErrors.ErrInvalidProgressData
	}

	if !isValidScenarioScores(scenario) {
		return domainErrors.ErrInvalidProgressData
	}

	if len(scenario.RecentAttempts) > HistoryLimit {
		return domainErrors.ErrInvalidProgressData
	}

	if err := validateInitialLatestScore(scenario.InitialScore, scenario.LatestScore); err != nil {
		return err
	}

	return validateAttempts(scenario.RecentAttempts)
}

func isValidScenarioIdentity(scenario domain.ScenarioProgress) bool {
	return scenario.ID != uuid.Nil &&
		scenario.Slug != "" &&
		scenario.Title != "" &&
		isValidRole(scenario.Role)
}

func isValidRole(role domain.Role) bool {
	return role == domain.RoleBuyer || role == domain.RoleSeller
}

func isValidScenarioState(scenario domain.ScenarioProgress) bool {
	if scenario.AttemptsCount < 0 {
		return false
	}

	if scenario.Passed && !scenario.Completed {
		return false
	}

	return scenario.BestScore == nil || scenario.Completed
}

func isValidScenarioScores(scenario domain.ScenarioProgress) bool {
	if scenario.BestScore != nil && scenario.BestScore.MaxPoints() != 100 {
		return false
	}

	return scenario.AttemptsCount != 0 ||
		(scenario.InitialScore == nil && scenario.LatestScore == nil)
}

func validateInitialLatestScore(initialScore, latestScore *domain.Score) error {
	if (initialScore == nil && latestScore != nil) ||
		(initialScore != nil && latestScore == nil) {
		return domainErrors.ErrInvalidProgressData
	}

	if (initialScore != nil && initialScore.MaxPoints() != 100) ||
		(latestScore != nil && latestScore.MaxPoints() != 100) {
		return domainErrors.ErrInvalidProgressData
	}

	return nil
}

func validateAttempts(attempts []domain.AttemptResult) error {
	for index, attempt := range attempts {
		if attempt.ID == uuid.Nil || attempt.CompletedAt.IsZero() || attempt.Score.MaxPoints() != 100 || !validOutcome(attempt.Outcome) {
			return domainErrors.ErrInvalidProgressData
		}

		if index > 0 && attempt.CompletedAt.After(attempts[index-1].CompletedAt) {
			return domainErrors.ErrInvalidProgressData
		}
	}

	return nil
}

func validOutcome(outcome domain.Outcome) bool {
	return outcome == domain.OutcomeSafe || outcome == domain.OutcomePartial || outcome == domain.OutcomeScammed
}
