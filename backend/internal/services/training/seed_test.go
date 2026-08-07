package training_test

import (
	"context"
	"encoding/json"
	"errors"
	"testing"
	"testing/fstest"

	"github.com/34Minnesota/avito-antiscam-monorepo/backend/internal/services/training"
)

func mustJSON(t *testing.T, slug string) []byte {
	t.Helper()
	doc := testDoc()
	doc.Slug = slug
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
