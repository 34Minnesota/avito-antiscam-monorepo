package training_test

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"

	"github.com/34Minnesota/avito-antiscam-monorepo/backend/internal/domain"
	domainErrors "github.com/34Minnesota/avito-antiscam-monorepo/backend/internal/domain/errors"
	"github.com/34Minnesota/avito-antiscam-monorepo/backend/internal/services/training"
)

var errBoom = errors.New("boom")

func TestCatalogBuildsCardsWithPersonalStats(t *testing.T) {
	t.Parallel()
	first, second := uuid.New(), uuid.New()
	repo := &repositoryStub{
		scenarios: []domain.Scenario{testScenario(first), testScenario(second)},
		stats:     map[uuid.UUID]training.ScenarioStats{first: {BestScore: 80, AttemptsCount: 3}},
	}

	cards, err := training.New(repo).Catalog(context.Background(), mustUserID(t), domain.RoleSeller)
	if err != nil {
		t.Fatal(err)
	}
	if len(cards) != 2 {
		t.Fatalf("cards = %d", len(cards))
	}
	if cards[0].ID != first || cards[0].Slug != "test-scenario" || cards[0].Difficulty != 2 {
		t.Fatalf("unexpected card: %+v", cards[0])
	}
	if cards[0].Stats == nil || cards[0].Stats.BestScore != 80 || cards[0].Stats.AttemptsCount != 3 {
		t.Fatalf("stats were not attached: %+v", cards[0].Stats)
	}
	if cards[1].Stats != nil {
		t.Fatalf("scenario without attempts must have no stats: %+v", cards[1].Stats)
	}
}

func TestCatalogPropagatesRepositoryErrors(t *testing.T) {
	t.Parallel()
	for name, repo := range map[string]*repositoryStub{
		"list":  {listErr: errBoom},
		"stats": {statsErr: errBoom},
	} {
		if _, err := training.New(repo).Catalog(context.Background(), mustUserID(t), ""); !errors.Is(err, errBoom) {
			t.Fatalf("%s: got %v", name, err)
		}
	}
}

func TestStartCreatesAttemptOnFirstRun(t *testing.T) {
	t.Parallel()
	scenarioID := uuid.New()
	userID := mustUserID(t)
	repo := &repositoryStub{byID: map[uuid.UUID]domain.Scenario{scenarioID: testScenario(scenarioID)}}

	result, err := training.New(repo).Start(context.Background(), userID, scenarioID)
	if err != nil {
		t.Fatal(err)
	}
	if result.Scene.SceneID != "s1" || result.ScenesTotal != 3 || result.Title != "Тестовый сценарий" {
		t.Fatalf("unexpected result: %+v", result)
	}
	if len(repo.created) != 1 {
		t.Fatalf("attempts created = %d", len(repo.created))
	}
	created := repo.created[0]
	if created.UserID != userID || created.ScenarioID != scenarioID {
		t.Fatalf("unexpected attempt owner: %+v", created)
	}
	if created.Status != domain.AttemptInProgress || created.ID != result.AttemptID {
		t.Fatalf("unexpected attempt: %+v", created)
	}
}

func TestStartReusesActiveAttempt(t *testing.T) {
	t.Parallel()
	scenarioID, attemptID := uuid.New(), uuid.New()
	repo := &repositoryStub{
		byID:   map[uuid.UUID]domain.Scenario{scenarioID: testScenario(scenarioID)},
		active: &domain.Attempt{ID: attemptID, ScenarioID: scenarioID, State: domain.State{SceneIndex: 1}},
	}

	result, err := training.New(repo).Start(context.Background(), mustUserID(t), scenarioID)
	if err != nil {
		t.Fatal(err)
	}
	if result.AttemptID != attemptID || result.Scene.SceneID != "s2" {
		t.Fatalf("unexpected result: %+v", result)
	}
	if len(repo.created) != 0 {
		t.Fatalf("no new attempt must be created, got %d", len(repo.created))
	}
}

