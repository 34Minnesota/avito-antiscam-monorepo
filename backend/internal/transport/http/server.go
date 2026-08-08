package httptransport

import (
	"errors"
	"net/http"

	coreErrors "github.com/34Minnesota/avito-antiscam-monorepo/backend/internal/domain/errors"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go.uber.org/zap"

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
// HealthCheck godoc
//
//	@Summary	Проверка состояния сервиса
//	@Tags		health
//	@Produce	plain
//	@Success	200
//	@Router		/healthz [get]
func (s *Server) HealthCheck(c *gin.Context) {
	c.Status(http.StatusOK)
}

// -----------------------------------------------------
// Session
// -----------------------------------------------------

// Login godoc
//
//	@Summary		Вход в систему
//	@Description	Выполняет вход по email и паролю и создаёт сессию.
//	@Tags			auth
//	@Accept			json
//	@Produce		json
//	@Param			request	body		LoginRequest	true	"Учётные данные"
//	@Success		200		{object}	LoginResponse
//	@Failure		400		{object}	ErrorResponse
//	@Failure		401		{object}	ErrorResponse
//	@Router			/v1/auth/login [post]
func (s *Server) Login(c *gin.Context) {
	var req LoginRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    "bad_request",
			"message": err.Error(),
		})
		return
	}

	session, err := s.auth.Login(
		c.Request.Context(),
		req.Email,
		req.Password,
	)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"code":    "unauthorized",
			"message": "invalid email or password",
		})
		return
	}

	c.JSON(http.StatusOK, LoginResponse{
		SessionID: session.ID,
	})
}

// Register godoc
//
//	@Summary		Регистрация пользователя
//	@Description	Создаёт пользователя и связанную с ним сессию.
//	@Tags			auth
//	@Accept			json
//	@Produce		json
//	@Param			request	body		RegisterRequest	true	"Данные для регистрации"
//	@Success		201		{object}	RegisterResponse
//	@Failure		400		{object}	ErrorResponse
//	@Failure		409		{object}	ErrorResponse
//	@Failure		500		{object}	ErrorResponse
//	@Router			/v1/auth/register [post]
func (s *Server) Register(c *gin.Context) {
	var req RegisterRequest

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
			Email:    req.Email,
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

	c.JSON(http.StatusCreated, RegisterResponse{
		SessionID: session.ID,
	})
}

// Logout godoc
//
//	@Summary		Выход из системы
//	@Description	Завершает текущую сессию.
//	@Tags			auth
//	@Produce		json
//	@Security		SessionID
//	@Success		204
//	@Failure		401	{object}	ErrorResponse
//	@Failure		500	{object}	ErrorResponse
//	@Router			/v1/auth/logout [post]
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

// GetCurrentUser godoc
//
//	@Summary		Получить текущего пользователя
//	@Description	Возвращает данные пользователя, связанного с текущей сессией.
//	@Tags			users
//	@Produce		json
//	@Security		SessionID
//	@Success		200	{object}	UserResponse
//	@Failure		401	{object}	ErrorResponse
//	@Failure		500	{object}	ErrorResponse
//	@Router			/v1/users/me [get]
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

	c.JSON(http.StatusOK, UserResponse{
		ID:        user.ID,
		Nickname:  user.Nickname,
		Email:     user.Email,
		CreatedAt: user.CreatedAt,
	})
}

// -----------------------------------------------------
// Остальные endpoint'ы пока заглушки
// -----------------------------------------------------

// ListScenarios godoc
//
//	@Summary		Получить список сценариев
//	@Description	Возвращает доступные пользователю сценарии.
//	@Tags			training
//	@Produce		json
//	@Security		SessionID
//	@Param			role	query		string	false	"Роль пользователя"	Enums(buyer,seller)
//	@Success		200		{object}	map[string]interface{}
//	@Failure		400		{object}	ErrorResponse
//	@Failure		401		{object}	ErrorResponse
//	@Failure		500		{object}	ErrorResponse
//	@Router			/v1/scenarios [get]
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

// StartAttempt godoc
//
//	@Summary		Начать попытку
//	@Description	Создаёт новую попытку прохождения сценария.
//	@Tags			attempts
//	@Accept			json
//	@Produce		json
//	@Security		SessionID
//	@Param			request	body		StartAttemptRequest	true	"Данные сценария"
//	@Success		201		{object}	map[string]interface{}
//	@Failure		400		{object}	ErrorResponse
//	@Failure		401		{object}	ErrorResponse
//	@Failure		500		{object}	ErrorResponse
//	@Router			/v1/attempts [post]
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
		writeTrainingError(c, err)
		return
	}
	c.JSON(http.StatusCreated, result)
}

// SubmitChoice godoc
//
//	@Summary		Сделать выбор в попытке
//	@Description	Отправляет выбор пользователя для текущей сцены.
//	@Tags			attempts
//	@Accept			json
//	@Produce		json
//	@Security		SessionID
//	@Param			attemptID	path		string			true	"ID попытки"
//	@Param			request		body		ChoiceRequest	true	"Выбор пользователя"
//	@Success		200			{object}	map[string]interface{}
//	@Failure		400			{object}	ErrorResponse
//	@Failure		401			{object}	ErrorResponse
//	@Failure		403			{object}	ErrorResponse
//	@Failure		404			{object}	ErrorResponse
//	@Failure		409			{object}	ErrorResponse
//	@Failure		500			{object}	ErrorResponse
//	@Router			/v1/attempts/{attemptID}/choice [post]
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

	result, err := s.training.Choose(c.Request.Context(), userID, attemptID, req.SceneID, req.OptionID, *req.ExpectedRevision)
	if err != nil {
		if s.logger != nil {
			s.logger.Error("submit choice failed", zap.Error(err))
		}
		writeTrainingError(c, err)
		return
	}
	c.JSON(http.StatusOK, result)
}

// GetSummary godoc
//
//	@Summary		Получить результат попытки
//	@Description	Возвращает итоговую информацию по попытке пользователя.
//	@Tags			attempts
//	@Produce		json
//	@Security		SessionID
//	@Param			attemptID	path		string	true	"ID попытки"
//	@Success		200			{object}	map[string]interface{}
//	@Failure		400			{object}	ErrorResponse
//	@Failure		401			{object}	ErrorResponse
//	@Failure		403			{object}	ErrorResponse
//	@Failure		404			{object}	ErrorResponse
//	@Failure		500			{object}	ErrorResponse
//	@Router			/v1/attempts/{attemptID}/summary [get]
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
		writeTrainingError(c, err)
		return
	}
	c.JSON(http.StatusOK, summary)
}
