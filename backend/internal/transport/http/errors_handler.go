package httptransport

import (
	"errors"
	"net/http"

	domainErrors "github.com/34Minnesota/avito-antiscam-monorepo/backend/internal/domain/errors"
	"github.com/gin-gonic/gin"
)

type errorResponse struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

// Функция writeTrainingError отображает ошибки, которые намеренно выявляются в сценариях обучения.
// Ошибки, вызванные некорректным сценарием или неожиданной зависимостью, остаются с кодом 500.
func writeTrainingError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, domainErrors.ErrNotFound),
		errors.Is(err, domainErrors.ErrForbidden):
		c.JSON(http.StatusNotFound, errorResponse{
			Code:    "not_found",
			Message: "resource not found",
		})

	case errors.Is(err, domainErrors.ErrConflict),
		errors.Is(err, domainErrors.ErrAttemptFinished):
		c.JSON(http.StatusConflict, errorResponse{
			Code:    "conflict",
			Message: "attempt is already finished",
		})

	case errors.Is(err, domainErrors.ErrOutOfOrder):
		c.JSON(http.StatusUnprocessableEntity, errorResponse{
			Code:    "out_of_order",
			Message: "choice does not match current scene",
		})

	case errors.Is(err, domainErrors.ErrValidation):
		c.JSON(http.StatusUnprocessableEntity, errorResponse{
			Code:    "validation_error",
			Message: "request violates a business rule",
		})

	case errors.Is(err, domainErrors.ErrUnknownOption):
		c.JSON(http.StatusUnprocessableEntity, errorResponse{
			Code:    "unknown_option",
			Message: "option does not exist",
		})

	default:
		c.JSON(http.StatusInternalServerError, errorResponse{
			Code:    "internal_error",
			Message: "internal server error",
		})
	}
}
