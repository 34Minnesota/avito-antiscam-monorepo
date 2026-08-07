package domainErrors

import "errors"

var (
	ErrInvalidProgressData = errors.New("progress data violates domain invariants")
)
