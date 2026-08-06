package training

import (
	"context"
	"errors"
	"net"
	"net/url"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/34Minnesota/avito-antiscam-monorepo/backend/internal/domain"
	domainErrors "github.com/34Minnesota/avito-antiscam-monorepo/backend/internal/domain/errors"
	postgrespool "github.com/34Minnesota/avito-antiscam-monorepo/backend/internal/storage/postgres/pool"
)

const schema = `
CREATE TABLE IF NOT EXISTS sessions (
    id           UUID PRIMARY KEY,
    created_at   TIMESTAMPTZ NOT NULL,
    last_seen_at TIMESTAMPTZ NOT NULL
);

CREATE TABLE IF NOT EXISTS scenarios (
    id         UUID PRIMARY KEY,
    doc        JSONB NOT NULL,
    is_active  BOOLEAN NOT NULL DEFAULT true,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),

    slug       TEXT     GENERATED ALWAYS AS (doc ->> 'slug') STORED,
    role       TEXT     GENERATED ALWAYS AS (doc ->> 'role') STORED,
    title      TEXT     GENERATED ALWAYS AS (doc ->> 'title') STORED,
    difficulty SMALLINT GENERATED ALWAYS AS ((doc ->> 'difficulty')::smallint) STORED
);
CREATE UNIQUE INDEX IF NOT EXISTS idx_scenarios_slug ON scenarios (slug);

CREATE TABLE IF NOT EXISTS attempts (
    id          UUID PRIMARY KEY,
    user_id     UUID NOT NULL REFERENCES sessions (id) ON DELETE CASCADE,
    scenario_id UUID NOT NULL REFERENCES scenarios (id) ON DELETE CASCADE,
    status      TEXT NOT NULL DEFAULT 'in_progress' CHECK (status IN ('in_progress', 'finished')),
    state       JSONB NOT NULL,
    score       SMALLINT CHECK (score BETWEEN 0 AND 100),
    outcome     TEXT CHECK (outcome IN ('safe', 'partial', 'scammed')),
    started_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    finished_at TIMESTAMPTZ
);
CREATE UNIQUE INDEX IF NOT EXISTS idx_attempts_active
    ON attempts (user_id, scenario_id) WHERE status = 'in_progress';

CREATE TABLE IF NOT EXISTS attempt_steps (
    id         BIGSERIAL PRIMARY KEY,
    attempt_id UUID NOT NULL REFERENCES attempts (id) ON DELETE CASCADE,
    scene_id   TEXT NOT NULL,
    option_id  TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);
`

type repoEnv struct {
	repo      *Repository
	pool      *postgrespool.Pool
	ctx       context.Context
	userID    domain.UserID
	buyerID   uuid.UUID
	sellerID  uuid.UUID
	attemptID uuid.UUID
}

func TestRepositoryPostgres(t *testing.T) {
	env := setup(t)

	t.Run("UpsertScenario is idempotent by slug", env.upsertIsIdempotent)
	t.Run("ListScenarios filters by role", env.listFiltersByRole)
	t.Run("ScenarioByID reports missing scenario", env.scenarioNotFound)
	t.Run("CreateAttempt stores a fresh attempt", env.createAttempt)
	t.Run("AttemptByID and ActiveAttempt find it", env.lookupAttempt)
	t.Run("SaveStep does not touch status", env.saveStepKeepsStatus)
	t.Run("Steps returns journal in play order", env.stepsInPlayOrder)
	t.Run("FinishAttempt is idempotent", env.finishIsIdempotent)
	t.Run("ScenarioStats counts every attempt", env.scenarioStats)
	t.Run("BestScore excludes the given attempt", env.bestScoreExcludes)
}

func (e *repoEnv) upsertIsIdempotent(t *testing.T) {
	if err := e.repo.UpsertScenario(e.ctx, scenario(e.buyerID, "buyer-one", domain.RoleBuyer, 1)); err != nil {
		t.Fatal(err)
	}

	updated := scenario(uuid.New(), "buyer-one", domain.RoleBuyer, 2)
	updated.Doc.Title = "Обновлённый"
	if err := e.repo.UpsertScenario(e.ctx, updated); err != nil {
		t.Fatal(err)
	}

	stored, err := e.repo.ScenarioByID(e.ctx, e.buyerID)
	if err != nil {
		t.Fatal(err)
	}
	if stored.Doc.Title != "Обновлённый" {
		t.Fatalf("title = %q", stored.Doc.Title)
	}
	if stored.Doc.Difficulty != 2 {
		t.Fatalf("difficulty = %d", stored.Doc.Difficulty)
	}
}

