package httptransport

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go.uber.org/zap"

	"github.com/34Minnesota/avito-antiscam-monorepo/backend/internal/domain"
	applogger "github.com/34Minnesota/avito-antiscam-monorepo/backend/internal/logger"
	authservice "github.com/34Minnesota/avito-antiscam-monorepo/backend/internal/services/auth"
)

type Server struct {
	auth     *authservice.Service
	training *training.Service
	logger   *applogger.Logger
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

// ListScenarios — GET /v1/scenarios?role=buyer|seller
func (s *Server) ListScenarios(c *gin.Context) {
	userID, ok := CurrentUser(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"message": "unauthorized"})
		return
	}

	role := domain.Role(c.Query("role"))
	if role != "" && !(role == domain.RoleBuyer || role == domain.RoleSeller) {
		c.JSON(http.StatusBadRequest, gin.H{"code": "bad_request", "message": "неизвестная роль: " + string(role)})
		return
	}

	scenarios, err := s.training.Catalog(c.Request.Context(), userID, role)
	if err != nil {
		if s.logger != nil {
			s.logger.Error("list scenarios failed", zap.Error(err))
		}
		c.JSON(http.StatusInternalServerError, gin.H{"message": "failed to list scenarios"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"scenarios": scenarios})
}

// StartAttempt — POST /v1/attempts
func (s *Server) StartAttempt(c *gin.Context) {
	userID, ok := CurrentUser(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"message": "unauthorized"})
		return
	}

	var req StartAttemptRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": "bad_request", "message": err.Error()})
		return
	}

	result, err := s.training.Start(c.Request.Context(), userID, req.ScenarioID)
	if err != nil {
		if s.logger != nil {
			s.logger.Error("start attempt failed", zap.Error(err))
		}
		c.JSON(http.StatusInternalServerError, gin.H{"message": "failed to start attempt"})
		return
	}
	c.JSON(http.StatusCreated, result)
}

// SubmitChoice — POST /v1/attempts/:attemptID/choice
func (s *Server) SubmitChoice(c *gin.Context) {
	userID, ok := CurrentUser(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"message": "unauthorized"})
		return
	}

	attemptID, err := uuid.Parse(c.Param("attemptID"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": "bad_request", "message": err.Error()})
		return
	}

	var req ChoiceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": "bad_request", "message": err.Error()})
		return
	}

	result, err := s.training.Choose(c.Request.Context(), userID, attemptID, req.SceneID, req.OptionID)
	if err != nil {
		if s.logger != nil {
			s.logger.Error("submit choice failed", zap.Error(err))
		}
		c.JSON(http.StatusInternalServerError, gin.H{"message": "failed to submit choice"})
		return
	}
	c.JSON(http.StatusOK, result)
}

// GetSummary — GET /v1/attempts/:attemptID/summary
func (s *Server) GetSummary(c *gin.Context) {
	userID, ok := CurrentUser(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"message": "unauthorized"})
		return
	}

	attemptID, err := uuid.Parse(c.Param("attemptID"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": "bad_request", "message": err.Error()})
		return
	}

	summary, err := s.training.Summary(c.Request.Context(), userID, attemptID)
	if err != nil {
		if s.logger != nil {
			s.logger.Error("get summary failed", zap.Error(err))
		}
		c.JSON(http.StatusInternalServerError, gin.H{"message": "failed to get summary"})
		return
	}
	c.JSON(http.StatusOK, summary)
}
