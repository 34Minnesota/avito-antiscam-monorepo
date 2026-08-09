package progress_service

import (
	"cmp"
	"context"
	"fmt"
	"slices"
	"sort"

	"github.com/google/uuid"

	"github.com/34Minnesota/avito-antiscam-monorepo/backend/internal/domain"
	domainErrors "github.com/34Minnesota/avito-antiscam-monorepo/backend/internal/domain/errors"
)

const HistoryLimit = 20

const recommendationLimit = 3

const experienceLevelSize = 100

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

	buyer := roles[domain.RoleBuyer]
	seller := roles[domain.RoleSeller]

	completionPercentDelta := buyer.CompletionPercent - seller.CompletionPercent
	passedPercentDelta := buyer.PassedPercent - seller.PassedPercent

	overall.RoleComparison = domain.RoleComparison{
		CompletionPercentDelta: completionPercentDelta,
		PassedPercentDelta:     passedPercentDelta,
	}
	overall.Recommendations = recommendations(snapshot.Scenarios)
	overall.Experience = experience(snapshot.Scenarios)

	return overall
}

func experience(scenarios []domain.ScenarioProgress) domain.ExperienceProgress {
	completedScenarios := 0
	passedScenarios := 0
	improvedScenarios := 0
	perfectScenarios := 0
	passedRoles := make(map[domain.Role]struct{})

	for _, scenario := range scenarios {
		if scenario.Completed {
			completedScenarios++
		}

		if scenario.Passed {
			passedScenarios++
			passedRoles[scenario.Role] = struct{}{}
		}

		if scenario.InitialScore != nil &&
			scenario.LatestScore != nil &&
			scenario.LatestScore.Percent() > scenario.InitialScore.Percent() {
			improvedScenarios++
		}

		if scenario.Passed &&
			scenario.BestScore != nil &&
			scenario.BestScore.Percent() == 100 {
			perfectScenarios++
		}
	}

	totalXP := completedScenarios*10 + passedScenarios*15 + improvedScenarios*10 + perfectScenarios*15
	return domain.ExperienceProgress{
		TotalXP:     totalXP,
		Level:       totalXP/experienceLevelSize + 1,
		CurrentXP:   totalXP % experienceLevelSize,
		NextLevelXP: experienceLevelSize,
		Achievements: achievements(
			completedScenarios,
			passedScenarios,
			improvedScenarios,
			len(passedRoles) == 2,
			perfectScenarios,
		),
	}
}

func achievements(
	completedScenarios, passedScenarios, improvedScenarios int,
	bothRolesPassed bool,
	perfectScenarios int,
) []domain.Achievement {
	return []domain.Achievement{
		{
			Code:        "FIRST_COMPLETION",
			Title:       "Первый шаг",
			Description: "Завершён первый сценарий.",
			Earned:      completedScenarios > 0,
		},
		{
			Code:        "FIRST_SAFE_RESULT",
			Title:       "Безопасный исход",
			Description: "Получен первый безопасный результат.",
			Earned:      passedScenarios > 0,
		},
		{
			Code:        "IMPROVEMENT",
			Title:       "Работа над ошибками",
			Description: "Результат по сценарию улучшен.",
			Earned:      improvedScenarios > 0,
		},
		{
			Code:        "BOTH_ROLES",
			Title:       "Две роли",
			Description: "Есть успешно пройденные сценарии покупателя и продавца.",
			Earned:      bothRolesPassed,
		},
		{
			Code:        "PERFECT_SCORE",
			Title:       "Безупречный результат",
			Description: "Лучший score равен 100.",
			Earned:      perfectScenarios > 0,
		},
	}
}

type recommendationCandidate struct {
	recommendation domain.Recommendation
	priority       int
	bestScore      int
}

func recommendations(scenarios []domain.ScenarioProgress) []domain.Recommendation {
	candidates := make([]recommendationCandidate, 0, len(scenarios))

	for _, scenario := range scenarios {
		candidate, ok := recommendationFor(scenario)
		if ok {
			candidates = append(candidates, candidate)
		}
	}

	slices.SortFunc(candidates, func(a, b recommendationCandidate) int {
		if n := cmp.Compare(a.priority, b.priority); n != 0 {
			return n
		}

		if n := cmp.Compare(a.bestScore, b.bestScore); n != 0 {
			return n
		}

		return cmp.Compare(
			a.recommendation.ScenarioSlug,
			b.recommendation.ScenarioSlug,
		)
	})

	if len(candidates) > recommendationLimit {
		candidates = candidates[:recommendationLimit]
	}

	result := make([]domain.Recommendation, 0, len(candidates))
	for _, candidate := range candidates {
		result = append(result, candidate.recommendation)
	}

	return result
}

func recommendationFor(scenario domain.ScenarioProgress) (recommendationCandidate, bool) {
	if scenario.ActiveAttemptID != nil {
		return newRecommendationCandidate(scenario, "ACTIVE_ATTEMPT", "У вас есть незавершённая попытка. Продолжите сценарий.", 0), true
	}

	if !scenario.Passed && scenario.AttemptsCount > 0 && scenario.BestScore == nil {
		return newRecommendationCandidate(scenario, "NOT_PASSED", "Сценарий пока не пройден безопасно. Попробуйте ещё раз.", 1), true
	}

	if scenario.Completed && !scenario.Passed && scenario.BestScore != nil {
		return newRecommendationCandidate(scenario, "LOW_BEST_SCORE", fmt.Sprintf("Лучший результат по сценарию — %d%%. Попробуйте улучшить его.", scenario.BestScore.Percent()), 2), true
	}

	if scenario.AttemptsCount == 0 {
		return newRecommendationCandidate(scenario, "NOT_STARTED", "Сценарий ещё не пройден. Начните его.", 3), true
	}

	if scenario.Passed {
		return newRecommendationCandidate(scenario, "REPEAT_FOR_REINFORCEMENT", "Сценарий пройден безопасно. Повторите его, чтобы закрепить навык.", 4), true
	}

	return recommendationCandidate{}, false
}

func newRecommendationCandidate(scenario domain.ScenarioProgress, reasonCode, reasonText string, priority int) recommendationCandidate {
	bestScore := 101 // 101 чтобы сценарий отсортировался после тех, у которых это поле задано
	if scenario.BestScore != nil {
		bestScore = scenario.BestScore.Percent()
	}

	return recommendationCandidate{
		recommendation: domain.Recommendation{
			ScenarioSlug: scenario.Slug,
			ReasonCode:   reasonCode,
			ReasonText:   reasonText,
		},
		priority:  priority,
		bestScore: bestScore,
	}
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
