package users_http_transport

import (
	"context"

	"github.com/gin-gonic/gin"

	"github.com/34Minnesota/avito-antiscam-monorepo/backend/internal/domain"
	usersservice "github.com/34Minnesota/avito-antiscam-monorepo/backend/internal/services/users"
)

type UsersHandler struct {
	usersService UsersService
}

type CreateUserInput = usersservice.CreateUserInput

type UsersService interface {
	CreateUser(
		ctx context.Context,
		userInput CreateUserInput,
	) (domain.User, error)
}

func NewUsersHandler(usersService UsersService) *UsersHandler {
	return &UsersHandler{
		usersService: usersService,
	}
}

func RegisterUsersRoutes(router gin.IRoutes, handler *UsersHandler) {
	router.POST("/v1/users", handler.CreateUser)
}
