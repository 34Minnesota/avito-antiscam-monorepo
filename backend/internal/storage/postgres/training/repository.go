package training

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	domainErrors "github.com/34Minnesota/avito-antiscam-monorepo/backend/internal/domain/errors"
	"github.com/34Minnesota/avito-antiscam-monorepo/backend/internal/domain/models"
	trainingservice "github.com/34Minnesota/avito-antiscam-monorepo/backend/internal/services/training"
	postgrespool "github.com/34Minnesota/avito-antiscam-monorepo/backend/internal/storage/postgres/pool"
)

type Repository struct {
	pool *postgrespool.Pool
}

func New(pool *postgrespool.Pool) *Repository {
	return &Repository{pool: pool}
}

// UpsertScenario создаёт сценарий либо подтверждает, что существующий сценарий
// с тем же slug содержит идентичный документ.
func (r *Repository) UpsertScenario(ctx context.Context, s models.Scenario) error {
	ctx, cancel := context.WithTimeout(ctx, r.pool.OpTimeout())
	defer cancel()

	doc, err := json.Marshal(s.Doc)
	if err != nil {
		return fmt.Errorf("scenario serializing %s: %w", s.Doc.Slug, err)
	}

	// Документ обновляется только при возросшем version. Так правка контента
	// доезжает обычным сидом, но случайное расхождение на той же версии
	// по-прежнему считается ошибкой и разбирается ниже.
	const insert = `
		INSERT INTO scenarios (id, doc)
		VALUES ($1, $2)
		ON CONFLICT (slug) DO UPDATE SET
			doc        = EXCLUDED.doc,
			updated_at = now()
		WHERE COALESCE((scenarios.doc ->> 'version')::int, 0)
		    < COALESCE((EXCLUDED.doc  ->> 'version')::int, 0)
		RETURNING id`

	var id uuid.UUID
	err = r.pool.QueryRow(ctx, insert, s.ID, doc).Scan(&id)
	if err == nil {
		return nil
	}
	if !errors.Is(err, pgx.ErrNoRows) {
		return fmt.Errorf("creating scenario %s: %w", s.Doc.Slug, err)
	}

	const sameDocument = `
		SELECT doc = $2::jsonb
		FROM scenarios
		WHERE slug = $1`

	var same bool
	err = r.pool.QueryRow(ctx, sameDocument, s.Doc.Slug, doc).Scan(&same)
	if errors.Is(err, pgx.ErrNoRows) {
		return fmt.Errorf("reading existing scenario %s: %w", s.Doc.Slug, domainErrors.ErrNotFound)
	}
	if err != nil {
		return fmt.Errorf("reading existing scenario %s: %w", s.Doc.Slug, err)
	}
	if !same {
		return fmt.Errorf("scenario %q already exists with different document", s.Doc.Slug)
	}

	return nil
}

// ListScenarios возвращает активные сценарии, опционально фильтруя по роли.
func (r *Repository) ListScenarios(ctx context.Context, role models.Role) ([]models.Scenario, error) {
	ctx, cancel := context.WithTimeout(ctx, r.pool.OpTimeout())
	defer cancel()

	const q = `
		SELECT id, doc, is_active
		FROM scenarios
		WHERE is_active AND ($1 = '' OR role = $1)
		ORDER BY role, difficulty, title`

	rows, err := r.pool.Query(ctx, q, string(role))
	if err != nil {
		return nil, fmt.Errorf("list scenarios: %w", err)
	}
	defer rows.Close()

	scenarios := make([]models.Scenario, 0, 8)
	for rows.Next() {
		s, err := scanScenario(rows)
		if err != nil {
			return nil, err
		}
		scenarios = append(scenarios, s)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("reading scenarios: %w", err)
	}
	return scenarios, nil
}

// ScenarioByID возвращает сценарий вместе с документом.
func (r *Repository) ScenarioByID(ctx context.Context, id uuid.UUID) (models.Scenario, error) {
	ctx, cancel := context.WithTimeout(ctx, r.pool.OpTimeout())
	defer cancel()

	const q = `SELECT id, doc, is_active FROM scenarios WHERE id = $1 AND is_active`

	s, err := scanScenario(r.pool.QueryRow(ctx, q, id))
	if errors.Is(err, pgx.ErrNoRows) {
		return models.Scenario{}, fmt.Errorf("scenario %s: %w", id, domainErrors.ErrNotFound)
	}
	if err != nil {
		return models.Scenario{}, err
	}
	return s, nil
}

