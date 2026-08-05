package progress

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/34Minnesota/avito-antiscam-monorepo/backend/internal/domain"
	progressservice "github.com/34Minnesota/avito-antiscam-monorepo/backend/internal/services/progress"
	"github.com/34Minnesota/avito-antiscam-monorepo/backend/internal/storage/postgres/pool"
)

// TODO(owner: scenario): provide immutable scenario_versions.pass_percent (1..100).
// TODO(owner: training-engine): persist max_score_points snapshots and attempt statuses.

// тут навайбкожено т.к. нету общего пула постгреса, в другой ветке добавлю и смержу

type Repository struct {
	pool *pool.Pool
}

func New(pool *pool.Pool) *Repository {
	return &Repository{
		pool: pool,
	}
}

func (r *Repository) Load(
	ctx context.Context,
	userID domain.UserID,
	historyLimit int,
) (domain.ProgressSnapshot, error) {
	if r == nil || r.pool == nil {
		return domain.ProgressSnapshot{}, fmt.Errorf(
			"%w: postgres is not configured",
			progressservice.ErrDependencyUnavailable,
		)
	}

	ctx, cancel := context.WithTimeout(ctx, r.pool.OpTimeout())
	defer cancel()

	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return domain.ProgressSnapshot{}, dependencyError(err)
	}
	defer func() {
		_ = tx.Rollback(ctx) //nolint:errcheck
	}()

	scenarios, indexes, err := loadVersions(ctx, tx, userID)
	if err != nil {
		return domain.ProgressSnapshot{}, err
	}

	if err := loadRecentAttempts(ctx, tx, userID, historyLimit, scenarios, indexes); err != nil {
		return domain.ProgressSnapshot{}, err
	}

	outdated, err := loadOutdatedAttempts(ctx, tx, userID)
	if err != nil {
		return domain.ProgressSnapshot{}, err
	}

	return domain.ProgressSnapshot{Scenarios: scenarios, OutdatedActiveAttempts: outdated}, nil
}

type scenarioIndexes struct {
	byID     map[uuid.UUID]int
	versions map[uuid.UUID]domain.Version
}

const versionsQuery = `
	SELECT s.id, s.slug, s.title, s.role,
		cv.id, cv.version, cv.max_score_points, cv.pass_percent, cv.published_at,
		v.id, v.version, v.max_score_points, v.pass_percent, v.published_at,
		COUNT(a.id), COUNT(a.id) FILTER (WHERE a.status = 'completed'),
		MAX(a.score_points) FILTER (WHERE a.status = 'completed'),
		COUNT(a.id) FILTER (WHERE a.status = 'in_progress'),
		(MIN(a.id::text) FILTER (WHERE a.status = 'in_progress'))::uuid,
		COUNT(a.id) FILTER (WHERE a.max_score_points <> v.max_score_points),
		COALESCE(BOOL_OR(a.status = 'completed' AND
			a.score_points::bigint * 100 >= a.max_score_points::bigint * v.pass_percent), false)
	FROM scenarios s
	JOIN scenario_versions cv ON cv.id = s.current_version_id AND cv.scenario_id = s.id
	JOIN scenario_versions v ON v.scenario_id = s.id
	LEFT JOIN attempts a ON a.scenario_version_id = v.id AND a.user_id = $1
	GROUP BY s.id, s.slug, s.title, s.role,
			cv.id, cv.version, cv.max_score_points, cv.pass_percent, cv.published_at,
			v.id, v.version, v.max_score_points, v.pass_percent, v.published_at
	ORDER BY s.slug, v.version DESC
	`

func loadVersions(ctx context.Context, tx pgx.Tx, userID domain.UserID) ([]domain.ScenarioProgress, scenarioIndexes, error) {
	rows, err := tx.Query(ctx, versionsQuery, userID.UUID())
	if err != nil {
		return nil, scenarioIndexes{}, dependencyError(err)
	}
	defer rows.Close()

	indexes := scenarioIndexes{
		byID:     make(map[uuid.UUID]int),
		versions: make(map[uuid.UUID]domain.Version),
	}

	var scenarios []domain.ScenarioProgress
	for rows.Next() {
		var scenarioID uuid.UUID
		var slug, title, role string
		var current, version domain.Version
		var attempts, completed, active, mismatched int
		var best sql.NullInt64
		var activeID uuid.NullUUID
		var passed bool

		if err := scanVersionRow(rows, &scenarioID, &slug, &title, &role, &current, &version,
			&attempts, &completed, &best, &active, &activeID, &mismatched, &passed); err != nil {
			return nil, scenarioIndexes{}, dependencyError(err)
		}

		if invalidAttemptState(mismatched, version.ID == current.ID, active) {
			return nil, scenarioIndexes{}, inconsistentError("attempt snapshot or active-attempt invariant is broken")
		}

		versionProgress, err := makeVersionProgress(version, attempts, completed, passed, best, activeID)
		if err != nil {
			return nil, scenarioIndexes{}, err
		}

		index, exists := indexes.byID[scenarioID]
		if !exists {
			index = len(scenarios)
			indexes.byID[scenarioID] = index
			scenarios = append(scenarios, domain.ScenarioProgress{ID: scenarioID, Slug: slug, Title: title, Role: domain.ScenarioRole(role)})
		}

		indexes.versions[version.ID] = version

		if version.ID == current.ID {
			scenarios[index].Current = versionProgress
		} else {
			versionProgress.ActiveAttemptID = nil
			scenarios[index].History = append(scenarios[index].History, versionProgress)
		}
	}

	if err := rows.Err(); err != nil {
		return nil, scenarioIndexes{}, dependencyError(err)
	}

	return scenarios, indexes, nil
}

