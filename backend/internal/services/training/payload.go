package training

import "github.com/34Minnesota/avito-antiscam-monorepo/backend/internal/domain/models"

func BuildScene(s models.Scene) ScenePayload {
	opts := make([]OptionPayload, 0, len(s.Decision.Options))
	for _, o := range s.Decision.Options {
		opts = append(opts, OptionPayload{
			ID:   o.ID,
			Text: o.Text,
		})
	}

	return ScenePayload{
		SceneID: s.ID,
		Intro:   s.Intro,
		Prompt:  s.Decision.Prompt,
		Options: opts,
	}
}