// ScenarioStats возвращает статистику пользователя по каждому сценарию.
func (r *Repository) ScenarioStats(
	ctx context.Context,
	userID models.UserID,
) (map[uuid.UUID]trainingservice.ScenarioStats, error) {
	ctx, cancel := context.WithTimeout(ctx, r.pool.OpTimeout())
	defer cancel()

	const q = `
		SELECT scenario_id,
		       COALESCE(MAX(score) FILTER (WHERE status = 'finished'), 0),
		       COUNT(*)
		FROM attempts
		WHERE user_id = $1
		GROUP BY scenario_id`

	rows, err := r.pool.Query(ctx, q, userID.UUID())
	if err != nil {
		return nil, fmt.Errorf("scenario stats: %w", err)
	}
	defer rows.Close()

	stats := make(map[uuid.UUID]trainingservice.ScenarioStats)
	for rows.Next() {
		var (
			id   uuid.UUID
			stat trainingservice.ScenarioStats
		)
		if err := rows.Scan(&id, &stat.BestScore, &stat.AttemptsCount); err != nil {
			return nil, fmt.Errorf("reading scenario stats: %w", err)
		}
		stats[id] = stat
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("reading scenario stats: %w", err)
	}
	return stats, nil
}

// CreateAttempt создаёт новую попытку прохождения.
func (r *Repository) CreateAttempt(ctx context.Context, a models.Attempt) error {
	ctx, cancel := context.WithTimeout(ctx, r.pool.OpTimeout())
	defer cancel()

	state, err := json.Marshal(a.State)
	if err != nil {
		return fmt.Errorf("serializing state: %w", err)
	}

	const q = `
		INSERT INTO attempts (id, user_id, scenario_id, status, state, started_at)
		VALUES ($1, $2, $3, $4, $5, $6)`

	if _, err := r.pool.Exec(ctx, q, a.ID, a.UserID.UUID(), a.ScenarioID, a.Status, state, a.StartedAt); err != nil {
		return fmt.Errorf("creating attempt: %w", err)
	}
	return nil
}

// AttemptByID возвращает попытку по идентификатору.
func (r *Repository) AttemptByID(ctx context.Context, id uuid.UUID) (models.Attempt, error) {
	ctx, cancel := context.WithTimeout(ctx, r.pool.OpTimeout())
	defer cancel()

	const q = `
		SELECT id, user_id, scenario_id, status, state, score, outcome, started_at, finished_at, revision
		FROM attempts WHERE id = $1`

	a, err := scanAttempt(r.pool.QueryRow(ctx, q, id))
	if errors.Is(err, pgx.ErrNoRows) {
		return models.Attempt{}, fmt.Errorf("attempt %s: %w", id, domainErrors.ErrNotFound)
	}
	if err != nil {
		return models.Attempt{}, err
	}
	return a, nil
}

// ActiveAttempt возвращает незавершённую попытку пользователя по сценарию.
func (r *Repository) ActiveAttempt(
	ctx context.Context,
	userID models.UserID,
	scenarioID uuid.UUID,
) (models.Attempt, error) {
	ctx, cancel := context.WithTimeout(ctx, r.pool.OpTimeout())
	defer cancel()

	const q = `
		SELECT id, user_id, scenario_id, status, state, score, outcome, started_at, finished_at, revision
		FROM attempts
		WHERE user_id = $1 AND scenario_id = $2 AND status = 'in_progress'`

	a, err := scanAttempt(r.pool.QueryRow(ctx, q, userID.UUID(), scenarioID))
	if errors.Is(err, pgx.ErrNoRows) {
		return models.Attempt{}, domainErrors.ErrNotFound
	}
	if err != nil {
		return models.Attempt{}, err
	}
	return a, nil
}

// SaveStep фиксирует шаг и новую позицию попытки одной транзакцией.
func (r *Repository) SaveStep(
	ctx context.Context,
	step models.AttemptStep,
	a models.Attempt,
	userID models.UserID,
	expectedRevision int,
) (int, error) {
	ctx, cancel := context.WithTimeout(ctx, r.pool.OpTimeout())
	defer cancel()

	state, err := json.Marshal(a.State)
	if err != nil {
		return 0, fmt.Errorf("serializing state: %w", err)
	}

	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return 0, fmt.Errorf("beginning transaction: %w", err)
	}
	//nolint:errcheck
	defer tx.Rollback(ctx)

	const updateAttempt = `
		UPDATE attempts
		SET state = $4,
			revision = revision + 1
		WHERE id = $1
			AND user_id = $2
			AND revision = $3
			AND status = 'in_progress'
		RETURNING revision;
	`

	var newRevision int

	err = tx.QueryRow(
		ctx,
		updateAttempt,
		a.ID,
		userID,
		expectedRevision,
		state,
	).Scan(&newRevision)

	if errors.Is(err, pgx.ErrNoRows) {
		return 0, domainErrors.ErrConflict
	}
	if err != nil {
		return 0, fmt.Errorf("updating attempt: %w", err)
	}

	const insertStep = `
		INSERT INTO attempt_steps (attempt_id, scene_id, option_id, created_at)
		VALUES ($1, $2, $3, $4)`
	if _, err := tx.Exec(ctx, insertStep, step.AttemptID, step.SceneID, step.OptionID, step.CreatedAt); err != nil {
		return 0, fmt.Errorf("saving step: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return 0, fmt.Errorf("committing step: %w", err)
	}

	return newRevision, nil
}

