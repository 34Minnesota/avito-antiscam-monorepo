package users_http_transport

import (
	"errors"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	coreErrors "github.com/34Minnesota/avito-antiscam-monorepo/backend/internal/errors"
)

type CreateUserRequest struct {
	Nickname string `json:"nickname" binding:"required,min=3,max=30"`
	Email    string `json:"email" binding:"required,email,max=254"`
	Password string `json:"password" binding:"required,min=8,max=72"`
}

type CreateUserResponse struct {
	ID        uuid.UUID `json:"id"`
	Nickname  string    `json:"nickname"`
	Email     string    `json:"email"`
	CreatedAt time.Time `json:"created_at"`
}

func (h *UsersHandler) CreateUser(c *gin.Context) {
	var request CreateUserRequest

	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": "bad_request", "message": "invalid request"})
		return
	}
	if h == nil || h.usersService == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": "internal_error", "message": "internal server error"})
		return
	}

	user, err := h.usersService.CreateUser(
		c.Request.Context(),
		CreateUserInput{
			Nickname: request.Nickname,
			Email:    request.Email,
			Password: request.Password,
		},
	)
	if err != nil {
		switch {
		case errors.Is(err, coreErrors.ErrInvalidArgument):
			c.JSON(http.StatusBadRequest, gin.H{"code": "bad_request", "message": "invalid request"})
		case errors.Is(err, coreErrors.ErrConflict):
			c.JSON(http.StatusConflict, gin.H{"code": "conflict", "message": "user already exists"})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"code": "internal_error", "message": "internal server error"})
		}
		return
	}

	response := CreateUserResponse{
		ID:        user.ID,
		Nickname:  user.Nickname,
		Email:     user.Email,
		CreatedAt: user.CreatedAt,
	}

	c.JSON(http.StatusCreated, response)
}
