package domain_test

import (
	"testing"

	"github.com/34Minnesota/avito-antiscam-monorepo/backend/internal/domain"
)

func validScenarioDoc() domain.ScenarioDoc {
	return domain.ScenarioDoc{
		Version:    1,
		Slug:       "valid-scenario",
		Role:       domain.RoleSeller,
		Difficulty: 2,
		Scenes: []domain.Scene{{
			ID:     "scene-1",
			Weight: 1,
			Decision: domain.Decision{Options: []domain.Option{
				{ID: "safe", Verdict: domain.VerdictSafe},
				{ID: "risky", Verdict: domain.VerdictRisky, Flag: "payment-outside"},
				{ID: "fatal", Verdict: domain.VerdictFatal, Ending: "lost"},
			}},
		}},
		Endings: map[string]domain.Ending{
			"safe":    {Outcome: domain.OutcomeSafe},
			"partial": {Outcome: domain.OutcomePartial},
			"lost":    {Outcome: domain.OutcomeScammed},
		},
		Debrief: domain.Debrief{KeyFlags: []domain.FlagInfo{{ID: "payment-outside"}}},
	}
}

func TestScenarioDocValidateAcceptsPlayableDocument(t *testing.T) {
	t.Parallel()
	if err := validScenarioDoc().Validate(); err != nil {
		t.Fatal(err)
	}
}

func TestScenarioDocValidateRejectsBrokenStructure(t *testing.T) {
	t.Parallel()

	tests := map[string]func(*domain.ScenarioDoc){
		"missing slug": func(doc *domain.ScenarioDoc) {
			doc.Slug = ""
		},
		"unsupported role": func(doc *domain.ScenarioDoc) {
			doc.Role = "moderator"
		},
		"difficulty outside database constraint": func(doc *domain.ScenarioDoc) {
			doc.Difficulty = 5
		},
		"no scenes": func(doc *domain.ScenarioDoc) {
			doc.Scenes = nil
		},
		"duplicate scene id": func(doc *domain.ScenarioDoc) {
			doc.Scenes = append(doc.Scenes, doc.Scenes[0])
		},
		"non-positive weight": func(doc *domain.ScenarioDoc) {
			doc.Scenes[0].Weight = 0
		},
		"no options": func(doc *domain.ScenarioDoc) {
			doc.Scenes[0].Decision.Options = nil
		},
		"duplicate option id": func(doc *domain.ScenarioDoc) {
			doc.Scenes[0].Decision.Options[1].ID = "safe"
		},
		"duplicate option id across scenes": func(doc *domain.ScenarioDoc) {
			doc.Scenes = append(doc.Scenes, domain.Scene{
				ID:     "scene-2",
				Weight: 1,
				Decision: domain.Decision{Options: []domain.Option{{
					ID:      "safe",
					Verdict: domain.VerdictSafe,
				}}},
			})
		},
		"unsupported verdict": func(doc *domain.ScenarioDoc) {
			doc.Scenes[0].Decision.Options[0].Verdict = "unknown"
		},
		"empty key flag id": func(doc *domain.ScenarioDoc) {
			doc.Debrief.KeyFlags[0].ID = " "
		},
		"duplicate key flag id": func(doc *domain.ScenarioDoc) {
			doc.Debrief.KeyFlags = append(doc.Debrief.KeyFlags, doc.Debrief.KeyFlags[0])
		},
		"option references missing flag": func(doc *domain.ScenarioDoc) {
			doc.Scenes[0].Decision.Options[1].Flag = "missing"
		},
		"option has blank flag id": func(doc *domain.ScenarioDoc) {
			doc.Scenes[0].Decision.Options[1].Flag = " "
		},
		"missing required ending": func(doc *domain.ScenarioDoc) {
			delete(doc.Endings, "safe")
		},
		"fatal option without ending": func(doc *domain.ScenarioDoc) {
			doc.Scenes[0].Decision.Options[2].Ending = ""
		},
		"fatal option references missing ending": func(doc *domain.ScenarioDoc) {
			doc.Scenes[0].Decision.Options[2].Ending = "missing"
		},
	}

	for name, mutate := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			doc := validScenarioDoc()
			mutate(&doc)
			if err := doc.Validate(); err == nil {
				t.Fatal("expected validation error")
			}
		})
	}
}
