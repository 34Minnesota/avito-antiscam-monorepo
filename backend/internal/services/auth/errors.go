package auth

import "errors"

var (
	// ErrInvalidCredentials возвращается,
	// если пользователь не найден или пароль неверный.
	ErrInvalidCredentials = errors.New("invalid credentials")

	// ErrSessionExpired возвращается,
	// если срок действия сессии истек.
	ErrSessionExpired = errors.New("session expired")
)
