package training

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/34Minnesota/avito-antiscam-monorepo/backend/internal/domain"
	domainErrors "github.com/34Minnesota/avito-antiscam-monorepo/backend/internal/domain/errors"
	"github.com/google/uuid"
)

type Repository interface {
	UpsertScenario(ctx context.Context, s domain.Scenario) error
	ListScenarios(ctx context.Context, role domain.Role) ([]domain.Scenario, error)
	ScenarioByID(ctx context.Context, id uuid.UUID) (domain.Scenario, error)
	ScenarioStats(ctx context.Context, userID domain.UserID) (map[uuid.UUID]ScenarioStats, error)

	CreateAttempt(ctx context.Context, a domain.Attempt) error
	AttemptByID(ctx context.Context, id uuid.UUID) (domain.Attempt, error)
	ActiveAttempt(ctx context.Context, userID domain.UserID, scenarioID uuid.UUID) (domain.Attempt, error)
	SaveStep(ctx context.Context, step domain.AttemptStep, a domain.Attempt, userID domain.UserID, expectedRevision int) (int, error)
	FinishAttempt(ctx context.Context, id uuid.UUID, score int, outcome domain.Outcome, at time.Time) error
	Steps(ctx context.Context, attemptID uuid.UUID) ([]domain.AttemptStep, error)
	BestScore(ctx context.Context, userID domain.UserID, scenarioID, exclude uuid.UUID) (*int, error)
}

type Service struct {
	repo Repository
}

func New(repo Repository) *Service {
	return &Service{repo: repo}
}

// Catalog возвращает список сценариев с персональной статистикой пользователя.
func (s *Service) Catalog(ctx context.Context, userID domain.UserID, role domain.Role) ([]ScenarioCard, error) {
	scenarios, err := s.repo.ListScenarios(ctx, role)
	if err != nil {
		return nil, err
	}

	stats, err := s.repo.ScenarioStats(ctx, userID)
	if err != nil {
		return nil, err
	}

	cards := make([]ScenarioCard, 0, len(scenarios))
	for _, sc := range scenarios {
		card := ScenarioCard{
			ID:          sc.ID,
			Slug:        sc.Doc.Slug,
			Role:        sc.Doc.Role,
			Category:    sc.Doc.Category,
			Difficulty:  sc.Doc.Difficulty,
			Title:       sc.Doc.Title,
			Description: sc.Doc.Description,
		}
		if stat, ok := stats[sc.ID]; ok {
			card.Stats = &stat
		}
		cards = append(cards, card)
	}
	return cards, nil
}

// Start начинает новое прохождение сценария.
//
// Если у пользователя уже есть незавершённая попытка по этому сценарию,
// она переиспользуется: так перезагрузка страницы не создаёт мусорных
// попыток и не сбрасывает прогресс. Позиция восстанавливается из state.
func (s *Service) Start(ctx context.Context, userID domain.UserID, scenarioID uuid.UUID) (StartResult, error) {
	sc, err := s.repo.ScenarioByID(ctx, scenarioID)
	if err != nil {
		return StartResult{}, err
	}

	existing, err := s.repo.ActiveAttempt(ctx, userID, scenarioID)
	switch {
	case err == nil:
		scene, ok := sc.Doc.SceneByID(CurrentSceneID(sc.Doc, existing.State))
		if !ok {
			return StartResult{}, fmt.Errorf(
				"%w: active attempt %s points outside scenario scenes",
				domainErrors.ErrInvalidScenario, existing.ID)
		}
		return buildStartResult(existing.ID, sc, BuildScene(scene)), nil

	case errors.Is(err, domainErrors.ErrNotFound):

	default:
		return StartResult{}, err
	}

	scene, state, err := Start(sc.Doc)
	if err != nil {
		return StartResult{}, err
	}

	attempt := domain.Attempt{
		ID:         uuid.New(),
		UserID:     userID,
		ScenarioID: scenarioID,
		Status:     domain.AttemptInProgress,
		State:      state,
		StartedAt:  time.Now().UTC(),
	}
	if err := s.repo.CreateAttempt(ctx, attempt); err != nil {
		return StartResult{}, err
	}

	return buildStartResult(attempt.ID, sc, scene), nil
}

func buildStartResult(attemptID uuid.UUID, sc domain.Scenario, scene ScenePayload) StartResult {
	return StartResult{
		AttemptID:   attemptID,
		Listing:     sc.Doc.Listing,
		Counterpart: sc.Doc.Counterpart,
		Role:        sc.Doc.Role,
		Title:       sc.Doc.Title,
		Scene:       scene,
		ScenesTotal: len(sc.Doc.Scenes),
	}
}