func TestStartRejectsActiveAttemptOutsideScenes(t *testing.T) {
	t.Parallel()
	scenarioID := uuid.New()
	repo := &repositoryStub{
		byID:   map[uuid.UUID]domain.Scenario{scenarioID: testScenario(scenarioID)},
		active: &domain.Attempt{ID: uuid.New(), ScenarioID: scenarioID, State: domain.State{SceneIndex: 99}},
	}

	_, err := training.New(repo).Start(context.Background(), mustUserID(t), scenarioID)
	if !errors.Is(err, domainErrors.ErrInvalidScenario) {
		t.Fatalf("got %v", err)
	}
	if len(repo.created) != 0 {
		t.Fatalf("broken state must not create a second attempt")
	}
}

func TestStartPropagatesErrors(t *testing.T) {
	t.Parallel()
	scenarioID := uuid.New()
	withScenario := func(mutate func(*repositoryStub)) *repositoryStub {
		repo := &repositoryStub{byID: map[uuid.UUID]domain.Scenario{scenarioID: testScenario(scenarioID)}}
		mutate(repo)
		return repo
	}

	cases := map[string]*repositoryStub{
		"scenario not found": {},
		"scenario lookup":    withScenario(func(r *repositoryStub) { r.byIDErr = errBoom }),
		"active lookup":      withScenario(func(r *repositoryStub) { r.activeErr = errBoom }),
		"create":             withScenario(func(r *repositoryStub) { r.createErr = errBoom }),
	}

	for name, repo := range cases {
		if _, err := training.New(repo).Start(context.Background(), mustUserID(t), scenarioID); err == nil {
			t.Fatalf("%s: expected error", name)
		}
	}
}

func TestStartRejectsStoredScenarioWithoutScenes(t *testing.T) {
	t.Parallel()
	scenarioID := uuid.New()
	broken := testScenario(scenarioID)
	broken.Doc.Scenes = nil
	repo := &repositoryStub{byID: map[uuid.UUID]domain.Scenario{scenarioID: broken}}

	_, err := training.New(repo).Start(context.Background(), mustUserID(t), scenarioID)
	if !errors.Is(err, domainErrors.ErrInvalidScenario) {
		t.Fatalf("got %v", err)
	}
}

func TestChooseSavesStepAndReturnsNextScene(t *testing.T) {
	t.Parallel()
	userID := mustUserID(t)
	scenarioID, attemptID := uuid.New(), uuid.New()
	repo := &repositoryStub{
		byID:    map[uuid.UUID]domain.Scenario{scenarioID: testScenario(scenarioID)},
		attempt: &domain.Attempt{ID: attemptID, UserID: userID, ScenarioID: scenarioID, Status: domain.AttemptInProgress},
	}

	result, err := training.New(repo).Choose(context.Background(), userID, attemptID, "s1", "a1", 0)
	if err != nil {
		t.Fatal(err)
	}
	if result.Finished || result.NextScene == nil || result.NextScene.SceneID != "s2" {
		t.Fatalf("unexpected result: %+v", result)
	}
	if result.Summary != nil {
		t.Fatalf("unfinished attempt must not carry a summary")
	}
	if result.Revision != 1 {
		t.Fatalf("revision = %d, want 1", result.Revision)
	}
	if len(repo.savedSteps) != 1 || repo.savedSteps[0].AttemptID != attemptID {
		t.Fatalf("unexpected saved step: %+v", repo.savedSteps)
	}
	if repo.savedSteps[0].CreatedAt.IsZero() {
		t.Fatalf("step must be timestamped")
	}
	if repo.finishCalls != 0 {
		t.Fatalf("attempt must not be finished yet")
	}
}

