package domain

import (
	"math"

	core_errors "github.com/34Minnesota/avito-antiscam-monorepo/backend/internal/errors"
)

type Score struct {
	points    int
	maxPoints int
}

func NewScore(points, maxPoints int) (Score, error) {
	if maxPoints <= 0 {
		return Score{}, core_errors.ErrInvalidMaxPoints
	}

	if points < 0 {
		return Score{}, core_errors.ErrNegativePoints
	}

	if points > maxPoints {
		return Score{}, core_errors.ErrScoreOverflow
	}

	return Score{
		points:    points,
		maxPoints: maxPoints,
	}, nil
}

func (s Score) Points() int {
	return s.points
}

func (s Score) MaxPoints() int {
	return s.maxPoints
}

func (s Score) Percent() int {
	return int(math.Round(float64(s.points) * 100 / float64(s.maxPoints)))
}

func (s Score) Apply(scoreDelta int) (Score, error) {
	if scoreDelta < 0 {
		return Score{}, core_errors.ErrNegativeDelta
	}

	if scoreDelta > s.maxPoints-s.points {
		return Score{}, core_errors.ErrScoreOverflow
	}

	return NewScore(s.points+scoreDelta, s.maxPoints)
}