// Choose обрабатывает выбор пользователя на развилке.
func (s *Service) Choose(
	ctx context.Context,
	userID domain.UserID,
	attemptID uuid.UUID,
	sceneID, optionID string,
	expectedRevision int,
) (ChoiceResult, error) {
	attempt, sc, err := s.loadAttempt(ctx, userID, attemptID)
	if err != nil {
		return ChoiceResult{}, err
	}
	if attempt.Status != domain.AttemptInProgress {
		return ChoiceResult{}, domainErrors.ErrAttemptFinished
	}

	tr, err := Advance(sc.Doc, attempt.State, sceneID, optionID)
	if err != nil {
		return ChoiceResult{}, err
	}

	step := tr.Step
	step.AttemptID = attempt.ID
	step.CreatedAt = time.Now().UTC()

	attempt.State = tr.State
	if tr.Finished {
		attempt.Status = domain.AttemptFinished
	}

	newRevision, err := s.repo.SaveStep(
		ctx,
		step,
		attempt,
		userID,
		expectedRevision,
	)
	if err != nil {
		return ChoiceResult{}, err
	}

	result := ChoiceResult{
		Feedback:  Feedback{Verdict: tr.Verdict, Text: tr.Feedback},
		Reaction:  tr.Reaction,
		NextScene: tr.NextScene,
		Finished:  tr.Finished,
		Revision:  newRevision,
	}
	if !tr.Finished {
		return result, nil
	}

	summary, err := s.finish(ctx, attempt, sc)
	if err != nil {
		return ChoiceResult{}, err
	}
	result.Summary = &summary
	return result, nil
}

// Summary возвращает итог завершённой попытки: балл, концовку и разбор.
func (s *Service) Summary(ctx context.Context, userID domain.UserID, attemptID uuid.UUID) (SummaryResult, error) {
	attempt, sc, err := s.loadAttempt(ctx, userID, attemptID)
	if err != nil {
		return SummaryResult{}, err
	}

	steps, err := s.repo.Steps(ctx, attemptID)
	if err != nil {
		return SummaryResult{}, err
	}

	result, err := Evaluate(sc.Doc, steps)
	if err != nil {
		return SummaryResult{}, err
	}
	if err := s.attachDelta(ctx, &result, attempt); err != nil {
		return SummaryResult{}, err
	}
	return result, nil
}

// finish считает итог, сохраняет его и возвращает разбор.
func (s *Service) finish(ctx context.Context, attempt domain.Attempt, sc domain.Scenario) (SummaryResult, error) {
	steps, err := s.repo.Steps(ctx, attempt.ID)
	if err != nil {
		return SummaryResult{}, err
	}

	result, err := Evaluate(sc.Doc, steps)
	if err != nil {
		return SummaryResult{}, err
	}
	if err := s.attachDelta(ctx, &result, attempt); err != nil {
		return SummaryResult{}, err
	}
	if err := s.repo.FinishAttempt(ctx, attempt.ID, result.Score, result.Outcome, time.Now().UTC()); err != nil {
		return SummaryResult{}, err
	}
	return result, nil
}

// attachDelta добавляет к результату разницу с лучшей предыдущей попыткой
func (s *Service) attachDelta(ctx context.Context, result *SummaryResult, attempt domain.Attempt) error {
	prev, err := s.repo.BestScore(ctx, attempt.UserID, attempt.ScenarioID, attempt.ID)
	if err != nil {
		return err
	}
	if prev != nil {
		delta := result.Score - *prev
		result.DeltaVsPrevious = &delta
	}
	return nil
}

func (s *Service) loadAttempt(ctx context.Context, userID domain.UserID, attemptID uuid.UUID) (domain.Attempt, domain.Scenario, error) {
	attempt, err := s.repo.AttemptByID(ctx, attemptID)
	if err != nil {
		return domain.Attempt{}, domain.Scenario{}, err
	}
	if attempt.UserID != userID {
		return domain.Attempt{}, domain.Scenario{}, domainErrors.ErrForbidden
	}

	sc, err := s.repo.ScenarioByID(ctx, attempt.ScenarioID)
	if err != nil {
		return domain.Attempt{}, domain.Scenario{}, err
	}
	return attempt, sc, nil
}
