package api

import (
	"net/http"

	"github.com/gin-gonic/gin"

	openapi "github.com/34Minnesota/avito-antiscam-monorepo/backend/generated/openapi"
	"github.com/34Minnesota/avito-antiscam-monorepo/backend/internal/services/auth"
)

type Server struct {
	auth *auth.Service
}

func NewServer(authService *auth.Service) *Server {
	return &Server{
		auth: authService,
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
	session := s.auth.CreateSession()

	response := openapi.CreateSessionResponse{
		SessionId: session.ID,
	}

	c.JSON(http.StatusCreated, response)
}

// -----------------------------------------------------
// Остальные endpoint'ы пока заглушки
// -----------------------------------------------------

func (s *Server) ListScenarios(
	c *gin.Context,
	params openapi.ListScenariosParams,
) {
	c.JSON(http.StatusNotImplemented, gin.H{
		"message": "not implemented",
	})
}

func (s *Server) CreateAttempt(
	c *gin.Context,
	slug string,
	params openapi.CreateAttemptParams,
) {
	c.JSON(http.StatusNotImplemented, gin.H{
		"message": "not implemented",
	})
}

func (s *Server) GetAttempt(
	c *gin.Context,
	attemptID openapi.AttemptID,
	params openapi.GetAttemptParams,
) {
	c.JSON(http.StatusNotImplemented, gin.H{
		"message": "not implemented",
	})
}

func (s *Server) ExecuteAction(
	c *gin.Context,
	attemptID openapi.AttemptID,
	params openapi.ExecuteActionParams,
) {
	c.JSON(http.StatusNotImplemented, gin.H{
		"message": "not implemented",
	})
}

func (s *Server) GetProgress(
	c *gin.Context,
	params openapi.GetProgressParams,
) {
	c.JSON(http.StatusNotImplemented, gin.H{
		"message": "not implemented",
	})
}