func (e *repoEnv) listFiltersByRole(t *testing.T) {
	if err := e.repo.UpsertScenario(e.ctx, scenario(e.sellerID, "seller-one", domain.RoleSeller, 3)); err != nil {
		t.Fatal(err)
	}

	all, err := e.repo.ListScenarios(e.ctx, "")
	if err != nil {
		t.Fatal(err)
	}
	if len(all) != 2 {
		t.Fatalf("all scenarios = %d", len(all))
	}

	sellers, err := e.repo.ListScenarios(e.ctx, domain.RoleSeller)
	if err != nil {
		t.Fatal(err)
	}
	if len(sellers) != 1 {
		t.Fatalf("seller scenarios = %d", len(sellers))
	}
	if sellers[0].Doc.Slug != "seller-one" {
		t.Fatalf("slug = %q", sellers[0].Doc.Slug)
	}
}

func (e *repoEnv) scenarioNotFound(t *testing.T) {
	if _, err := e.repo.ScenarioByID(e.ctx, uuid.New()); !errors.Is(err, domainErrors.ErrNotFound) {
		t.Fatalf("got %v", err)
	}
}

func (e *repoEnv) createAttempt(t *testing.T) {
	if _, err := e.repo.ActiveAttempt(e.ctx, e.userID, e.buyerID); !errors.Is(err, domainErrors.ErrNotFound) {
		t.Fatalf("no active attempt expected yet: %v", err)
	}
	if _, err := e.repo.AttemptByID(e.ctx, e.attemptID); !errors.Is(err, domainErrors.ErrNotFound) {
		t.Fatalf("attempt must not exist yet: %v", err)
	}

	attempt := domain.Attempt{
		ID:         e.attemptID,
		UserID:     e.userID,
		ScenarioID: e.buyerID,
		Status:     domain.AttemptInProgress,
		State:      domain.State{SceneIndex: 0, Flags: []string{}},
		StartedAt:  time.Now().UTC().Truncate(time.Microsecond),
	}
	if err := e.repo.CreateAttempt(e.ctx, attempt); err != nil {
		t.Fatal(err)
	}
}

func (e *repoEnv) lookupAttempt(t *testing.T) {
	stored, err := e.repo.AttemptByID(e.ctx, e.attemptID)
	if err != nil {
		t.Fatal(err)
	}
	if stored.UserID != e.userID || stored.ScenarioID != e.buyerID {
		t.Fatalf("unexpected owner: %+v", stored)
	}
	if stored.Status != domain.AttemptInProgress {
		t.Fatalf("status = %q", stored.Status)
	}

	if stored.Score != nil || stored.Outcome != nil || stored.FinishedAt != nil {
		t.Fatalf("fresh attempt must have empty result: %+v", stored)
	}

	active, err := e.repo.ActiveAttempt(e.ctx, e.userID, e.buyerID)
	if err != nil {
		t.Fatal(err)
	}
	if active.ID != e.attemptID {
		t.Fatalf("active attempt = %s", active.ID)
	}
}

func (e *repoEnv) saveStepKeepsStatus(t *testing.T) {
	attempt, err := e.repo.AttemptByID(e.ctx, e.attemptID)
	if err != nil {
		t.Fatal(err)
	}
	attempt.State = domain.State{SceneIndex: 1, Earned: 1, Flags: []string{"f1"}}

	attempt.Status = domain.AttemptFinished

	step := domain.AttemptStep{
		AttemptID: e.attemptID,
		SceneID:   "s1",
		OptionID:  "a1",
		CreatedAt: time.Now().UTC().Truncate(time.Microsecond),
	}
	if err := e.repo.SaveStep(e.ctx, step, attempt); err != nil {
		t.Fatal(err)
	}

	stored, err := e.repo.AttemptByID(e.ctx, e.attemptID)
	if err != nil {
		t.Fatal(err)
	}
	if stored.Status != domain.AttemptInProgress {
		t.Fatalf("status must be owned by FinishAttempt, got %q", stored.Status)
	}
	if stored.State.SceneIndex != 1 || stored.State.Earned != 1 {
		t.Fatalf("state was not persisted: %+v", stored.State)
	}
}

func (e *repoEnv) stepsInPlayOrder(t *testing.T) {
	attempt, err := e.repo.AttemptByID(e.ctx, e.attemptID)
	if err != nil {
		t.Fatal(err)
	}
	second := domain.AttemptStep{
		AttemptID: e.attemptID, SceneID: "s2", OptionID: "b1",
		CreatedAt: time.Now().UTC().Truncate(time.Microsecond),
	}
	if err := e.repo.SaveStep(e.ctx, second, attempt); err != nil {
		t.Fatal(err)
	}

	journal, err := e.repo.Steps(e.ctx, e.attemptID)
	if err != nil {
		t.Fatal(err)
	}
	if len(journal) != 2 {
		t.Fatalf("journal = %d", len(journal))
	}
	if journal[0].SceneID != "s1" || journal[1].SceneID != "s2" {
		t.Fatalf("unexpected order: %+v", journal)
	}

	if journal[0].Verdict != "" || journal[0].Flag != "" || journal[0].Weight != 0 {
		t.Fatalf("journal must stay minimal: %+v", journal[0])
	}
}

