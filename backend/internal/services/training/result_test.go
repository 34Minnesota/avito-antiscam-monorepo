package training_test

import (
	"errors"
	"testing"

	"github.com/34Minnesota/avito-antiscam-monorepo/backend/internal/domain"
	domainErrors "github.com/34Minnesota/avito-antiscam-monorepo/backend/internal/domain/errors"
	"github.com/34Minnesota/avito-antiscam-monorepo/backend/internal/services/training"
)

func TestEvaluateScoresRuns(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name    string
		journal []domain.AttemptStep
		score   int
		outcome domain.Outcome
		missed  int
	}{
		{
			name:    "все безопасные варианты дают сто процентов",
			journal: steps([2]string{"s1", "a1"}, [2]string{"s2", "b1"}, [2]string{"s3", "c1"}),
			score:   100,
			outcome: domain.OutcomeSafe,
		},
		{
			name:    "один рискованный выбор снимает вес своей сцены",
			journal: steps([2]string{"s1", "a2"}, [2]string{"s2", "b1"}, [2]string{"s3", "c1"}),
			score:   75,
			outcome: domain.OutcomePartial,
			missed:  1,
		},
		{
			name:    "повторный признак не удваивается в разборе",
			journal: steps([2]string{"s1", "a2"}, [2]string{"s2", "b2"}, [2]string{"s3", "c1"}),
			score:   25,
			outcome: domain.OutcomePartial,
			missed:  1,
		},
		{
			name:    "fatal обрывает прохождение своей концовкой",
			journal: steps([2]string{"s1", "a3"}),
			score:   0,
			outcome: domain.OutcomeScammed,
			missed:  1,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			result, err := training.Evaluate(testDoc(), tc.journal)
			if err != nil {
				t.Fatal(err)
			}
			if result.Score != tc.score || result.Outcome != tc.outcome {
				t.Fatalf("unexpected result: %+v", result)
			}
			if len(result.MissedFlags) != tc.missed {
				t.Fatalf("missed flags = %d: %+v", len(result.MissedFlags), result.MissedFlags)
			}
			if result.StepsTotal != len(tc.journal) || result.Takeaway != "Вывод" {
				t.Fatalf("unexpected summary: %+v", result)
			}
		})
	}
}

func TestEvaluateSkipsFlagMissingFromDebrief(t *testing.T) {
	t.Parallel()
	result, err := training.Evaluate(testDoc(), steps(
		[2]string{"s1", "a1"}, [2]string{"s2", "b1"}, [2]string{"s3", "c2"},
	))
	if err != nil {
		t.Fatal(err)
	}
	if result.Outcome != domain.OutcomePartial {
		t.Fatalf("outcome = %s", result.Outcome)
	}
	if len(result.MissedFlags) != 0 {
		t.Fatalf("unknown flag must not reach debrief: %+v", result.MissedFlags)
	}
}

func TestEvaluateKeepsFlagOrder(t *testing.T) {
	t.Parallel()
	result, err := training.Evaluate(testDoc(), steps([2]string{"s1", "a3"}, [2]string{"s2", "b2"}))
	if err != nil {
		t.Fatal(err)
	}
	if len(result.MissedFlags) != 2 {
		t.Fatalf("missed = %+v", result.MissedFlags)
	}
	if result.MissedFlags[0].ID != "f2" || result.MissedFlags[1].ID != "f1" {
		t.Fatalf("flags must follow the journal order: %+v", result.MissedFlags)
	}
}

func TestEvaluateEmptyJournalEndsSafe(t *testing.T) {
	t.Parallel()
	result, err := training.Evaluate(testDoc(), nil)
	if err != nil {
		t.Fatal(err)
	}
	if result.Score != 0 || result.Outcome != domain.OutcomeSafe || result.StepsTotal != 0 {
		t.Fatalf("unexpected result: %+v", result)
	}
}

func TestEvaluateRejectsJournalThatDoesNotMatchScenario(t *testing.T) {
	t.Parallel()

	_, err := training.Evaluate(testDoc(), steps([2]string{"ghost", "a1"}))
	if !errors.Is(err, domainErrors.ErrInvalidScenario) {
		t.Fatalf("unknown scene: got %v", err)
	}

	_, err = training.Evaluate(testDoc(), steps([2]string{"s1", "ghost"}))
	if !errors.Is(err, domainErrors.ErrUnknownOption) {
		t.Fatalf("unknown option: got %v", err)
	}
}

func TestEvaluateRejectsMissingEnding(t *testing.T) {
	t.Parallel()

	doc := testDoc()
	delete(doc.Endings, "safe")
	if _, err := training.Evaluate(doc, nil); !errors.Is(err, domainErrors.ErrInvalidScenario) {
		t.Fatalf("missing safe ending: got %v", err)
	}

	doc = testDoc()
	delete(doc.Endings, "partial")
	_, err := training.Evaluate(doc, steps([2]string{"s1", "a2"}))
	if !errors.Is(err, domainErrors.ErrInvalidScenario) {
		t.Fatalf("missing partial ending: got %v", err)
	}

	doc = testDoc()
	delete(doc.Endings, "lost")
	_, err = training.Evaluate(doc, steps([2]string{"s1", "a3"}))
	if !errors.Is(err, domainErrors.ErrInvalidScenario) {
		t.Fatalf("missing fatal ending: got %v", err)
	}
}

func TestEvaluateZeroWeightScenarioScoresZero(t *testing.T) {
	t.Parallel()
	doc := testDoc()
	for i := range doc.Scenes {
		doc.Scenes[i].Weight = 0
	}
	result, err := training.Evaluate(doc, steps([2]string{"s1", "a1"}))
	if err != nil {
		t.Fatal(err)
	}
	if result.Score != 0 {
		t.Fatalf("score = %d", result.Score)
	}
}
