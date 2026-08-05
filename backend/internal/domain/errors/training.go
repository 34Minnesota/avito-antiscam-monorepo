package domainErrors

import "errors"

var (
	ErrNotFound        = errors.New("not found")
	ErrForbidden       = errors.New("forbidden")
	ErrUnauthorized    = errors.New("unauthorized")
	ErrOutOfOrder      = errors.New("scene is out of order")
	ErrUnknownOption   = errors.New("unknown option")
	ErrAttemptFinished = errors.New("attempt already finished")
	ErrInvalidScenario = errors.New("invalid scenario")
	ErrValidation      = errors.New("validation failed")
)
