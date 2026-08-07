package users_http_transport

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"

	domainErrors "github.com/34Minnesota/avito-antiscam-monorepo/backend/internal/domain/errors"
)

type CreateUserRequest struct {
	Nickname string `json:"nickname" binding:"required,min=3,max=30"`
	Email    string `json:"email" binding:"required,email,max=254"`
	Password string `json:"password" binding:"required,min=8,max=72"`
}

type CreateUserResponse UserDTOResponse

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
		case errors.Is(err, domainErrors.ErrInvalidArgument):
			c.JSON(http.StatusBadRequest, gin.H{"code": "bad_request", "message": "invalid request"})
		case errors.Is(err, domainErrors.ErrConflict):
			c.JSON(http.StatusConflict, gin.H{"code": "conflict", "message": "user already exists"})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"code": "internal_error", "message": "internal server error"})
		}
		return
	}

	response := CreateUserResponse(UserDTOResponse{
		ID:        user.ID,
		Nickname:  user.Nickname,
		Email:     user.Email,
		CreatedAt: user.CreatedAt,
	})

	c.JSON(http.StatusCreated, response)
}
