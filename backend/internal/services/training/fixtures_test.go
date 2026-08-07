package training_test

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/34Minnesota/avito-antiscam-monorepo/backend/internal/domain"
	domainErrors "github.com/34Minnesota/avito-antiscam-monorepo/backend/internal/domain/errors"
	"github.com/34Minnesota/avito-antiscam-monorepo/backend/internal/services/training"
)

func testDoc() domain.ScenarioDoc {
	return domain.ScenarioDoc{
		Version:     1,
		Slug:        "test-scenario",
		Role:        domain.RoleSeller,
		Category:    "test",
		Difficulty:  2,
		Title:       "Тестовый сценарий",
		Description: "Описание",
		Listing:     domain.Listing{Title: "Товар", Price: 100, Location: "Москва"},
		Counterpart: domain.Counterpart{Name: "Игорь", Rating: 4.5, Reviews: 10},
		Scenes: []domain.Scene{
			{
				ID:     "s1",
				Weight: 1,
				Intro:  []domain.Message{{Author: domain.AuthorCounterpart, Text: "Привет"}},
				Decision: domain.Decision{
					Prompt: "Что делать?",
					Options: []domain.Option{
						{ID: "a1", Text: "Безопасно", Verdict: domain.VerdictSafe, Feedback: "ок"},
						{ID: "a2", Text: "Рискованно", Verdict: domain.VerdictRisky, Flag: "f1", Feedback: "так себе"},
						{ID: "a3", Text: "Провал", Verdict: domain.VerdictFatal, Flag: "f2", Ending: "lost", Feedback: "плохо"},
					},
				},
			},
			{
				ID:     "s2",
				Weight: 2,
				Decision: domain.Decision{
					Prompt: "И теперь?",
					Options: []domain.Option{
						{ID: "b1", Text: "Безопасно", Verdict: domain.VerdictSafe, Feedback: "ок"},
						{ID: "b2", Text: "Рискованно", Verdict: domain.VerdictRisky, Flag: "f1", Feedback: "так себе"},
					},
				},
			},
			{
				ID:     "s3",
				Weight: 1,
				Decision: domain.Decision{
					Prompt: "Финал",
					Options: []domain.Option{
						{ID: "c1", Text: "Безопасно", Verdict: domain.VerdictSafe, Feedback: "ок"},
						{ID: "c2", Text: "Рискованно", Verdict: domain.VerdictRisky, Flag: "ghost", Feedback: "так себе"},
					},
				},
			},
		},
		Endings: map[string]domain.Ending{
			"safe":    {Outcome: domain.OutcomeSafe, Title: "Чисто", Text: "Молодец"},
			"partial": {Outcome: domain.OutcomePartial, Title: "Почти", Text: "Бывает"},
			"lost":    {Outcome: domain.OutcomeScammed, Title: "Обманули", Text: "Увы"},
		},
		Debrief: domain.Debrief{
			KeyFlags: []domain.FlagInfo{
				{ID: "f1", Title: "Первый", Text: "Описание первого"},
				{ID: "f2", Title: "Второй", Text: "Описание второго"},
			},
			Takeaway: "Вывод",
		},
	}
}

func testScenario(id uuid.UUID) domain.Scenario {
	return domain.Scenario{ID: id, IsActive: true, Doc: testDoc()}
}

func steps(pairs ...[2]string) []domain.AttemptStep {
	out := make([]domain.AttemptStep, 0, len(pairs))
	for _, p := range pairs {
		out = append(out, domain.AttemptStep{SceneID: p[0], OptionID: p[1]})
	}
	return out
}

func mustUserID(t *testing.T) domain.UserID {
	t.Helper()
	id, err := domain.NewUserID(uuid.New())
	if err != nil {
		t.Fatal(err)
	}
	return id
}

type repositoryStub struct {
	scenarios []domain.Scenario
	byID      map[uuid.UUID]domain.Scenario
	stats     map[uuid.UUID]training.ScenarioStats
	active    *domain.Attempt
	attempt   *domain.Attempt
	journal   []domain.AttemptStep
	best      *int

	listErr    error
	statsErr   error
	byIDErr    error
	activeErr  error
	attemptErr error
	createErr  error
	saveErr    error
	finishErr  error
	stepsErr   error
	bestErr    error
	upsertErr  error

	created         []domain.Attempt
	savedSteps      []domain.AttemptStep
	savedAttempts   []domain.Attempt
	upserted        []domain.Scenario
	finishCalls     int
	finishedScore   int
	finishedOutcome domain.Outcome
}

func (r *repositoryStub) UpsertScenario(_ context.Context, s domain.Scenario) error {
	if r.upsertErr != nil {
		return r.upsertErr
	}
	r.upserted = append(r.upserted, s)
	return nil
}

func (r *repositoryStub) ListScenarios(_ context.Context, _ domain.Role) ([]domain.Scenario, error) {
	return r.scenarios, r.listErr
}

func (r *repositoryStub) ScenarioByID(_ context.Context, id uuid.UUID) (domain.Scenario, error) {
	if r.byIDErr != nil {
		return domain.Scenario{}, r.byIDErr
	}
	if s, ok := r.byID[id]; ok {
		return s, nil
	}
	return domain.Scenario{}, domainErrors.ErrNotFound
}

func (r *repositoryStub) ScenarioStats(_ context.Context, _ domain.UserID) (map[uuid.UUID]training.ScenarioStats, error) {
	return r.stats, r.statsErr
}

func (r *repositoryStub) CreateAttempt(_ context.Context, a domain.Attempt) error {
	if r.createErr != nil {
		return r.createErr
	}
	r.created = append(r.created, a)
	return nil
}

func (r *repositoryStub) AttemptByID(_ context.Context, _ uuid.UUID) (domain.Attempt, error) {
	if r.attemptErr != nil {
		return domain.Attempt{}, r.attemptErr
	}
	if r.attempt == nil {
		return domain.Attempt{}, domainErrors.ErrNotFound
	}
	return *r.attempt, nil
}

func (r *repositoryStub) ActiveAttempt(_ context.Context, _ domain.UserID, _ uuid.UUID) (domain.Attempt, error) {
	if r.activeErr != nil {
		return domain.Attempt{}, r.activeErr
	}
	if r.active == nil {
		return domain.Attempt{}, domainErrors.ErrNotFound
	}
	return *r.active, nil
}

func (r *repositoryStub) SaveStep(
	_ context.Context,
	step domain.AttemptStep,
	a domain.Attempt,
	_ domain.UserID,
	expectedRevision int,
) (int, error) {
	if r.saveErr != nil {
		return 0, r.saveErr
	}
	r.savedSteps = append(r.savedSteps, step)
	r.savedAttempts = append(r.savedAttempts, a)
	return expectedRevision + 1, nil
}

func (r *repositoryStub) FinishAttempt(
	_ context.Context,
	_ uuid.UUID,
	score int,
	outcome domain.Outcome,
	_ time.Time,
) error {
	if r.finishErr != nil {
		return r.finishErr
	}
	r.finishCalls++
	r.finishedScore = score
	r.finishedOutcome = outcome
	return nil
}

func (r *repositoryStub) Steps(_ context.Context, _ uuid.UUID) ([]domain.AttemptStep, error) {
	return r.journal, r.stepsErr
}

func (r *repositoryStub) BestScore(_ context.Context, _ domain.UserID, _, _ uuid.UUID) (*int, error) {
	return r.best, r.bestErr
}
