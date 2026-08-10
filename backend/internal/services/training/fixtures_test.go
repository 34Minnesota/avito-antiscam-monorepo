package training_test

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"

	domainErrors "github.com/34Minnesota/avito-antiscam-monorepo/backend/internal/domain/errors"
	"github.com/34Minnesota/avito-antiscam-monorepo/backend/internal/domain/models"
	"github.com/34Minnesota/avito-antiscam-monorepo/backend/internal/services/training"
)

func testDoc() models.ScenarioDoc {
	return models.ScenarioDoc{
		Version:     1,
		Slug:        "test-scenario",
		Role:        models.RoleSeller,
		Category:    "test",
		Difficulty:  2,
		Title:       "Тестовый сценарий",
		Description: "Описание",
		Listing:     models.Listing{Title: "Товар", Price: 100, Location: "Москва"},
		Counterpart: models.Counterpart{Name: "Игорь", Rating: 4.5, Reviews: 10},
		Scenes: []models.Scene{
			{
				ID:     "s1",
				Weight: 1,
				Intro:  []models.Message{{Author: models.AuthorCounterpart, Text: "Привет"}},
				Decision: models.Decision{
					Prompt: "Что делать?",
					Options: []models.Option{
						{
							ID: "a1", Text: "Безопасно", Reply: "Пишу безопасно",
							Verdict: models.VerdictSafe, Feedback: "ок",
							Reaction: []models.Message{{Author: models.AuthorCounterpart, Text: "Ответ"}},
						},
						{ID: "a2", Text: "Рискованно", Reply: "Пишу рискованно", Verdict: models.VerdictRisky, Flag: "f1", Feedback: "так себе"},
						{ID: "a3", Text: "Провал", Reply: "Пишу провал", Verdict: models.VerdictFatal, Flag: "f2", Ending: "lost", Feedback: "плохо"},
					},
				},
			},
			{
				ID:     "s2",
				Weight: 2,
				Decision: models.Decision{
					Prompt: "И теперь?",
					Options: []models.Option{
						{ID: "b1", Text: "Безопасно", Verdict: models.VerdictSafe, Feedback: "ок"},
						{ID: "b2", Text: "Рискованно", Verdict: models.VerdictRisky, Flag: "f1", Feedback: "так себе"},
					},
				},
			},
			{
				ID:     "s3",
				Weight: 1,
				Decision: models.Decision{
					Prompt: "Финал",
					Options: []models.Option{
						{ID: "c1", Text: "Безопасно", Verdict: models.VerdictSafe, Feedback: "ок"},
						{ID: "c2", Text: "Рискованно", Verdict: models.VerdictRisky, Flag: "ghost", Feedback: "так себе"},
					},
				},
			},
		},
		Endings: map[string]models.Ending{
			"safe":    {Outcome: models.OutcomeSafe, Title: "Чисто", Text: "Молодец"},
			"partial": {Outcome: models.OutcomePartial, Title: "Почти", Text: "Бывает"},
			"lost":    {Outcome: models.OutcomeScammed, Title: "Обманули", Text: "Увы"},
		},
		Debrief: models.Debrief{
			KeyFlags: []models.FlagInfo{
				{ID: "f1", Title: "Первый", Text: "Описание первого"},
				{ID: "f2", Title: "Второй", Text: "Описание второго"},
			},
			Takeaway: "Вывод",
		},
	}
}

func testScenario(id uuid.UUID) models.Scenario {
	return models.Scenario{ID: id, IsActive: true, Doc: testDoc()}
}

func steps(pairs ...[2]string) []models.AttemptStep {
	out := make([]models.AttemptStep, 0, len(pairs))
	for _, p := range pairs {
		out = append(out, models.AttemptStep{SceneID: p[0], OptionID: p[1]})
	}
	return out
}

func mustUserID(t *testing.T) models.UserID {
	t.Helper()
	id, err := models.NewUserID(uuid.New())
	if err != nil {
		t.Fatal(err)
	}
	return id
}

type repositoryStub struct {
	scenarios []models.Scenario
	byID      map[uuid.UUID]models.Scenario
	stats     map[uuid.UUID]training.ScenarioStats
	active    *models.Attempt
	attempt   *models.Attempt
	journal   []models.AttemptStep
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

	created         []models.Attempt
	savedSteps      []models.AttemptStep
	savedAttempts   []models.Attempt
	upserted        []models.Scenario
	finishCalls     int
	finishedScore   int
	finishedOutcome models.Outcome
}

func (r *repositoryStub) UpsertScenario(_ context.Context, s models.Scenario) error {
	if r.upsertErr != nil {
		return r.upsertErr
	}
	r.upserted = append(r.upserted, s)
	return nil
}

func (r *repositoryStub) ListScenarios(_ context.Context, _ models.Role) ([]models.Scenario, error) {
	return r.scenarios, r.listErr
}

func (r *repositoryStub) ScenarioByID(_ context.Context, id uuid.UUID) (models.Scenario, error) {
	if r.byIDErr != nil {
		return models.Scenario{}, r.byIDErr
	}
	if s, ok := r.byID[id]; ok {
		return s, nil
	}
	return models.Scenario{}, domainErrors.ErrNotFound
}

func (r *repositoryStub) ScenarioStats(_ context.Context, _ models.UserID) (map[uuid.UUID]training.ScenarioStats, error) {
	return r.stats, r.statsErr
}

func (r *repositoryStub) CreateAttempt(_ context.Context, a models.Attempt) error {
	if r.createErr != nil {
		return r.createErr
	}
	r.created = append(r.created, a)
	return nil
}

func (r *repositoryStub) AttemptByID(_ context.Context, _ uuid.UUID) (models.Attempt, error) {
	if r.attemptErr != nil {
		return models.Attempt{}, r.attemptErr
	}
	if r.attempt == nil {
		return models.Attempt{}, domainErrors.ErrNotFound
	}
	return *r.attempt, nil
}

func (r *repositoryStub) ActiveAttempt(_ context.Context, _ models.UserID, _ uuid.UUID) (models.Attempt, error) {
	if r.activeErr != nil {
		return models.Attempt{}, r.activeErr
	}
	if r.active == nil {
		return models.Attempt{}, domainErrors.ErrNotFound
	}
	return *r.active, nil
}

func (r *repositoryStub) SaveStep(
	_ context.Context,
	step models.AttemptStep,
	a models.Attempt,
	_ models.UserID,
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
	outcome models.Outcome,
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

func (r *repositoryStub) Steps(_ context.Context, _ uuid.UUID) ([]models.AttemptStep, error) {
	return r.journal, r.stepsErr
}

func (r *repositoryStub) BestScore(_ context.Context, _ models.UserID, _, _ uuid.UUID) (*int, error) {
	return r.best, r.bestErr
}
