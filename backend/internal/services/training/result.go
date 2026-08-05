package training

import (
	"fmt"
	"math"

	"github.com/34Minnesota/avito-antiscam-monorepo/backend/internal/domain"
	domainErrors "github.com/34Minnesota/avito-antiscam-monorepo/backend/internal/domain/errors"
)

// Подсчёт итога прохождения.

// Evaluate восстанавливает итог попытки из журнала шагов.
func Evaluate(doc domain.ScenarioDoc, steps []domain.AttemptStep) (SummaryResult, error) {
	var earned float64
	var endingKey string

	missed := make([]domain.FlagInfo, 0, len(steps))
	seen := make(map[string]struct{}, len(steps))

	for _, step := range steps {
		scene, ok := doc.SceneByID(step.SceneID)
		if !ok {
			return SummaryResult{}, fmt.Errorf(
				"%w: logged scene %q is missing from scenario", domainErrors.ErrInvalidScenario, step.SceneID)
		}

		opt, ok := scene.OptionByID(step.OptionID)
		if !ok {
			return SummaryResult{}, fmt.Errorf(
				"%w: %q in scene %q", domainErrors.ErrUnknownOption, step.OptionID, step.SceneID)
		}

		if opt.Verdict == domain.VerdictSafe {
			earned += scene.Weight
		} else if opt.Flag != "" {
			if _, dup := seen[opt.Flag]; !dup {
				seen[opt.Flag] = struct{}{}
				if info, found := doc.FlagByID(opt.Flag); found {
					missed = append(missed, info)
				}
			}
		}

		if opt.Verdict == domain.VerdictFatal {
			endingKey = opt.Ending
		}
	}

	if endingKey == "" {
		endingKey = endingKeySafe
		if len(seen) > 0 {
			endingKey = endingKeyPartial
		}
	}

	ending, found := doc.Endings[endingKey]
	if !found {
		return SummaryResult{}, fmt.Errorf(
			"%w: ending %q not found", domainErrors.ErrInvalidScenario, endingKey)
	}

	return SummaryResult{
		Score:       scorePercent(earned, doc.TotalWeight()),
		Outcome:     ending.Outcome,
		Ending:      ending,
		MissedFlags: missed,
		Takeaway:    doc.Debrief.Takeaway,
		StepsTotal:  len(steps),
	}, nil
}

// scorePercent переводит накопленный вес в проценты.
func scorePercent(earned, total float64) int {
	if total <= 0 {
		return 0
	}
	return int(math.Round(earned / total * 100))
}
