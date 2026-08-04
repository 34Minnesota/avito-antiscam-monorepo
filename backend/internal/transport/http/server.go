package http

import (
	"net/http"

	"github.com/gin-gonic/gin"

	openapi "github.com/34Minnesota/avito-antiscam-monorepo/backend/generated/openapi"
	authservice "github.com/34Minnesota/avito-antiscam-monorepo/backend/internal/services/auth"
)

type Server struct {
	auth *authservice.Service
}

func NewServer(auth *authservice.Service) *Server {
	return &Server{
		auth: auth,
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
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": "failed to create session",
		})
		return
	}

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