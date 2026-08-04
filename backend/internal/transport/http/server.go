package httptransport

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	openapi "github.com/34Minnesota/avito-antiscam-monorepo/backend/generated/openapi"
	applogger "github.com/34Minnesota/avito-antiscam-monorepo/backend/internal/logger"
	authservice "github.com/34Minnesota/avito-antiscam-monorepo/backend/internal/services/auth"
)

type Server struct {
	auth   *authservice.Service
	logger *applogger.Logger
}

func NewServer(auth *authservice.Service, logger *applogger.Logger) *Server {
	return &Server{
		auth:   auth,
		logger: logger,
	}
}

// -----------------------------------------------------
// Healthcheck
// -----------------------------------------------------

func (s *Server) HealthCheck(c *gin.Context) {
	c.Status(http.StatusOK)
}

// -----------------------------------------------------
// Session
// -----------------------------------------------------

func (s *Server) CreateSession(c *gin.Context) {

	session, err := s.auth.CreateSession(c.Request.Context())
	if err != nil {
		if s.logger != nil {
			s.logger.Error("create session request failed", zap.Error(err))
		}
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": "failed to create session",
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"sessionId": session.ID,
	})
}

// -----------------------------------------------------
// Остальные endpoint'ы пока заглушки
// -----------------------------------------------------

func (s *Server) ListScenarios(
	c *gin.Context,
) {
	c.JSON(http.StatusNotImplemented, gin.H{
		"message": "not implemented",
	})
}

func (s *Server) CreateAttempt(
	c *gin.Context,
	slug string,
) {
	c.JSON(http.StatusNotImplemented, gin.H{
		"message": "not implemented",
	})
}

func (s *Server) GetAttempt(
	c *gin.Context,
	attemptID openapi.AttemptID,
) {
	c.JSON(http.StatusNotImplemented, gin.H{
		"message": "not implemented",
	})
}

func (s *Server) ExecuteAction(
	c *gin.Context,
	attemptID openapi.AttemptID,
) {
	c.JSON(http.StatusNotImplemented, gin.H{
		"message": "not implemented",
	})
}

func (s *Server) GetProgress(
	c *gin.Context,
) {
	c.JSON(http.StatusNotImplemented, gin.H{
		"message": "not implemented",
	})
}
