package httptransport

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"

	domainErrors "github.com/34Minnesota/avito-antiscam-monorepo/backend/internal/domain/errors"
)

func TestWriteTrainingError(t *testing.T) {
	gin.SetMode(gin.TestMode)

	for _, test := range []struct {
		name       string
		err        error
		wantStatus int
		wantCode   string
	}{
		{name: "not found", err: domainErrors.ErrNotFound, wantStatus: http.StatusNotFound, wantCode: "not_found"},
		{name: "forbidden", err: domainErrors.ErrForbidden, wantStatus: http.StatusNotFound, wantCode: "not_found"},
		{name: "revision conflict", err: domainErrors.ErrConflict, wantStatus: http.StatusConflict, wantCode: "conflict"},
		{name: "finished attempt", err: domainErrors.ErrAttemptFinished, wantStatus: http.StatusConflict, wantCode: "conflict"},
		{name: "out of order", err: domainErrors.ErrOutOfOrder, wantStatus: http.StatusUnprocessableEntity, wantCode: "out_of_order"},
		{name: "unknown option", err: domainErrors.ErrUnknownOption, wantStatus: http.StatusUnprocessableEntity, wantCode: "unknown_option"},
		{name: "validation", err: domainErrors.ErrValidation, wantStatus: http.StatusUnprocessableEntity, wantCode: "validation_error"},
		{name: "invalid scenario", err: domainErrors.ErrInvalidScenario, wantStatus: http.StatusInternalServerError, wantCode: "internal_error"},
		{name: "unexpected", err: errors.New("database is unavailable"), wantStatus: http.StatusInternalServerError, wantCode: "internal_error"},
	} {
		t.Run(test.name, func(t *testing.T) {
			recorder := httptest.NewRecorder()
			ctx, _ := gin.CreateTestContext(recorder)

			writeTrainingError(ctx, test.err)

			if recorder.Code != test.wantStatus {
				t.Fatalf("status = %d, want %d", recorder.Code, test.wantStatus)
			}

			var response ErrorResponse
			if err := json.Unmarshal(recorder.Body.Bytes(), &response); err != nil {
				t.Fatal(err)
			}
			if response.Code != test.wantCode {
				t.Fatalf("code = %q, want %q", response.Code, test.wantCode)
			}
			if response.Message == "" {
				t.Fatal("error response must contain a public message")
			}
		})
	}
}
