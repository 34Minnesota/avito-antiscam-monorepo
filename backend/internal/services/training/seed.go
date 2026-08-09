package training

import (
	"context"
	"encoding/json"
	"fmt"
	"io/fs"
	"path"

	"github.com/34Minnesota/avito-antiscam-monorepo/backend/internal/domain/models"
	"github.com/google/uuid"
)

// Seed заливает сценарии из файловой системы в базу.
func Seed(ctx context.Context, repo Repository, files fs.FS, dir string) (int, error) {
	entries, err := fs.ReadDir(files, dir)
	if err != nil {
		return 0, fmt.Errorf("reading scenario directory: %w", err)
	}

	loaded := 0
	for _, entry := range entries {
		if entry.IsDir() || path.Ext(entry.Name()) != ".json" {
			continue
		}

		raw, err := fs.ReadFile(files, path.Join(dir, entry.Name()))
		if err != nil {
			return loaded, fmt.Errorf("reading %s: %w", entry.Name(), err)
		}

		var doc models.ScenarioDoc
		if err := json.Unmarshal(raw, &doc); err != nil {
			return loaded, fmt.Errorf("parsing %s: %w", entry.Name(), err)
		}

		scenario := models.Scenario{
			ID:       uuid.New(),
			IsActive: true,
			Doc:      doc,
		}
		if err := repo.UpsertScenario(ctx, scenario); err != nil {
			return loaded, fmt.Errorf("saving %s: %w", doc.Slug, err)
		}
		loaded++
	}

	return loaded, nil
}
