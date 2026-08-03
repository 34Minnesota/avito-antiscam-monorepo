package core_errors

import "errors"

var (
	ErrInvalidMaxPoints = errors.New("max points must be greater than zero")
	ErrNegativePoints   = errors.New("points must not be negative")
	ErrNegativeDelta    = errors.New("delta must not be negative")
	ErrScoreOverflow    = errors.New("score points exceed max points")
)