func TestChooseFinishesAttemptAndAttachesSummary(t *testing.T) {
	t.Parallel()
	userID := mustUserID(t)
	scenarioID, attemptID := uuid.New(), uuid.New()
	best := 40
	repo := &repositoryStub{
		byID: map[uuid.UUID]domain.Scenario{scenarioID: testScenario(scenarioID)},
		attempt: &domain.Attempt{
			ID: attemptID, UserID: userID, ScenarioID: scenarioID,
			Status: domain.AttemptInProgress, State: domain.State{SceneIndex: 2, Earned: 3},
		},
		journal: steps([2]string{"s1", "a1"}, [2]string{"s2", "b1"}, [2]string{"s3", "c1"}),
		best:    &best,
	}

	result, err := training.New(repo).Choose(context.Background(), userID, attemptID, "s3", "c1", 0)
	if err != nil {
		t.Fatal(err)
	}
	if !result.Finished || result.Summary == nil {
		t.Fatalf("unexpected result: %+v", result)
	}
	if result.Summary.Score != 100 || result.Summary.Outcome != domain.OutcomeSafe {
		t.Fatalf("unexpected summary: %+v", result.Summary)
	}
	if result.Summary.DeltaVsPrevious == nil || *result.Summary.DeltaVsPrevious != 60 {
		t.Fatalf("delta against previous best is wrong: %+v", result.Summary.DeltaVsPrevious)
	}
	if repo.finishCalls != 1 || repo.finishedScore != 100 || repo.finishedOutcome != domain.OutcomeSafe {
		t.Fatalf("attempt was not finished properly: calls=%d score=%d", repo.finishCalls, repo.finishedScore)
	}
	if len(repo.savedAttempts) != 1 || repo.savedAttempts[0].Status != domain.AttemptFinished {
		t.Fatalf("saved attempt must be marked finished: %+v", repo.savedAttempts)
	}
}

func TestChooseRejectsForeignAndFinishedAttempts(t *testing.T) {
	t.Parallel()
	scenarioID, attemptID := uuid.New(), uuid.New()
	userID := mustUserID(t)

	foreign := &repositoryStub{
		byID:    map[uuid.UUID]domain.Scenario{scenarioID: testScenario(scenarioID)},
		attempt: &domain.Attempt{ID: attemptID, UserID: mustUserID(t), ScenarioID: scenarioID},
	}
	_, err := training.New(foreign).Choose(context.Background(), userID, attemptID, "s1", "a1", 0)
	if !errors.Is(err, domainErrors.ErrForbidden) {
		t.Fatalf("foreign attempt: got %v", err)
	}

	finished := &repositoryStub{
		byID:    map[uuid.UUID]domain.Scenario{scenarioID: testScenario(scenarioID)},
		attempt: &domain.Attempt{ID: attemptID, UserID: userID, ScenarioID: scenarioID, Status: domain.AttemptFinished},
	}
	_, err = training.New(finished).Choose(context.Background(), userID, attemptID, "s1", "a1", 0)
	if !errors.Is(err, domainErrors.ErrAttemptFinished) {
		t.Fatalf("finished attempt: got %v", err)
	}
}

func TestChoosePropagatesErrors(t *testing.T) {
	t.Parallel()
	userID := mustUserID(t)
	scenarioID, attemptID := uuid.New(), uuid.New()

	base := func(mutate func(*repositoryStub)) *repositoryStub {
		repo := &repositoryStub{
			byID: map[uuid.UUID]domain.Scenario{scenarioID: testScenario(scenarioID)},
			attempt: &domain.Attempt{
				ID: attemptID, UserID: userID, ScenarioID: scenarioID,
				Status: domain.AttemptInProgress, State: domain.State{SceneIndex: 2},
			},
			journal: steps([2]string{"s3", "c1"}),
		}
		mutate(repo)
		return repo
	}

	cases := map[string]struct {
		repo  *repositoryStub
		scene string
		opt   string
	}{
		"attempt lookup": {repo: base(func(r *repositoryStub) { r.attemptErr = errBoom }), scene: "s3", opt: "c1"},
		"scenario lookup": {
			repo:  base(func(r *repositoryStub) { r.byIDErr = errBoom }),
			scene: "s3", opt: "c1",
		},
		"out of order": {repo: base(func(*repositoryStub) {}), scene: "s1", opt: "a1"},
		"save step":    {repo: base(func(r *repositoryStub) { r.saveErr = errBoom }), scene: "s3", opt: "c1"},
		"steps lookup": {repo: base(func(r *repositoryStub) { r.stepsErr = errBoom }), scene: "s3", opt: "c1"},
		"best score":   {repo: base(func(r *repositoryStub) { r.bestErr = errBoom }), scene: "s3", opt: "c1"},
		"finish attempt": {
			repo:  base(func(r *repositoryStub) { r.finishErr = errBoom }),
			scene: "s3", opt: "c1",
		},
	}

	for name, tc := range cases {
		if _, err := training.New(tc.repo).Choose(context.Background(), userID, attemptID, tc.scene, tc.opt, 0); err == nil {
			t.Fatalf("%s: expected error", name)
		}
	}
}

