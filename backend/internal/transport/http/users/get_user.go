package users_http_transport

import (
	"errors"
	"net/http"

	domainErrors "github.com/34Minnesota/avito-antiscam-monorepo/backend/internal/domain/errors"
	core_utils "github.com/34Minnesota/avito-antiscam-monorepo/backend/internal/utils"
	"github.com/gin-gonic/gin"
)

type GetUserResponse UserDTOResponse

func (h *UsersHandler) GetUser(c *gin.Context) {
	userID, err := core_utils.GetUUIDPathVal(c, "id")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": "bad_request", "message": "failed to fetch userID"})
		return
	}

	user, err := h.usersService.GetUser(c.Request.Context(), userID)
	if err != nil {
		switch {
		case errors.Is(err, domainErrors.ErrInvalidArgument):
			c.JSON(http.StatusBadRequest, gin.H{"code": "bad_request", "message": "invalid request"})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"code": "internal_error", "message": "internal server error"})
			return
		}
	}

	response := GetUserResponse(UserDTOResponse{
		ID:        user.ID,
		Nickname:  user.Nickname,
		Email:     user.Email,
		CreatedAt: user.CreatedAt,
	})

	c.JSON(http.StatusOK, response)
}
