package training

import (
	"context"
	"encoding/json"
	"fmt"
	"io/fs"
	"path"

	"github.com/34Minnesota/avito-antiscam-monorepo/backend/internal/domain"
	"github.com/google/uuid"
)

// Seed заливает сценарии из файловой системы в базу.
func Seed(ctx context.Context, repo Repository, files fs.FS, dir string) (int, error) {
	entries, err := fs.ReadDir(files, dir)
	if err != nil {
		return 0, fmt.Errorf("Reading scenario directory: %w", err)
	}

	loaded := 0
	for _, entry := range entries {
		if entry.IsDir() || path.Ext(entry.Name()) != ".json" {
			continue
		}

		raw, err := fs.ReadFile(files, path.Join(dir, entry.Name()))
		if err != nil {
			return loaded, fmt.Errorf("Reading %s: %w", entry.Name(), err)
		}

		var doc domain.ScenarioDoc
		if err := json.Unmarshal(raw, &doc); err != nil {
			return loaded, fmt.Errorf("Parsing %s: %w", entry.Name(), err)
		}

		scenario := domain.Scenario{
			ID:       uuid.New(),
			IsActive: true,
			Doc:      doc,
		}
		if err := repo.UpsertScenario(ctx, scenario); err != nil {
			return loaded, fmt.Errorf("Saving %s: %w", doc.Slug, err)
		}
		loaded++
	}

	return loaded, nil
}