func TestSummaryReturnsResultWithDelta(t *testing.T) {
	t.Parallel()
	userID := mustUserID(t)
	scenarioID, attemptID := uuid.New(), uuid.New()
	best := 30
	repo := &repositoryStub{
		byID:    map[uuid.UUID]domain.Scenario{scenarioID: testScenario(scenarioID)},
		attempt: &domain.Attempt{ID: attemptID, UserID: userID, ScenarioID: scenarioID, Status: domain.AttemptFinished},
		journal: steps([2]string{"s1", "a2"}, [2]string{"s2", "b1"}, [2]string{"s3", "c1"}),
		best:    &best,
	}

	result, err := training.New(repo).Summary(context.Background(), userID, attemptID)
	if err != nil {
		t.Fatal(err)
	}
	if result.Score != 75 || result.Outcome != domain.OutcomePartial || len(result.MissedFlags) != 1 {
		t.Fatalf("unexpected summary: %+v", result)
	}
	if result.DeltaVsPrevious == nil || *result.DeltaVsPrevious != 45 {
		t.Fatalf("unexpected delta: %+v", result.DeltaVsPrevious)
	}
}

func TestSummaryWithoutPreviousAttemptsHasNoDelta(t *testing.T) {
	t.Parallel()
	userID := mustUserID(t)
	scenarioID, attemptID := uuid.New(), uuid.New()
	repo := &repositoryStub{
		byID:    map[uuid.UUID]domain.Scenario{scenarioID: testScenario(scenarioID)},
		attempt: &domain.Attempt{ID: attemptID, UserID: userID, ScenarioID: scenarioID, Status: domain.AttemptFinished},
	}

	result, err := training.New(repo).Summary(context.Background(), userID, attemptID)
	if err != nil {
		t.Fatal(err)
	}
	if result.DeltaVsPrevious != nil {
		t.Fatalf("first attempt must have no delta: %+v", result.DeltaVsPrevious)
	}
}

func TestSummaryPropagatesErrors(t *testing.T) {
	t.Parallel()
	userID := mustUserID(t)
	scenarioID, attemptID := uuid.New(), uuid.New()

	base := func(mutate func(*repositoryStub)) *repositoryStub {
		repo := &repositoryStub{
			byID:    map[uuid.UUID]domain.Scenario{scenarioID: testScenario(scenarioID)},
			attempt: &domain.Attempt{ID: attemptID, UserID: userID, ScenarioID: scenarioID, Status: domain.AttemptFinished},
		}
		mutate(repo)
		return repo
	}

	cases := map[string]*repositoryStub{
		"attempt not found": {},
		"attempt lookup":    base(func(r *repositoryStub) { r.attemptErr = errBoom }),
		"scenario lookup":   base(func(r *repositoryStub) { r.byIDErr = errBoom }),
		"steps lookup":      base(func(r *repositoryStub) { r.stepsErr = errBoom }),
		"best score":        base(func(r *repositoryStub) { r.bestErr = errBoom }),
		"broken journal":    base(func(r *repositoryStub) { r.journal = steps([2]string{"ghost", "a1"}) }),
	}

	for name, repo := range cases {
		if _, err := training.New(repo).Summary(context.Background(), userID, attemptID); err == nil {
			t.Fatalf("%s: expected error", name)
		}
	}
}
