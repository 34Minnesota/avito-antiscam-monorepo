package scoring_test

import (
	"errors"
	"math"
	"testing"

	domainErrors "github.com/34Minnesota/avito-antiscam-monorepo/backend/internal/domain/errors"
	"github.com/34Minnesota/avito-antiscam-monorepo/backend/internal/services/scoring"
)

func TestEvaluateUsesExactThreshold(t *testing.T) {
	t.Parallel()
	service := scoring.New()
	below, err := service.Evaluate(696, 1000, 70)
	if err != nil {
		t.Fatal(err)
	}
	if below.Score.Percent() != 70 || below.Passed {
		t.Fatalf("rounded 70%% must remain below exact threshold")
	}
	at, err := service.Evaluate(700, 1000, 70)
	if err != nil {
		t.Fatal(err)
	}
	if !at.Passed {
		t.Fatal("exact threshold must pass")
	}
}

func TestEvaluateDoesNotOverflow(t *testing.T) {
	t.Parallel()
	result, err := scoring.New().Evaluate(math.MaxInt, math.MaxInt, 100)
	if err != nil {
		t.Fatal(err)
	}
	if !result.Passed || result.Score.Percent() != 100 {
		t.Fatalf("unexpected result: %+v", result)
	}
}

func TestEvaluateValidatesPassPercent(t *testing.T) {
	t.Parallel()
	_, err := scoring.New().Evaluate(1, 1, 0)
	if !errors.Is(err, domainErrors.ErrInvalidPassRate) {
		t.Fatalf("got %v", err)
	}
}
