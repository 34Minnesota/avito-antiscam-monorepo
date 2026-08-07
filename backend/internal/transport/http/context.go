package httptransport

import (
	"github.com/gin-gonic/gin"

	"github.com/34Minnesota/avito-antiscam-monorepo/backend/internal/domain"
)

const sessionContextKey = "session"

// CurrentUser достаёт пользователя из сессии, проверенной SessionMiddleware.
func CurrentUser(c *gin.Context) (domain.UserID, bool) {
	value, exists := c.Get(sessionContextKey)
	if !exists {
		return domain.UserID{}, false
	}

	session, ok := value.(domain.Session)
	if !ok {
		return domain.UserID{}, false
	}

	userID, err := domain.NewUserID(session.UserID)
	if err != nil {
		return domain.UserID{}, false
	}
	return userID, true
}

// CurrentSession достает текущую сессию из контекста.
func CurrentSession(c *gin.Context) (domain.Session, bool) {
	value, exists := c.Get(sessionContextKey)
	if !exists {
		return domain.Session{}, false
	}

	session, ok := value.(domain.Session)
	if !ok {
		return domain.Session{}, false
	}

	return session, true
}
