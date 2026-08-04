package progress

import (
	"context"
	"database/sql"
	"os"
	"testing"
	"time"

	"github.com/google/uuid"
	_ "github.com/jackc/pgx/v5/stdlib"

	"github.com/34Minnesota/avito-antiscam-monorepo/backend/internal/domain"
)

func TestRepositoryLoadPostgres(t *testing.T) {
	dsn := os.Getenv("TEST_DATABASE_URL")
	if dsn == "" {
		t.Skip("TEST_DATABASE_URL is not set; PostgreSQL integration test was not run")
	}
	db, err := sql.Open("pgx", dsn)
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		if err := db.Close(); err != nil {
			t.Errorf("close database: %v", err)
		}
	}()
	db.SetMaxOpenConns(1)
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()
	if err := db.PingContext(ctx); err != nil {
		t.Fatal(err)
	}
	fixture := `
CREATE TEMP TABLE scenarios (id uuid PRIMARY KEY, slug text, title text, role text, current_version_id uuid);
CREATE TEMP TABLE scenario_versions (id uuid PRIMARY KEY, scenario_id uuid, version integer, max_score_points integer, pass_percent smallint, published_at timestamptz);
CREATE TEMP TABLE attempts (id uuid PRIMARY KEY, user_id uuid, scenario_version_id uuid, status text, score_points integer, max_score_points integer, started_at timestamptz, completed_at timestamptz);
`
	mustExec(t, db, ctx, fixture)
	userUUID, scenarioID, versionID := uuid.New(), uuid.New(), uuid.New()
	now := time.Now().UTC().Truncate(time.Microsecond)
	mustExec(t, db, ctx, `INSERT INTO scenarios VALUES ($1, 'buyer-test', 'Buyer test', 'buyer', $2)`, scenarioID, versionID)
	mustExec(t, db, ctx, `INSERT INTO scenario_versions VALUES ($1, $2, 1, 100, 70, $3)`, versionID, scenarioID, now)
	mustExec(t, db, ctx, `INSERT INTO attempts VALUES
($1, $2, $3, 'completed', 69, 100, $4, $5),
($6, $2, $3, 'in_progress', 10, 100, $4, NULL)`, uuid.New(), userUUID, versionID, now.Add(-time.Hour), now, uuid.New())
	userID, err := domain.NewUserID(userUUID)
	if err != nil {
		t.Fatal(err)
	}
	snapshot, err := New(db).Load(ctx, userID, 20)
	if err != nil {
		t.Fatal(err)
	}
	if len(snapshot.Scenarios) != 1 {
		t.Fatalf("scenarios = %d", len(snapshot.Scenarios))
	}
	current := snapshot.Scenarios[0].Current
	if current.AttemptsCount != 2 || !current.Completed || current.Passed || current.ActiveAttemptID == nil {
		t.Fatalf("unexpected current progress: %+v", current)
	}
	if len(snapshot.Scenarios[0].RecentAttempts) != 1 || snapshot.Scenarios[0].RecentAttempts[0].Passed {
		t.Fatalf("unexpected history: %+v", snapshot.Scenarios[0].RecentAttempts)
	}
}

func mustExec(t *testing.T, db *sql.DB, ctx context.Context, query string, args ...any) {
	t.Helper()
	if _, err := db.ExecContext(ctx, query, args...); err != nil {
		t.Fatal(err)
	}
}
