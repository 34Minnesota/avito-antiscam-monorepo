package httptransport

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

const sessionHeader = "X-Session-ID"

func (s *Server) SessionMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {

		sessionID := c.GetHeader(sessionHeader)
		if sessionID == "" {
			s.logAuthFailure("missing session id")
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"message": "missing session id",
			})
			return
		}

		id, err := uuid.Parse(sessionID)
		if err != nil {
			s.logAuthFailure(
				"invalid session id",
				zap.Error(err),
			)

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
			s.logAuthFailure(
				"session validation failed",
				zap.String("session_id", id.String()),
				zap.Error(err),
			)

			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"message": "invalid session",
			})
			return
		}

		if time.Now().UTC().After(session.ExpiresAt) {
			s.logAuthFailure(
				"session expired",
				zap.String("session_id", id.String()),
			)

			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"message": "session expired",
			})
			return
		}

		// Пока кладем Session.
		// После появления UsersService здесь будет User.
		c.Set(sessionContextKey, session)

		c.Next()
	}
}

func (s *Server) LoggerMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {

		startedAt := time.Now()

		c.Next()

		if s.logger == nil {
			return
		}

		fields := []zap.Field{
			zap.String("method", c.Request.Method),
			zap.String("path", c.Request.URL.Path),
			zap.Int("status", c.Writer.Status()),
			zap.Duration("duration", time.Since(startedAt)),
		}

		if c.Errors.Last() != nil {
			fields = append(
				fields,
				zap.Error(c.Errors.Last().Err),
			)
		}

		s.logger.Info("http request", fields...)
	}
}

func (s *Server) logAuthFailure(
	message string,
	fields ...zap.Field,
) {
	if s.logger != nil {
		s.logger.Warn(message, fields...)
	}
}
