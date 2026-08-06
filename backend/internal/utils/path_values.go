package core_utils

import (
	"fmt"

	core_errors "github.com/34Minnesota/avito-antiscam-monorepo/backend/internal/errors"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

func GetUUIDPathVal(c *gin.Context, key string) (uuid.UUID, error) {
	pathValue := c.Param(key)
	if pathValue == "" {
		return uuid.UUID{}, fmt.Errorf(
			"no key %s in path values: %w",
			key, core_errors.ErrInvalidArgument,
		)
	}

	val, err := uuid.Parse(pathValue)
	if err != nil {
		return uuid.UUID{}, fmt.Errorf(
			"path value %s by key %s not a valid uuid: %w: %w",
			pathValue, key, err, core_errors.ErrInvalidArgument,
		)
	}

	return val, nil
}
