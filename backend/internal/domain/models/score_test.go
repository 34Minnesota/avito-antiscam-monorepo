package models_test

import (
	"errors"
	"math"
	"testing"

	domainErrors "github.com/34Minnesota/avito-antiscam-monorepo/backend/internal/domain/errors"
	"github.com/34Minnesota/avito-antiscam-monorepo/backend/internal/domain/models"
)

func TestNewScoreValidation(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name        string
		points, max int
		want        error
	}{
		{name: "zero max", points: 0, max: 0, want: domainErrors.ErrInvalidMaxPoints},
		{name: "negative points", points: -1, max: 10, want: domainErrors.ErrNegativePoints},
		{name: "overflow", points: 11, max: 10, want: domainErrors.ErrScoreOverflow},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			_, err := models.NewScore(tt.points, tt.max)
			if !errors.Is(err, tt.want) {
				t.Fatalf("got %v, want %v", err, tt.want)
			}
		})
	}
}

func TestScorePercentDoesNotOverflow(t *testing.T) {
	t.Parallel()
	score, err := models.NewScore(math.MaxInt, math.MaxInt)
	if err != nil {
		t.Fatal(err)
	}
	if score.Percent() != 100 {
		t.Fatalf("percent = %d, want 100", score.Percent())
	}
}

func TestScoreApplyAndPercent(t *testing.T) {
	t.Parallel()
	score, err := models.NewScore(691, 1000)
	if err != nil {
		t.Fatal(err)
	}
	if score.Percent() != 69 {
		t.Fatalf("percent = %d, want 69", score.Percent())
	}
	updated, err := score.Apply(5)
	if err != nil {
		t.Fatal(err)
	}
	if updated.Points() != 696 || updated.Percent() != 70 {
		t.Fatalf("unexpected score: %d/%d%%", updated.Points(), updated.Percent())
	}
	if _, err := updated.Apply(305); !errors.Is(err, domainErrors.ErrScoreOverflow) {
		t.Fatalf("got %v", err)
	}
	if _, err := updated.Apply(-1); !errors.Is(err, domainErrors.ErrNegativeDelta) {
		t.Fatalf("got %v", err)
	}
}
