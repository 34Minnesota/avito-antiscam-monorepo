package training

import (
	"fmt"
	"math"

	"github.com/34Minnesota/avito-antiscam-monorepo/backend/internal/domain"
	domainErrors "github.com/34Minnesota/avito-antiscam-monorepo/backend/internal/domain/errors"
)

// Подсчёт итога прохождения.

// stepEffect — вклад одного шага журнала в итог.
type stepEffect struct {
	weight    float64
	flag      string
	endingKey string
}

// Evaluate восстанавливает итог попытки из журнала шагов.
func Evaluate(doc domain.ScenarioDoc, steps []domain.AttemptStep) (SummaryResult, error) {
	var earned float64
	var endingKey string

	flags := make([]string, 0, len(steps))
	seen := make(map[string]struct{}, len(steps))

	for _, step := range steps {
		effect, err := resolveStep(doc, step)
		if err != nil {
			return SummaryResult{}, err
		}

		earned += effect.weight
		if effect.endingKey != "" {
			endingKey = effect.endingKey
		}

		if effect.flag != "" {
			if _, dup := seen[effect.flag]; !dup {
				seen[effect.flag] = struct{}{}
				flags = append(flags, effect.flag)
			}
		}
	}

	ending, err := pickEnding(doc, endingKey, len(flags) > 0)
	if err != nil {
		return SummaryResult{}, err
	}

	return SummaryResult{
		Score:       scorePercent(earned, doc.TotalWeight()),
		Outcome:     ending.Outcome,
		Ending:      ending,
		MissedFlags: describeFlags(doc, flags),
		Takeaway:    doc.Debrief.Takeaway,
		StepsTotal:  len(steps),
	}, nil
}

// resolveStep восстанавливает последствия выбора по документу сценария:
// в attempt_steps лежит только пара (scene_id, option_id).
func resolveStep(doc domain.ScenarioDoc, step domain.AttemptStep) (stepEffect, error) {
	scene, ok := doc.SceneByID(step.SceneID)
	if !ok {
		return stepEffect{}, fmt.Errorf(
			"%w: logged scene %q is missing from scenario", domainErrors.ErrInvalidScenario, step.SceneID)
	}

	opt, ok := scene.OptionByID(step.OptionID)
	if !ok {
		return stepEffect{}, fmt.Errorf(
			"%w: %q in scene %q", domainErrors.ErrUnknownOption, step.OptionID, step.SceneID)
	}

	var effect stepEffect
	if opt.Verdict == domain.VerdictSafe {
		effect.weight = scene.Weight
	} else {
		effect.flag = opt.Flag
	}

	if opt.Verdict == domain.VerdictFatal {
		effect.endingKey = opt.Ending
	}

	return effect, nil
}

// pickEnding выбирает концовку: fatal задаёт её явно, иначе она следует
// из того, набрал ли пользователь пропущенные признаки.
func pickEnding(doc domain.ScenarioDoc, endingKey string, flagged bool) (domain.Ending, error) {
	if endingKey == "" {
		endingKey = endingKeySafe
		if flagged {
			endingKey = endingKeyPartial
		}
	}

	ending, ok := doc.Endings[endingKey]
	if !ok {
		return domain.Ending{}, fmt.Errorf(
			"%w: ending %q not found", domainErrors.ErrInvalidScenario, endingKey)
	}

	return ending, nil
}

// describeFlags разворачивает признаки в описания из разбора, сохраняя порядок.
// Признак без описания в разбор не попадёт, но на выбор концовки уже повлиял.
func describeFlags(doc domain.ScenarioDoc, flags []string) []domain.FlagInfo {
	out := make([]domain.FlagInfo, 0, len(flags))
	for _, id := range flags {
		if info, ok := doc.FlagByID(id); ok {
			out = append(out, info)
		}
	}
	return out
}

// scorePercent переводит накопленный вес в проценты.
func scorePercent(earned, total float64) int {
	if total <= 0 {
		return 0
	}
	return int(math.Round(earned / total * 100))
}
