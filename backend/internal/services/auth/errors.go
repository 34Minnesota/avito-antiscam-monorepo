package auth

import "errors"

var (
	// ErrInvalidSession возвращается, если сессия не существует
	// или идентификатор сессии некорректен.
	ErrInvalidSession = errors.New("invalid session")

	// ErrSessionExpired возвращается, если срок действия
	// пользовательской сессии истек.
	ErrSessionExpired = errors.New("session expired")

	// ErrUnauthorized возвращается, если пользователь
	// не прошел авторизацию.
	ErrUnauthorized = errors.New("unauthorized")

	// ErrInvalidCredentials возвращается при неверном
	// email или пароле.
	ErrInvalidCredentials = errors.New("invalid credentials")
)
