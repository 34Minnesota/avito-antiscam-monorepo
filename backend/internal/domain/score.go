package domain

import (
	"math"

	core_errors "github.com/34Minnesota/avito-antiscam-monorepo/backend/internal/errors"
)

type Score struct {
	Points    int
	MaxPoints int
	Percent   int
}

func NewScore(
	points, maxPoints, percent int,
) Score {
	return Score{
		Points:    points,
		MaxPoints: maxPoints,
		Percent:   percent,
	}
}

func newScore(points, maxPoints int) (Score, error) {
	if maxPoints <= 0 {
		return Score{}, core_errors.ErrInvalidMaxPoints
	}

	if points < 0 {
		return Score{}, core_errors.ErrNegativePoints
	}

	if points > maxPoints {
		return Score{}, core_errors.ErrScoreOverflow
	}

	percent := int(math.Round(
		float64(points) * 100 / float64(maxPoints),
	))

	return NewScore(points, maxPoints, percent), nil
}

func (s Score) Apply(scoreDelta int) (Score, error) {
	if scoreDelta < 0 {
		return Score{}, core_errors.ErrNegativeDelta
	}

	if scoreDelta > s.MaxPoints-s.Points {
		return Score{}, core_errors.ErrScoreOverflow
	}

	return newScore(s.Points+scoreDelta, s.MaxPoints)
}