// FinishAttempt проставляет итог попытки одним оператором.
func (r *Repository) FinishAttempt(
	ctx context.Context,
	id uuid.UUID,
	score int,
	outcome models.Outcome,
	at time.Time,
) error {
	ctx, cancel := context.WithTimeout(ctx, r.pool.OpTimeout())
	defer cancel()

	const q = `
		UPDATE attempts
		SET status = 'finished', score = $2, outcome = $3, finished_at = $4
		WHERE id = $1 AND status = 'in_progress'`

	tag, err := r.pool.Exec(ctx, q, id, score, outcome, at)
	if err != nil {
		return fmt.Errorf("ending attempt: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return fmt.Errorf("%w: attempt %s is not in progress", domainErrors.ErrAttemptFinished, id)
	}
	return nil
}

// Steps возвращает журнал шагов попытки в порядке прохождения.
func (r *Repository) Steps(ctx context.Context, attemptID uuid.UUID) ([]models.AttemptStep, error) {
	ctx, cancel := context.WithTimeout(ctx, r.pool.OpTimeout())
	defer cancel()

	const q = `
		SELECT attempt_id, scene_id, option_id, created_at
		FROM attempt_steps
		WHERE attempt_id = $1
		ORDER BY id`

	rows, err := r.pool.Query(ctx, q, attemptID)
	if err != nil {
		return nil, fmt.Errorf("listing steps: %w", err)
	}
	defer rows.Close()

	steps := make([]models.AttemptStep, 0, 8)
	for rows.Next() {
		var s models.AttemptStep
		if err := rows.Scan(&s.AttemptID, &s.SceneID, &s.OptionID, &s.CreatedAt); err != nil {
			return nil, fmt.Errorf("reading step: %w", err)
		}
		steps = append(steps, s)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("reading steps: %w", err)
	}
	return steps, nil
}

// BestScore возвращает лучший результат пользователя по сценарию.
func (r *Repository) BestScore(
	ctx context.Context,
	userID models.UserID,
	scenarioID, exclude uuid.UUID,
) (*int, error) {
	ctx, cancel := context.WithTimeout(ctx, r.pool.OpTimeout())
	defer cancel()

	const q = `
		SELECT MAX(score)
		FROM attempts
		WHERE user_id = $1 AND scenario_id = $2 AND id <> $3 AND status = 'finished'`

	var score *int
	if err := r.pool.QueryRow(ctx, q, userID.UUID(), scenarioID, exclude).Scan(&score); err != nil {
		return nil, fmt.Errorf("last score: %w", err)
	}
	return score, nil
}

type row interface {
	Scan(dest ...any) error
}

func scanScenario(r row) (models.Scenario, error) {
	var (
		s   models.Scenario
		doc []byte
	)
	if err := r.Scan(&s.ID, &doc, &s.IsActive); err != nil {
		return models.Scenario{}, err
	}
	if err := json.Unmarshal(doc, &s.Doc); err != nil {
		return models.Scenario{}, fmt.Errorf("unmarshaling scenario document %s: %w", s.ID, err)
	}
	return s, nil
}

func scanAttempt(r row) (models.Attempt, error) {
	var (
		a      models.Attempt
		state  []byte
		userID uuid.UUID
	)
	err := r.Scan(&a.ID, &userID, &a.ScenarioID, &a.Status, &state,
		&a.Score, &a.Outcome, &a.StartedAt, &a.FinishedAt, &a.Revision)
	if err != nil {
		return models.Attempt{}, err
	}
	a.UserID = models.UserID(userID)
	if err := json.Unmarshal(state, &a.State); err != nil {
		return models.Attempt{}, fmt.Errorf("unmarshaling attempt state %s: %w", a.ID, err)
	}
	return a, nil
}
