package httptransport

import (
	"errors"
	"net/http"

	domainErrors "github.com/34Minnesota/avito-antiscam-monorepo/backend/internal/domain/errors"
	"github.com/gin-gonic/gin"
)

type ErrorResponse struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

func writeTrainingError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, domainErrors.ErrNotFound),
		errors.Is(err, domainErrors.ErrForbidden):
		c.JSON(http.StatusNotFound, ErrorResponse{
			Code:    "not_found",
			Message: "resource not found",
		})

	case errors.Is(err, domainErrors.ErrConflict),
		errors.Is(err, domainErrors.ErrAttemptFinished):
		c.JSON(http.StatusConflict, ErrorResponse{
			Code:    "conflict",
			Message: "attempt is already finished",
		})

	case errors.Is(err, domainErrors.ErrOutOfOrder):
		c.JSON(http.StatusUnprocessableEntity, ErrorResponse{
			Code:    "out_of_order",
			Message: "choice does not match current scene",
		})

	case errors.Is(err, domainErrors.ErrValidation):
		c.JSON(http.StatusUnprocessableEntity, ErrorResponse{
			Code:    "validation_error",
			Message: "request violates a business rule",
		})

	case errors.Is(err, domainErrors.ErrUnknownOption):
		c.JSON(http.StatusUnprocessableEntity, ErrorResponse{
			Code:    "unknown_option",
			Message: "option does not exist",
		})

	default:
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Code:    "internal_error",
			Message: "internal server error",
		})
	}
}
