package httptransport

import (
	"context"
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/34Minnesota/avito-antiscam-monorepo/backend/generated/openapi"
	"github.com/34Minnesota/avito-antiscam-monorepo/backend/internal/domain"
	progressservice "github.com/34Minnesota/avito-antiscam-monorepo/backend/internal/services/progress"
	"github.com/34Minnesota/avito-antiscam-monorepo/backend/internal/transport/http/identity"
)

// тут вайбкод пока в ветке нет хендлера ручек

type ProgressGetter interface {
	Get(context.Context, domain.UserID) (domain.OverallProgress, error)
}

type ProgressHandler struct {
	service ProgressGetter
}

func NewProgressHandler(service ProgressGetter) *ProgressHandler {
	return &ProgressHandler{service: service}
}

// RegisterProgressRoutes ожидает маршрутизатор, защищённый middleware авторизации.
// TODO(owner: auth): подключить JWT-middleware к этой группе маршрутов при настройке приложения.
func RegisterProgressRoutes(router gin.IRoutes, handler *ProgressHandler) {
	router.GET("/v1/progress", handler.Get)
}

func (h *ProgressHandler) Get(c *gin.Context) {
	userID, ok := identity.UserID(c.Request.Context())
	if !ok {
		c.JSON(http.StatusUnauthorized, openapi.Error{Code: "unauthorized", Message: "authenticated user is required"})
		return
	}
	if h == nil || h.service == nil {
		c.JSON(http.StatusServiceUnavailable, openapi.Error{Code: "dependency_unavailable", Message: "progress service is unavailable"})
		return
	}
	result, err := h.service.Get(c.Request.Context(), userID)
	if err != nil {
		switch {
		case errors.Is(err, progressservice.ErrDependencyUnavailable):
			c.JSON(http.StatusServiceUnavailable, openapi.Error{Code: "dependency_unavailable", Message: "progress dependency is unavailable"})
		default:
			c.JSON(http.StatusInternalServerError, openapi.Error{Code: "internal_error", Message: "internal server error"})
		}
		return
	}
	c.JSON(http.StatusOK, mapProgress(result))
}

func mapProgress(progress domain.OverallProgress) openapi.ProgressResponse {
	roles := make([]openapi.RoleProgress, 0, len(progress.Roles))
	for _, role := range progress.Roles {
		roles = append(roles, mapRole(role))
	}
	outdated := make([]openapi.OutdatedActiveAttempt, 0, len(progress.OutdatedActiveAttempts))
	for _, attempt := range progress.OutdatedActiveAttempts {
		outdated = append(outdated, openapi.OutdatedActiveAttempt{
			AttemptId: attempt.AttemptID, ScenarioSlug: attempt.ScenarioSlug, ScenarioTitle: attempt.ScenarioTitle,
			Version: mapVersion(attempt.Version), CurrentVersion: mapVersion(attempt.CurrentVersion),
		})
	}
	return openapi.ProgressResponse{
		TotalScenarios: progress.TotalScenarios, CompletedScenarios: progress.CompletedScenarios,
		PassedScenarios: progress.PassedScenarios, CompletionPercent: progress.CompletionPercent,
		PassedPercent: progress.PassedPercent, Roles: roles, OutdatedActiveAttempts: outdated,
	}
}

func mapRole(role domain.RoleProgress) openapi.RoleProgress {
	scenarios := make([]openapi.ScenarioProgress, 0, len(role.Scenarios))
	for _, scenario := range role.Scenarios {
		scenarios = append(scenarios, mapScenario(scenario))
	}
	return openapi.RoleProgress{
		Role: openapi.RoleProgressRole(role.Role), TotalScenarios: role.TotalScenarios,
		CompletedScenarios: role.CompletedScenarios, PassedScenarios: role.PassedScenarios,
		CompletionPercent: role.CompletionPercent, PassedPercent: role.PassedPercent, Scenarios: scenarios,
	}
}

func mapScenario(scenario domain.ScenarioProgress) openapi.ScenarioProgress {
	history := make([]openapi.HistoricalVersionProgress, 0, len(scenario.History))
	for _, item := range scenario.History {
		history = append(history, openapi.HistoricalVersionProgress{
			Version: mapVersion(item.Version), Completed: item.Completed, Passed: item.Passed,
			AttemptsCount: item.AttemptsCount, BestScore: mapOptionalScore(item.BestScore),
		})
	}
	attempts := make([]openapi.CompletedAttemptResult, 0, len(scenario.RecentAttempts))
	for _, attempt := range scenario.RecentAttempts {
		attempts = append(attempts, openapi.CompletedAttemptResult{
			AttemptId: attempt.ID, Version: mapVersion(attempt.Version), Score: mapScore(attempt.Score),
			Passed: attempt.Passed, CompletedAt: attempt.CompletedAt,
		})
	}
	return openapi.ScenarioProgress{
		ScenarioSlug: scenario.Slug, Title: scenario.Title, CurrentVersion: mapVersion(scenario.Current.Version),
		CurrentProgress: openapi.CurrentVersionProgress{
			Completed: scenario.Current.Completed, Passed: scenario.Current.Passed,
			AttemptsCount: scenario.Current.AttemptsCount, BestScore: mapOptionalScore(scenario.Current.BestScore),
			ActiveAttemptId: scenario.Current.ActiveAttemptID,
		},
		History: history, RecentAttempts: attempts,
	}
}

func mapVersion(version domain.Version) openapi.ScenarioVersion {
	return openapi.ScenarioVersion{Number: version.Number, PassPercent: version.PassPercent, PublishedAt: version.PublishedAt}
}

func mapOptionalScore(score *domain.Score) *openapi.Score {
	if score == nil {
		return nil
	}
	mapped := mapScore(*score)
	return &mapped
}

func mapScore(score domain.Score) openapi.Score {
	return openapi.Score{Points: score.Points(), MaxPoints: score.MaxPoints(), Percent: score.Percent()}
}
