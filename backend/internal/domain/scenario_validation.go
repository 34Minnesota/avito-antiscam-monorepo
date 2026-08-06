package domain

import (
	"fmt"
	"math"
	"strings"
)

const (
	minScenarioDifficulty = 1
	maxScenarioDifficulty = 4
)

// Validate checks the structural invariants required to load and play a scenario.
func (d ScenarioDoc) Validate() error {
	if d.Version < 1 {
		return fmt.Errorf("version must be positive")
	}
	if strings.TrimSpace(d.Slug) == "" {
		return fmt.Errorf("slug must not be empty")
	}
	if !d.Role.IsValid() {
		return fmt.Errorf("unsupported role %q", d.Role)
	}
	if d.Difficulty < minScenarioDifficulty || d.Difficulty > maxScenarioDifficulty {
		return fmt.Errorf(
			"difficulty must be between %d and %d",
			minScenarioDifficulty,
			maxScenarioDifficulty,
		)
	}
	if len(d.Scenes) == 0 {
		return fmt.Errorf("at least one scene is required")
	}

	if err := d.validateEndings(); err != nil {
		return err
	}

	seenScenes := make(map[string]struct{}, len(d.Scenes))
	for sceneIndex, scene := range d.Scenes {
		if err := d.validateScene(scene, sceneIndex, seenScenes); err != nil {
			return err
		}
	}

	return nil
}

func (d ScenarioDoc) validateEndings() error {
	for _, required := range []string{"safe", "partial"} {
		if _, ok := d.Endings[required]; !ok {
			return fmt.Errorf("required ending %q is missing", required)
		}
	}

	for key, ending := range d.Endings {
		if strings.TrimSpace(key) == "" {
			return fmt.Errorf("ending key must not be empty")
		}
		if !ending.Outcome.isValid() {
			return fmt.Errorf("ending %q has unsupported outcome %q", key, ending.Outcome)
		}
	}

	return nil
}

func (d ScenarioDoc) validateScene(
	scene Scene,
	sceneIndex int,
	seenScenes map[string]struct{},
) error {
	if strings.TrimSpace(scene.ID) == "" {
		return fmt.Errorf("scene at index %d has an empty id", sceneIndex)
	}
	if _, duplicate := seenScenes[scene.ID]; duplicate {
		return fmt.Errorf("scene id %q is duplicated", scene.ID)
	}
	seenScenes[scene.ID] = struct{}{}

	if scene.Weight <= 0 || math.IsNaN(scene.Weight) || math.IsInf(scene.Weight, 0) {
		return fmt.Errorf("scene %q has invalid weight %v", scene.ID, scene.Weight)
	}
	if len(scene.Decision.Options) == 0 {
		return fmt.Errorf("scene %q has no options", scene.ID)
	}

	seenOptions := make(map[string]struct{}, len(scene.Decision.Options))
	for optionIndex, option := range scene.Decision.Options {
		if err := d.validateOption(scene.ID, option, optionIndex, seenOptions); err != nil {
			return err
		}
	}

	return nil
}

func (d ScenarioDoc) validateOption(
	sceneID string,
	option Option,
	optionIndex int,
	seenOptions map[string]struct{},
) error {
	if strings.TrimSpace(option.ID) == "" {
		return fmt.Errorf("option at index %d in scene %q has an empty id", optionIndex, sceneID)
	}
	if _, duplicate := seenOptions[option.ID]; duplicate {
		return fmt.Errorf("option id %q is duplicated in scene %q", option.ID, sceneID)
	}
	seenOptions[option.ID] = struct{}{}

	if !option.Verdict.isValid() {
		return fmt.Errorf(
			"option %q in scene %q has unsupported verdict %q",
			option.ID,
			sceneID,
			option.Verdict,
		)
	}
	if option.Verdict != VerdictFatal {
		return nil
	}
	if strings.TrimSpace(option.Ending) == "" {
		return fmt.Errorf("fatal option %q in scene %q has no ending", option.ID, sceneID)
	}
	if _, ok := d.Endings[option.Ending]; !ok {
		return fmt.Errorf(
			"fatal option %q in scene %q references missing ending %q",
			option.ID,
			sceneID,
			option.Ending,
		)
	}

	return nil
}

func (v Verdict) isValid() bool {
	return v == VerdictSafe || v == VerdictRisky || v == VerdictFatal
}

func (o Outcome) isValid() bool {
	return o == OutcomeSafe || o == OutcomePartial || o == OutcomeScammed
}