func invalidAttemptState(mismatched int, current bool, active int) bool {
	return mismatched > 0 || current && active > 1
}

func scanVersionRow(rows pgx.Rows, scenarioID *uuid.UUID, slug, title, role *string,
	current, version *domain.Version, attempts, completed *int, best *sql.NullInt64,
	active *int, activeID *uuid.NullUUID, mismatched *int, passed *bool,
) error {
	return rows.Scan(scenarioID, slug, title, role,
		&current.ID, &current.Number, &current.MaxPoints, &current.PassPercent, &current.PublishedAt,
		&version.ID, &version.Number, &version.MaxPoints, &version.PassPercent, &version.PublishedAt,
		attempts, completed, best, active, activeID, mismatched, passed)
}

func makeVersionProgress(version domain.Version, attempts, completed int, passed bool, best sql.NullInt64, activeID uuid.NullUUID) (domain.VersionProgress, error) {
	result := domain.VersionProgress{Version: version, AttemptsCount: attempts, Completed: completed > 0, Passed: passed}
	if best.Valid {
		score, err := domain.NewScore(int(best.Int64), version.MaxPoints)
		if err != nil {
			return domain.VersionProgress{}, inconsistentError(err.Error())
		}
		result.BestScore = &score
	}
	if activeID.Valid {
		id := activeID.UUID
		result.ActiveAttemptID = &id
	}
	return result, nil
}

const recentAttemptsQuery = `
	WITH ranked AS (
	SELECT s.id AS scenario_id, a.id, a.scenario_version_id, a.score_points,
			a.max_score_points, a.completed_at,
			ROW_NUMBER() OVER (PARTITION BY s.id ORDER BY a.completed_at DESC, a.id DESC) AS position
	FROM scenarios s
	JOIN scenario_versions v ON v.scenario_id = s.id
	JOIN attempts a ON a.scenario_version_id = v.id
	WHERE a.user_id = $1 AND a.status = 'completed' AND a.completed_at IS NOT NULL
	)
	SELECT scenario_id, id, scenario_version_id, score_points, max_score_points, completed_at
	FROM ranked WHERE position <= $2
	ORDER BY scenario_id, completed_at DESC, id DESC
	`

func loadRecentAttempts(ctx context.Context, tx pgx.Tx, userID domain.UserID, limit int, scenarios []domain.ScenarioProgress, indexes scenarioIndexes) error {
	rows, err := tx.Query(ctx, recentAttemptsQuery, userID.UUID(), limit)
	if err != nil {
		return dependencyError(err)
	}
	defer rows.Close()

	for rows.Next() {
		var scenarioID, attemptID, versionID uuid.UUID
		var points, maxPoints int
		var completedAt sql.NullTime

		if err := rows.Scan(&scenarioID, &attemptID, &versionID, &points, &maxPoints, &completedAt); err != nil {
			return dependencyError(err)
		}

		index, ok := indexes.byID[scenarioID]
		version, versionOK := indexes.versions[versionID]
		if !ok || !versionOK || !completedAt.Valid || maxPoints != version.MaxPoints {
			return inconsistentError("attempt references unknown or inconsistent scenario version")
		}

		score, err := domain.NewScore(points, maxPoints)
		if err != nil {
			return inconsistentError(err.Error())
		}

		passed := int64(points)*100 >= int64(maxPoints)*int64(version.PassPercent)

		scenarios[index].RecentAttempts = append(scenarios[index].RecentAttempts, domain.AttemptResult{
			ID: attemptID, Version: version, Score: score, Passed: passed, CompletedAt: completedAt.Time,
		})
	}

	if err := rows.Err(); err != nil {
		return dependencyError(err)
	}
	return nil
}

const outdatedAttemptsQuery = `
	SELECT a.id, s.slug, s.title,
		v.id, v.version, v.max_score_points, v.pass_percent, v.published_at,
		cv.id, cv.version, cv.max_score_points, cv.pass_percent, cv.published_at
	FROM attempts a
	JOIN scenario_versions v ON v.id = a.scenario_version_id
	JOIN scenarios s ON s.id = v.scenario_id
	JOIN scenario_versions cv ON cv.id = s.current_version_id
	WHERE a.user_id = $1 AND a.status = 'in_progress' AND v.id <> cv.id
	ORDER BY a.started_at DESC, a.id DESC
	`

func loadOutdatedAttempts(ctx context.Context, tx pgx.Tx, userID domain.UserID) ([]domain.OutdatedActiveAttempt, error) {
	rows, err := tx.Query(ctx, outdatedAttemptsQuery, userID.UUID())
	if err != nil {
		return nil, dependencyError(err)
	}
	defer rows.Close()

	var result []domain.OutdatedActiveAttempt
	for rows.Next() {
		var item domain.OutdatedActiveAttempt
		if err := rows.Scan(&item.AttemptID, &item.ScenarioSlug, &item.ScenarioTitle,
			&item.Version.ID, &item.Version.Number, &item.Version.MaxPoints, &item.Version.PassPercent, &item.Version.PublishedAt,
			&item.CurrentVersion.ID, &item.CurrentVersion.Number, &item.CurrentVersion.MaxPoints, &item.CurrentVersion.PassPercent, &item.CurrentVersion.PublishedAt); err != nil {
			return nil, dependencyError(err)
		}

		result = append(result, item)
	}

	if err := rows.Err(); err != nil {
		return nil, dependencyError(err)
	}

	return result, nil
}

func dependencyError(err error) error {
	return fmt.Errorf("%w: %w", progressservice.ErrDependencyUnavailable, err)
}

func inconsistentError(message string) error {
	return fmt.Errorf("%w: %s", progressservice.ErrDataInconsistent, message)
}
