package training

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/34Minnesota/avito-antiscam-monorepo/backend/internal/domain"
	domainErrors "github.com/34Minnesota/avito-antiscam-monorepo/backend/internal/domain/errors"
	trainingservice "github.com/34Minnesota/avito-antiscam-monorepo/backend/internal/services/training"
	postgrespool "github.com/34Minnesota/avito-antiscam-monorepo/backend/internal/storage/postgres/pool"
)

type Repository struct {
	pool *postgrespool.Pool
}

func New(pool *postgrespool.Pool) *Repository {
	return &Repository{pool: pool}
}

// UpsertScenario вставляет или обновляет сценарий
func (r *Repository) UpsertScenario(ctx context.Context, s domain.Scenario) error {
	ctx, cancel := context.WithTimeout(ctx, r.pool.OpTimeout())
	defer cancel()

	doc, err := json.Marshal(s.Doc)
	if err != nil {
		return fmt.Errorf("scenario serializing %s: %w", s.Doc.Slug, err)
	}

	const q = `
		INSERT INTO scenarios (id, doc)
		VALUES ($1, $2)
		ON CONFLICT (slug) DO UPDATE SET
			doc        = EXCLUDED.doc,
			is_active  = true,
			updated_at = now()`

	if _, err := r.pool.Exec(ctx, q, s.ID, doc); err != nil {
		return fmt.Errorf("upsert scenario %s: %w", s.Doc.Slug, err)
	}
	return nil
}

// ListScenarios возвращает активные сценарии, опционально фильтруя по роли.
func (r *Repository) ListScenarios(ctx context.Context, role domain.Role) ([]domain.Scenario, error) {
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

	scenarios := make([]domain.Scenario, 0, 8)
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
func (r *Repository) ScenarioByID(ctx context.Context, id uuid.UUID) (domain.Scenario, error) {
	ctx, cancel := context.WithTimeout(ctx, r.pool.OpTimeout())
	defer cancel()

	const q = `SELECT id, doc, is_active FROM scenarios WHERE id = $1 AND is_active`

	s, err := scanScenario(r.pool.QueryRow(ctx, q, id))
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.Scenario{}, fmt.Errorf("scenario %s: %w", id, domainErrors.ErrNotFound)
	}
	if err != nil {
		return domain.Scenario{}, err
	}
	return s, nil
}

// ScenarioStats возвращает статистику пользователя по каждому сценарию.
func (r *Repository) ScenarioStats(
	ctx context.Context,
	userID domain.UserID,
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
func (r *Repository) CreateAttempt(ctx context.Context, a domain.Attempt) error {
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
func (r *Repository) AttemptByID(ctx context.Context, id uuid.UUID) (domain.Attempt, error) {
	ctx, cancel := context.WithTimeout(ctx, r.pool.OpTimeout())
	defer cancel()

	const q = `
		SELECT id, user_id, scenario_id, status, state, score, outcome, started_at, finished_at
		FROM attempts WHERE id = $1`

	a, err := scanAttempt(r.pool.QueryRow(ctx, q, id))
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.Attempt{}, fmt.Errorf("attempt %s: %w", id, domainErrors.ErrNotFound)
	}
	if err != nil {
		return domain.Attempt{}, err
	}
	return a, nil
}

// ActiveAttempt возвращает незавершённую попытку пользователя по сценарию.
func (r *Repository) ActiveAttempt(
	ctx context.Context,
	userID domain.UserID,
	scenarioID uuid.UUID,
) (domain.Attempt, error) {
	ctx, cancel := context.WithTimeout(ctx, r.pool.OpTimeout())
	defer cancel()

	const q = `
		SELECT id, user_id, scenario_id, status, state, score, outcome, started_at, finished_at
		FROM attempts
		WHERE user_id = $1 AND scenario_id = $2 AND status = 'in_progress'`

	a, err := scanAttempt(r.pool.QueryRow(ctx, q, userID.UUID(), scenarioID))
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.Attempt{}, domainErrors.ErrNotFound
	}
	if err != nil {
		return domain.Attempt{}, err
	}
	return a, nil
}

// SaveStep фиксирует шаг и новую позицию попытки одной транзакцией.
func (r *Repository) SaveStep(ctx context.Context, step domain.AttemptStep, a domain.Attempt) error {
	ctx, cancel := context.WithTimeout(ctx, r.pool.OpTimeout())
	defer cancel()

	state, err := json.Marshal(a.State)
	if err != nil {
		return fmt.Errorf("serializing state: %w", err)
	}

	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("beginning transaction: %w", err)
	}
	//nolint:errcheck
	defer tx.Rollback(ctx)

	const insertStep = `
		INSERT INTO attempt_steps (attempt_id, scene_id, option_id, created_at)
		VALUES ($1, $2, $3, $4)`
	if _, err := tx.Exec(ctx, insertStep, step.AttemptID, step.SceneID, step.OptionID, step.CreatedAt); err != nil {
		return fmt.Errorf("saving step: %w", err)
	}

	const updateAttempt = `UPDATE attempts SET state = $2 WHERE id = $1`
	if _, err := tx.Exec(ctx, updateAttempt, a.ID, state); err != nil {
		return fmt.Errorf("updating attempt: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("committing step: %w", err)
	}
	return nil
}

// FinishAttempt проставляет итог попытки одним оператором.
func (r *Repository) FinishAttempt(
	ctx context.Context,
	id uuid.UUID,
	score int,
	outcome domain.Outcome,
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
func (r *Repository) Steps(ctx context.Context, attemptID uuid.UUID) ([]domain.AttemptStep, error) {
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

	steps := make([]domain.AttemptStep, 0, 8)
	for rows.Next() {
		var s domain.AttemptStep
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
	userID domain.UserID,
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

func scanScenario(r row) (domain.Scenario, error) {
	var (
		s   domain.Scenario
		doc []byte
	)
	if err := r.Scan(&s.ID, &doc, &s.IsActive); err != nil {
		return domain.Scenario{}, err
	}
	if err := json.Unmarshal(doc, &s.Doc); err != nil {
		return domain.Scenario{}, fmt.Errorf("unmarshaling scenario document %s: %w", s.ID, err)
	}
	return s, nil
}

func scanAttempt(r row) (domain.Attempt, error) {
	var (
		a      domain.Attempt
		state  []byte
		userID uuid.UUID
	)
	err := r.Scan(&a.ID, &userID, &a.ScenarioID, &a.Status, &state,
		&a.Score, &a.Outcome, &a.StartedAt, &a.FinishedAt)
	if err != nil {
		return domain.Attempt{}, err
	}
	a.UserID = domain.UserID(userID)
	if err := json.Unmarshal(state, &a.State); err != nil {
		return domain.Attempt{}, fmt.Errorf("unmarshaling attempt state %s: %w", a.ID, err)
	}
	return a, nil
}
