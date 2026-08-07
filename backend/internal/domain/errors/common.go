package domainErrors

import "errors"

var (
	ErrUnauthorized = errors.New("unauthorized")
	ErrValidation   = errors.New("validation failed")
)
