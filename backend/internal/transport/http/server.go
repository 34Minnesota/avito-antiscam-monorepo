package httptransport

import (
	"errors"
	"net/http"

	coreErrors "github.com/34Minnesota/avito-antiscam-monorepo/backend/internal/errors"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go.uber.org/zap"

	openapi "github.com/34Minnesota/avito-antiscam-monorepo/backend/generated/openapi"
	"github.com/34Minnesota/avito-antiscam-monorepo/backend/internal/domain"
	applogger "github.com/34Minnesota/avito-antiscam-monorepo/backend/internal/logger"
	authservice "github.com/34Minnesota/avito-antiscam-monorepo/backend/internal/services/auth"
	"github.com/34Minnesota/avito-antiscam-monorepo/backend/internal/services/training"
	usersservice "github.com/34Minnesota/avito-antiscam-monorepo/backend/internal/services/users"
)

type Server struct {
	auth     *authservice.Service
	users    *usersservice.UsersService
	training *training.Service
	logger   *applogger.Logger
}

func NewServer(
	auth *authservice.Service,
	users *usersservice.UsersService,
	training *training.Service,
	logger *applogger.Logger,
) *Server {
	return &Server{
		auth:     auth,
		users:    users,
		training: training,
		logger:   logger,
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

func (s *Server) Login(c *gin.Context) {
	var req openapi.LoginRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    "bad_request",
			"message": err.Error(),
		})
		return
	}

	session, err := s.auth.Login(
		c.Request.Context(),
		string(req.Email),
		req.Password,
	)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"code":    "unauthorized",
			"message": "invalid email or password",
		})
		return
	}

	c.JSON(http.StatusOK, openapi.LoginResponse{
		SessionId: session.ID,
	})
}

func (s *Server) Register(c *gin.Context) {
	var req openapi.CreateUserRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    "bad_request",
			"message": "invalid request",
		})
		return
	}

	user, err := s.users.CreateUser(
		c.Request.Context(),
		usersservice.CreateUserInput{
			Nickname: req.Nickname,
			Email:    string(req.Email),
			Password: req.Password,
		},
	)
	if err != nil {
		switch {
		case errors.Is(err, coreErrors.ErrInvalidArgument):
			c.JSON(http.StatusBadRequest, gin.H{
				"code":    "bad_request",
				"message": "invalid request",
			})
		case errors.Is(err, coreErrors.ErrConflict):
			c.JSON(http.StatusConflict, gin.H{
				"code":    "conflict",
				"message": "user already exists",
			})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{
				"code":    "internal_error",
				"message": "internal server error",
			})
		}
		return
	}

	session, err := s.auth.CreateSession(
		c.Request.Context(),
		user.ID,
	)
	if err != nil {
		if s.logger != nil {
			s.logger.Error("create session failed", zap.Error(err))
		}

		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    "internal_error",
			"message": "failed to create session",
		})
		return
	}

	c.JSON(http.StatusCreated, openapi.LoginResponse{
		SessionId: session.ID,
	})
}

func (s *Server) Logout(c *gin.Context) {
	value, exists := c.Get(sessionContextKey)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"code":    "unauthorized",
			"message": "unauthorized",
		})
		return
	}

	session, ok := value.(domain.Session)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{
			"code":    "unauthorized",
			"message": "unauthorized",
		})
		return
	}

	if err := s.auth.Logout(
		c.Request.Context(),
		session.ID,
	); err != nil {

		if s.logger != nil {
			s.logger.Error("logout failed", zap.Error(err))
		}

		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    "internal_error",
			"message": "logout failed",
		})
		return
	}

	c.Status(http.StatusNoContent)
}
func (s *Server) GetCurrentUser(c *gin.Context) {

	userID, ok := CurrentUser(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{
			"code":    "unauthorized",
			"message": "unauthorized",
		})
		return
	}

	user, err := s.users.GetUser(
		c.Request.Context(),
		userID.UUID(),
	)
	if err != nil {
		if s.logger != nil {
			s.logger.Error("get current user failed", zap.Error(err))
		}

		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    "internal_error",
			"message": "failed to get current user",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"id":         user.ID,
		"nickname":   user.Nickname,
		"email":      user.Email,
		"created_at": user.CreatedAt,
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
	if role != "" && (role != domain.RoleBuyer && role != domain.RoleSeller) {
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
