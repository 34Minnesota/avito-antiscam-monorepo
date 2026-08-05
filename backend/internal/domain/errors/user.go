package domainErrors

import "errors"

var (
	ErrInvalidUserID       = errors.New("user id must not be nil")
	ErrInvalidProgressData = errors.New("progress data violates domain invariants")
)
