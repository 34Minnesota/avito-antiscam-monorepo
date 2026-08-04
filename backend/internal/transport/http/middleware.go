package http

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

func (s *Server) SessionMiddleware() gin.HandlerFunc {

	return func(c *gin.Context) {

		sessionID := c.GetHeader("X-Session-ID")

		if sessionID == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"message": "missing session id",
			})
			return
		}

		id, err := uuid.Parse(sessionID)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"message": "invalid session id",
			})
			return
		}

		session, err := s.auth.ValidateSession(
			c.Request.Context(),
			id,
		)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"message": "session not found",
			})
			return
		}

		c.Set("session", session)

		c.Next()
	}
}
