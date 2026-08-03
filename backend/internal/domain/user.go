package domain

import (
	"time"

	"github.com/google/uuid"
)

// User представляет пользователя приложения.
//
// Пока мы поддерживаем только гостевые аккаунты.
// В будущем сюда можно добавить email, пароль,
// аватар, достижения и т.д.
type User struct {
	ID uuid.UUID

	Nickname string

	IsGuest bool

	CreatedAt time.Time
}