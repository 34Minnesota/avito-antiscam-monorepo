package progress

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/34Minnesota/avito-antiscam-monorepo/backend/internal/domain"
	domainErrors "github.com/34Minnesota/avito-antiscam-monorepo/backend/internal/domain/errors"
	"github.com/34Minnesota/avito-antiscam-monorepo/backend/internal/storage/postgres/pool"
)

type Repository struct {
	pool *pool.Pool
}

func New(pool *pool.Pool) *Repository {
	return &Repository{pool: pool}
}

func (r *Repository) Load(ctx context.Context, userID domain.UserID, historyLimit int) (domain.ProgressSnapshot, error) {
	if r == nil || r.pool == nil {
		return domain.ProgressSnapshot{}, fmt.Errorf("%w: postgres is not configured", domainErrors.ErrDependencyUnavailable)
	}

	ctx, cancel := context.WithTimeout(ctx, r.pool.OpTimeout())
	defer cancel()

	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return domain.ProgressSnapshot{}, dependencyError(err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	scenarios, indexes, err := loadScenarios(ctx, tx, userID)
	if err != nil {
		return domain.ProgressSnapshot{}, err
	}
	if err := loadRecentAttempts(ctx, tx, userID, historyLimit, scenarios, indexes); err != nil {
		return domain.ProgressSnapshot{}, err
	}

	return domain.ProgressSnapshot{Scenarios: scenarios}, nil
}

const scenariosQuery = `
	SELECT s.id, s.slug, s.title, s.role,
		COUNT(a.id),
		COUNT(a.id) FILTER (WHERE a.status = 'finished'),
		COALESCE(BOOL_OR(a.status = 'finished' AND a.outcome = 'safe'), false),
		MAX(a.score) FILTER (WHERE a.status = 'finished'),
		COUNT(a.id) FILTER (WHERE a.status = 'in_progress'),
		(MIN(a.id::text) FILTER (WHERE a.status = 'in_progress'))::uuid
	FROM scenarios s
	LEFT JOIN attempts a ON a.scenario_id = s.id AND a.user_id = $1
	WHERE s.is_active
	GROUP BY s.id, s.slug, s.title, s.role
	ORDER BY s.slug`

func loadScenarios(ctx context.Context, tx pgx.Tx, userID domain.UserID) ([]domain.ScenarioProgress, map[uuid.UUID]int, error) {
	rows, err := tx.Query(ctx, scenariosQuery, userID.UUID())
	if err != nil {
		return nil, nil, dependencyError(err)
	}
	defer rows.Close()

	scenarios := make([]domain.ScenarioProgress, 0)
	indexes := make(map[uuid.UUID]int)
	for rows.Next() {
		var scenario domain.ScenarioProgress
		var completed, active int
		var best sql.NullInt64
		var activeID uuid.NullUUID
		if err := rows.Scan(&scenario.ID, &scenario.Slug, &scenario.Title, &scenario.Role,
			&scenario.AttemptsCount, &completed, &scenario.Passed, &best, &active, &activeID); err != nil {
			return nil, nil, dependencyError(err)
		}
		if active > 1 || scenario.Passed && completed == 0 {
			return nil, nil, inconsistentError("attempt snapshot violates progress invariants")
		}
		scenario.Completed = completed > 0
		if best.Valid {
			score, err := domain.NewScore(int(best.Int64), 100)
			if err != nil {
				return nil, nil, inconsistentError(err.Error())
			}
			scenario.BestScore = &score
		}
		if activeID.Valid {
			id := activeID.UUID
			scenario.ActiveAttemptID = &id
		}
		indexes[scenario.ID] = len(scenarios)
		scenarios = append(scenarios, scenario)
	}
	if err := rows.Err(); err != nil {
		return nil, nil, dependencyError(err)
	}
	return scenarios, indexes, nil
}

const recentAttemptsQuery = `
	WITH ranked AS (
		SELECT a.scenario_id, a.id, a.score, a.outcome, a.finished_at,
			ROW_NUMBER() OVER (PARTITION BY a.scenario_id ORDER BY a.finished_at DESC, a.id DESC) AS position
		FROM attempts a
		JOIN scenarios s ON s.id = a.scenario_id
		WHERE a.user_id = $1 AND a.status = 'finished' AND a.finished_at IS NOT NULL AND s.is_active
	)
	SELECT scenario_id, id, score, outcome, finished_at
	FROM ranked
	WHERE position <= $2
	ORDER BY scenario_id, finished_at DESC, id DESC`

func loadRecentAttempts(ctx context.Context, tx pgx.Tx, userID domain.UserID, limit int, scenarios []domain.ScenarioProgress, indexes map[uuid.UUID]int) error {
	rows, err := tx.Query(ctx, recentAttemptsQuery, userID.UUID(), limit)
	if err != nil {
		return dependencyError(err)
	}
	defer rows.Close()

	for rows.Next() {
		var scenarioID uuid.UUID
		var attempt domain.AttemptResult
		var points sql.NullInt64
		var completedAt sql.NullTime
		if err := rows.Scan(&scenarioID, &attempt.ID, &points, &attempt.Outcome, &completedAt); err != nil {
			return dependencyError(err)
		}
		index, ok := indexes[scenarioID]
		if !ok || !points.Valid || !completedAt.Valid {
			return inconsistentError("finished attempt has an invalid progress snapshot")
		}
		score, err := domain.NewScore(int(points.Int64), 100)
		if err != nil {
			return inconsistentError(err.Error())
		}
		attempt.Score = score
		attempt.CompletedAt = completedAt.Time
		scenarios[index].RecentAttempts = append(scenarios[index].RecentAttempts, attempt)
	}
	if err := rows.Err(); err != nil {
		return dependencyError(err)
	}
	return nil
}

func dependencyError(err error) error {
	return fmt.Errorf("%w: %w", domainErrors.ErrDependencyUnavailable, err)
}

func inconsistentError(message string) error {
	return fmt.Errorf("%w: %s", domainErrors.ErrDataInconsistent, message)
}