func (e *repoEnv) finishIsIdempotent(t *testing.T) {
	at := time.Now().UTC().Truncate(time.Microsecond)
	if err := e.repo.FinishAttempt(e.ctx, e.attemptID, 75, domain.OutcomePartial, at); err != nil {
		t.Fatal(err)
	}

	stored, err := e.repo.AttemptByID(e.ctx, e.attemptID)
	if err != nil {
		t.Fatal(err)
	}
	if stored.Status != domain.AttemptFinished {
		t.Fatalf("status = %q", stored.Status)
	}
	if stored.Score == nil || *stored.Score != 75 {
		t.Fatalf("score = %+v", stored.Score)
	}
	if stored.Outcome == nil || *stored.Outcome != domain.OutcomePartial {
		t.Fatalf("outcome = %+v", stored.Outcome)
	}
	if stored.FinishedAt == nil {
		t.Fatal("finished_at must be set")
	}

	err = e.repo.FinishAttempt(e.ctx, e.attemptID, 10, domain.OutcomeScammed, at)
	if !errors.Is(err, domainErrors.ErrAttemptFinished) {
		t.Fatalf("second finish: got %v", err)
	}

	again, err := e.repo.AttemptByID(e.ctx, e.attemptID)
	if err != nil {
		t.Fatal(err)
	}
	if *again.Score != 75 {
		t.Fatalf("score was overwritten: %d", *again.Score)
	}
}

func (e *repoEnv) scenarioStats(t *testing.T) {
	fresh := domain.Attempt{
		ID: uuid.New(), UserID: e.userID, ScenarioID: e.buyerID,
		Status: domain.AttemptInProgress, State: domain.State{}, StartedAt: time.Now().UTC(),
	}
	if err := e.repo.CreateAttempt(e.ctx, fresh); err != nil {
		t.Fatal(err)
	}

	stats, err := e.repo.ScenarioStats(e.ctx, e.userID)
	if err != nil {
		t.Fatal(err)
	}
	stat, ok := stats[e.buyerID]
	if !ok {
		t.Fatalf("no stats for scenario: %+v", stats)
	}
	if stat.AttemptsCount != 2 {
		t.Fatalf("attempts must include unfinished ones, got %d", stat.AttemptsCount)
	}
	if stat.BestScore != 75 {
		t.Fatalf("best score = %d", stat.BestScore)
	}
}

func (e *repoEnv) bestScoreExcludes(t *testing.T) {
	best, err := e.repo.BestScore(e.ctx, e.userID, e.buyerID, uuid.New())
	if err != nil {
		t.Fatal(err)
	}
	if best == nil || *best != 75 {
		t.Fatalf("unexpected best: %+v", best)
	}

	excluded, err := e.repo.BestScore(e.ctx, e.userID, e.buyerID, e.attemptID)
	if err != nil {
		t.Fatal(err)
	}
	if excluded != nil {
		t.Fatalf("expected nil, got %d", *excluded)
	}
}

