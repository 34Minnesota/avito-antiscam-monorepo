package training_test

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"testing"
	"testing/fstest"

	"github.com/34Minnesota/avito-antiscam-monorepo/backend/internal/domain"
	domainErrors "github.com/34Minnesota/avito-antiscam-monorepo/backend/internal/domain/errors"
	"github.com/34Minnesota/avito-antiscam-monorepo/backend/internal/services/training"
)

func mustJSON(t *testing.T, slug string) []byte {
	t.Helper()
	doc := testDoc()
	doc.Slug = slug
	doc.Debrief.KeyFlags = append(doc.Debrief.KeyFlags, domain.FlagInfo{ID: "ghost"})
	raw, err := json.Marshal(doc)
	if err != nil {
		t.Fatal(err)
	}
	return raw
}

func TestSeedLoadsOnlyJSONFiles(t *testing.T) {
	t.Parallel()
	files := fstest.MapFS{
		"scenarios/first.json":       {Data: mustJSON(t, "first")},
		"scenarios/second.json":      {Data: mustJSON(t, "second")},
		"scenarios/readme.md":        {Data: []byte("# not a scenario")},
		"scenarios/nested/deep.json": {Data: mustJSON(t, "deep")},
	}
	repo := &repositoryStub{}

	loaded, err := training.Seed(context.Background(), repo, files, "scenarios")
	if err != nil {
		t.Fatal(err)
	}
	if loaded != 2 || len(repo.upserted) != 2 {
		t.Fatalf("loaded = %d, upserted = %d", loaded, len(repo.upserted))
	}
	for _, s := range repo.upserted {
		if !s.IsActive || s.ID.String() == "" {
			t.Fatalf("scenario must be active and identified: %+v", s)
		}
	}
}

func TestSeedAcceptsBundledScenarios(t *testing.T) {
	t.Parallel()
	repo := &repositoryStub{}

	loaded, err := training.Seed(
		context.Background(),
		repo,
		os.DirFS("../../../docs/scenarios"),
		".",
	)
	if err != nil {
		t.Fatal(err)
	}
	if loaded == 0 || loaded != len(repo.upserted) {
		t.Fatalf("loaded = %d, upserted = %d", loaded, len(repo.upserted))
	}
}

func TestSeedRejectsMissingDirectory(t *testing.T) {
	t.Parallel()
	if _, err := training.Seed(context.Background(), &repositoryStub{}, fstest.MapFS{}, "nope"); err == nil {
		t.Fatal("expected error")
	}
}

func TestSeedRejectsBrokenJSON(t *testing.T) {
	t.Parallel()
	files := fstest.MapFS{"scenarios/broken.json": {Data: []byte("{ not json")}}

	loaded, err := training.Seed(context.Background(), &repositoryStub{}, files, "scenarios")
	if err == nil {
		t.Fatal("expected error")
	}
	if loaded != 0 {
		t.Fatalf("loaded = %d", loaded)
	}
}

func TestSeedRejectsInvalidScenarioBeforeSaving(t *testing.T) {
	t.Parallel()
	doc := testDoc()
	doc.Scenes[0].Decision.Options = append(
		doc.Scenes[0].Decision.Options,
		domain.Option{ID: "broken", Verdict: domain.VerdictFatal, Ending: "missing"},
	)
	raw, err := json.Marshal(doc)
	if err != nil {
		t.Fatal(err)
	}
	files := fstest.MapFS{"scenarios/invalid.json": {Data: raw}}
	repo := &repositoryStub{}

	loaded, err := training.Seed(context.Background(), repo, files, "scenarios")
	if !errors.Is(err, domainErrors.ErrInvalidScenario) {
		t.Fatalf("got %v", err)
	}
	if loaded != 0 || len(repo.upserted) != 0 {
		t.Fatalf("invalid scenario was saved: loaded=%d, upserted=%d", loaded, len(repo.upserted))
	}
}

func TestSeedStopsOnRepositoryError(t *testing.T) {
	t.Parallel()
	files := fstest.MapFS{"scenarios/first.json": {Data: mustJSON(t, "first")}}

	_, err := training.Seed(context.Background(), &repositoryStub{upsertErr: errBoom}, files, "scenarios")
	if !errors.Is(err, errBoom) {
		t.Fatalf("got %v", err)
	}
}

func TestSeedEmptyDirectoryLoadsNothing(t *testing.T) {
	t.Parallel()
	files := fstest.MapFS{"scenarios/readme.md": {Data: []byte("empty")}}

	loaded, err := training.Seed(context.Background(), &repositoryStub{}, files, "scenarios")
	if err != nil {
		t.Fatal(err)
	}
	if loaded != 0 {
		t.Fatalf("loaded = %d", loaded)
	}
}
