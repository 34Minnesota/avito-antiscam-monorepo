package progress

import (
	"context"
	"errors"
	"fmt"
	"sort"

	"github.com/google/uuid"

	"github.com/34Minnesota/avito-antiscam-monorepo/backend/internal/domain"
	domainErrors "github.com/34Minnesota/avito-antiscam-monorepo/backend/internal/domain/errors"
)

const HistoryLimit = 20

var (
	ErrDependencyUnavailable = errors.New("progress dependency unavailable")
	ErrDataInconsistent      = errors.New("progress data is inconsistent")
)

type Repository interface {
	Load(context.Context, domain.UserID, int) (domain.ProgressSnapshot, error)
}

type Service struct{ repository Repository }

func New(repository Repository) Service {
	return Service{repository: repository}
}

func (s Service) Get(ctx context.Context, userID domain.UserID) (domain.OverallProgress, error) {
	if userID.IsZero() {
		return domain.OverallProgress{}, domainErrors.ErrInvalidUserID
	}

	if s.repository == nil {
		return domain.OverallProgress{}, fmt.Errorf("%w: repository is not configured", ErrDependencyUnavailable)
	}

	snapshot, err := s.repository.Load(ctx, userID, HistoryLimit)
	if err != nil {
		return domain.OverallProgress{}, err
	}

	if err := validate(snapshot); err != nil {
		return domain.OverallProgress{}, fmt.Errorf("%w: %w", ErrDataInconsistent, err)
	}

	return aggregate(snapshot), nil
}

func aggregate(snapshot domain.ProgressSnapshot) domain.OverallProgress {
	roles := map[domain.Role]*domain.RoleProgress{
		domain.RoleBuyer:  {Role: domain.RoleBuyer},
		domain.RoleSeller: {Role: domain.RoleSeller},
	}

	overall := domain.OverallProgress{OutdatedActiveAttempts: snapshot.OutdatedActiveAttempts}

	for _, scenario := range snapshot.Scenarios {
		role := roles[scenario.Role]

		role.Scenarios = append(role.Scenarios, scenario)

		role.TotalScenarios++
		overall.TotalScenarios++

		if scenario.Current.Completed {
			role.CompletedScenarios++
			overall.CompletedScenarios++
		}

		if scenario.Current.Passed {
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
	if scenario.ID == uuid.Nil || scenario.Slug == "" || scenario.Title == "" || (scenario.Role != domain.RoleBuyer && scenario.Role != domain.RoleSeller) {
		return domainErrors.ErrInvalidProgressData
	}

	if err := validateVersionProgress(scenario.Current); err != nil {
		return err
	}

	for _, version := range scenario.History {
		if err := validateVersionProgress(version); err != nil {
			return err
		}
	}

	if len(scenario.RecentAttempts) > HistoryLimit {
		return domainErrors.ErrInvalidProgressData
	}

	return validateAttempts(scenario.RecentAttempts)
}

func validateAttempts(attempts []domain.AttemptResult) error {
	for index, attempt := range attempts {
		if attempt.ID == uuid.Nil || attempt.CompletedAt.IsZero() || attempt.Score.MaxPoints() != attempt.Version.MaxPoints {
			return domainErrors.ErrInvalidProgressData
		}

		if index > 0 && attempt.CompletedAt.After(attempts[index-1].CompletedAt) {
			return domainErrors.ErrInvalidProgressData
		}
	}

	return nil
}

func validateVersionProgress(progress domain.VersionProgress) error {
	v := progress.Version

	if v.ID == uuid.Nil || v.Number < 1 || v.MaxPoints < 1 || v.PassPercent < 1 || v.PassPercent > 100 || v.PublishedAt.IsZero() {
		return domainErrors.ErrInvalidProgressData
	}

	if progress.AttemptsCount < 0 || progress.Passed && !progress.Completed || progress.BestScore != nil && !progress.Completed {
		return domainErrors.ErrInvalidProgressData
	}

	if progress.BestScore != nil && progress.BestScore.MaxPoints() != v.MaxPoints {
		return domainErrors.ErrInvalidProgressData
	}

	return nil
}