func setup(t *testing.T) *repoEnv {
	t.Helper()

	dsn := os.Getenv("TEST_DATABASE_URL")
	if dsn == "" {
		t.Skip("TEST_DATABASE_URL is not set; PostgreSQL integration test was not run")
	}

	config, err := configFromDSN(dsn)
	if err != nil {
		t.Fatal(err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	t.Cleanup(cancel)

	pool, err := postgrespool.NewPool(ctx, config)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(pool.Close)

	if _, err := pool.Exec(ctx, schema); err != nil {
		t.Fatal(err)
	}
	if _, err := pool.Exec(ctx, `TRUNCATE attempt_steps, attempts, scenarios, sessions CASCADE`); err != nil {
		t.Fatal(err)
	}

	sessionID := uuid.New()
	if _, err := pool.Exec(ctx, `INSERT INTO sessions VALUES ($1, $2, $2)`, sessionID, time.Now().UTC()); err != nil {
		t.Fatal(err)
	}

	userID, err := domain.NewUserID(sessionID)
	if err != nil {
		t.Fatal(err)
	}

	return &repoEnv{
		repo:      New(pool),
		pool:      pool,
		ctx:       ctx,
		userID:    userID,
		buyerID:   uuid.New(),
		sellerID:  uuid.New(),
		attemptID: uuid.New(),
	}
}
func TestRepositoryReportsQueryFailures(t *testing.T) {
	env := setup(t)
	env.pool.Close()

	repo, ctx := New(env.pool), env.ctx
	calls := map[string]func() error{
		"UpsertScenario": func() error {
			return repo.UpsertScenario(ctx, scenario(uuid.New(), "x", domain.RoleBuyer, 1))
		},
		"ListScenarios": func() error { _, err := repo.ListScenarios(ctx, ""); return err },
		"ScenarioByID":  func() error { _, err := repo.ScenarioByID(ctx, uuid.New()); return err },
		"ScenarioStats": func() error { _, err := repo.ScenarioStats(ctx, env.userID); return err },
		"CreateAttempt": func() error {
			return repo.CreateAttempt(ctx, domain.Attempt{ID: uuid.New(), UserID: env.userID})
		},
		"AttemptByID":   func() error { _, err := repo.AttemptByID(ctx, uuid.New()); return err },
		"ActiveAttempt": func() error { _, err := repo.ActiveAttempt(ctx, env.userID, uuid.New()); return err },
		"SaveStep": func() error {
			return repo.SaveStep(ctx, domain.AttemptStep{}, domain.Attempt{ID: uuid.New()})
		},
		"FinishAttempt": func() error {
			return repo.FinishAttempt(ctx, uuid.New(), 10, domain.OutcomeSafe, time.Now())
		},
		"Steps":     func() error { _, err := repo.Steps(ctx, uuid.New()); return err },
		"BestScore": func() error { _, err := repo.BestScore(ctx, env.userID, uuid.New(), uuid.New()); return err },
	}

	for name, call := range calls {
		if err := call(); err == nil {
			t.Errorf("%s: expected an error on a closed pool", name)
		}
	}
}

func TestRepositoryRejectsCorruptedJSON(t *testing.T) {
	env := setup(t)

	scenarioID := uuid.New()
	_, err := env.pool.Exec(env.ctx, `INSERT INTO scenarios (id, doc) VALUES ($1, '"not an object"'::jsonb)`, scenarioID)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := env.repo.ScenarioByID(env.ctx, scenarioID); err == nil {
		t.Fatal("expected an error on a scenario document that is not an object")
	}
	if _, err := env.repo.ListScenarios(env.ctx, ""); err == nil {
		t.Fatal("expected an error while listing a broken scenario")
	}

	if err := env.repo.UpsertScenario(env.ctx, scenario(env.buyerID, "buyer-one", domain.RoleBuyer, 1)); err != nil {
		t.Fatal(err)
	}
	attemptID := uuid.New()
	_, err = env.pool.Exec(env.ctx,
		`INSERT INTO attempts (id, user_id, scenario_id, status, state) VALUES ($1, $2, $3, 'in_progress', '"broken"'::jsonb)`,
		attemptID, env.userID.UUID(), env.buyerID)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := env.repo.AttemptByID(env.ctx, attemptID); err == nil {
		t.Fatal("expected an error on a broken attempt state")
	}
}

func TestSaveStepRejectsUnknownAttempt(t *testing.T) {
	env := setup(t)

	step := domain.AttemptStep{AttemptID: uuid.New(), SceneID: "s1", OptionID: "a1", CreatedAt: time.Now().UTC()}
	err := env.repo.SaveStep(env.ctx, step, domain.Attempt{ID: uuid.New(), State: domain.State{}})
	if err == nil {
		t.Fatal("expected a foreign key violation")
	}
}

func configFromDSN(dsn string) (postgrespool.Config, error) {
	parsed, err := url.Parse(dsn)
	if err != nil {
		return postgrespool.Config{}, err
	}

	host, port, err := net.SplitHostPort(parsed.Host)
	if err != nil {
		host, port = parsed.Host, "5432"
	}

	password, _ := parsed.User.Password()
	return postgrespool.Config{
		Host:     host,
		Port:     port,
		User:     parsed.User.Username(),
		Password: password,
		Database: strings.TrimPrefix(parsed.Path, "/"),
		Timeout:  20 * time.Second,
	}, nil
}

func scenario(id uuid.UUID, slug string, role domain.Role, difficulty int) domain.Scenario {
	return domain.Scenario{
		ID:       id,
		IsActive: true,
		Doc: domain.ScenarioDoc{
			Version:    1,
			Slug:       slug,
			Role:       role,
			Category:   "test",
			Difficulty: difficulty,
			Title:      "Сценарий " + slug,
			Scenes: []domain.Scene{{
				ID:     "s1",
				Weight: 1,
				Decision: domain.Decision{
					Prompt:  "Что делать?",
					Options: []domain.Option{{ID: "a1", Text: "Безопасно", Verdict: domain.VerdictSafe}},
				},
			}},
			Endings: map[string]domain.Ending{
				"safe":    {Outcome: domain.OutcomeSafe, Title: "Чисто"},
				"partial": {Outcome: domain.OutcomePartial, Title: "Почти"},
			},
		},
	}
}
