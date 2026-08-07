package httptransport

import (
	"context"
	"errors"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/34Minnesota/avito-antiscam-monorepo/backend/internal/domain"
	domainErrors "github.com/34Minnesota/avito-antiscam-monorepo/backend/internal/domain/errors"
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
		c.JSON(http.StatusUnauthorized, errorResponse{Code: "unauthorized", Message: "authenticated user is required"})
		return
	}
	if h == nil || h.service == nil {
		c.JSON(http.StatusServiceUnavailable, errorResponse{Code: "dependency_unavailable", Message: "progress service is unavailable"})
		return
	}
	result, err := h.service.Get(c.Request.Context(), userID)
	if err != nil {
		switch {
		case errors.Is(err, domainErrors.ErrDependencyUnavailable):
			c.JSON(http.StatusServiceUnavailable, errorResponse{Code: "dependency_unavailable", Message: "progress dependency is unavailable"})
		default:
			c.JSON(http.StatusInternalServerError, errorResponse{Code: "internal_error", Message: "internal server error"})
		}
		return
	}
	c.JSON(http.StatusOK, mapProgress(result))
}

type scoreResponse struct {
	Points    int `json:"points"`
	MaxPoints int `json:"maxPoints"`
	Percent   int `json:"percent"`
}

type versionResponse struct {
	Number      int       `json:"number"`
	PassPercent int       `json:"passPercent"`
	PublishedAt time.Time `json:"publishedAt"`
}

type currentVersionProgressResponse struct {
	Completed       bool           `json:"completed"`
	Passed          bool           `json:"passed"`
	AttemptsCount   int            `json:"attemptsCount"`
	BestScore       *scoreResponse `json:"bestScore"`
	ActiveAttemptID *uuid.UUID     `json:"activeAttemptId"`
}

type historicalVersionProgressResponse struct {
	Version       versionResponse `json:"version"`
	Completed     bool            `json:"completed"`
	Passed        bool            `json:"passed"`
	AttemptsCount int             `json:"attemptsCount"`
	BestScore     *scoreResponse  `json:"bestScore"`
}

type completedAttemptResultResponse struct {
	AttemptID   uuid.UUID       `json:"attemptId"`
	Version     versionResponse `json:"version"`
	Score       scoreResponse   `json:"score"`
	Passed      bool            `json:"passed"`
	CompletedAt time.Time       `json:"completedAt"`
}

type scenarioProgressResponse struct {
	ScenarioSlug    string                              `json:"scenarioSlug"`
	Title           string                              `json:"title"`
	CurrentVersion  versionResponse                     `json:"currentVersion"`
	CurrentProgress currentVersionProgressResponse      `json:"currentProgress"`
	History         []historicalVersionProgressResponse `json:"history"`
	RecentAttempts  []completedAttemptResultResponse    `json:"recentAttempts"`
}

type roleProgressResponse struct {
	Role               domain.Role                `json:"role"`
	TotalScenarios     int                        `json:"totalScenarios"`
	CompletedScenarios int                        `json:"completedScenarios"`
	PassedScenarios    int                        `json:"passedScenarios"`
	CompletionPercent  int                        `json:"completionPercent"`
	PassedPercent      int                        `json:"passedPercent"`
	Scenarios          []scenarioProgressResponse `json:"scenarios"`
}

type outdatedActiveAttemptResponse struct {
	AttemptID      uuid.UUID       `json:"attemptId"`
	ScenarioSlug   string          `json:"scenarioSlug"`
	ScenarioTitle  string          `json:"scenarioTitle"`
	Version        versionResponse `json:"version"`
	CurrentVersion versionResponse `json:"currentVersion"`
}

type progressResponse struct {
	TotalScenarios         int                             `json:"totalScenarios"`
	CompletedScenarios     int                             `json:"completedScenarios"`
	PassedScenarios        int                             `json:"passedScenarios"`
	CompletionPercent      int                             `json:"completionPercent"`
	PassedPercent          int                             `json:"passedPercent"`
	Roles                  []roleProgressResponse          `json:"roles"`
	OutdatedActiveAttempts []outdatedActiveAttemptResponse `json:"outdatedActiveAttempts"`
}

func mapProgress(progress domain.OverallProgress) progressResponse {
	roles := make([]roleProgressResponse, 0, len(progress.Roles))
	for _, role := range progress.Roles {
		roles = append(roles, mapRole(role))
	}
	outdated := make([]outdatedActiveAttemptResponse, 0, len(progress.OutdatedActiveAttempts))
	for _, attempt := range progress.OutdatedActiveAttempts {
		outdated = append(outdated, outdatedActiveAttemptResponse{
			AttemptID: attempt.AttemptID, ScenarioSlug: attempt.ScenarioSlug, ScenarioTitle: attempt.ScenarioTitle,
			Version: mapVersion(attempt.Version), CurrentVersion: mapVersion(attempt.CurrentVersion),
		})
	}
	return progressResponse{
		TotalScenarios: progress.TotalScenarios, CompletedScenarios: progress.CompletedScenarios,
		PassedScenarios: progress.PassedScenarios, CompletionPercent: progress.CompletionPercent,
		PassedPercent: progress.PassedPercent, Roles: roles, OutdatedActiveAttempts: outdated,
	}
}

func mapRole(role domain.RoleProgress) roleProgressResponse {
	scenarios := make([]scenarioProgressResponse, 0, len(role.Scenarios))
	for _, scenario := range role.Scenarios {
		scenarios = append(scenarios, mapScenario(scenario))
	}
	return roleProgressResponse{
		Role: role.Role, TotalScenarios: role.TotalScenarios,
		CompletedScenarios: role.CompletedScenarios, PassedScenarios: role.PassedScenarios,
		CompletionPercent: role.CompletionPercent, PassedPercent: role.PassedPercent, Scenarios: scenarios,
	}
}

func mapScenario(scenario domain.ScenarioProgress) scenarioProgressResponse {
	history := make([]historicalVersionProgressResponse, 0, len(scenario.History))
	for _, item := range scenario.History {
		history = append(history, historicalVersionProgressResponse{
			Version: mapVersion(item.Version), Completed: item.Completed, Passed: item.Passed,
			AttemptsCount: item.AttemptsCount, BestScore: mapOptionalScore(item.BestScore),
		})
	}
	attempts := make([]completedAttemptResultResponse, 0, len(scenario.RecentAttempts))
	for _, attempt := range scenario.RecentAttempts {
		attempts = append(attempts, completedAttemptResultResponse{
			AttemptID: attempt.ID, Version: mapVersion(attempt.Version), Score: mapScore(attempt.Score),
			Passed: attempt.Passed, CompletedAt: attempt.CompletedAt,
		})
	}
	return scenarioProgressResponse{
		ScenarioSlug: scenario.Slug, Title: scenario.Title, CurrentVersion: mapVersion(scenario.Current.Version),
		CurrentProgress: currentVersionProgressResponse{
			Completed: scenario.Current.Completed, Passed: scenario.Current.Passed,
			AttemptsCount: scenario.Current.AttemptsCount, BestScore: mapOptionalScore(scenario.Current.BestScore),
			ActiveAttemptID: scenario.Current.ActiveAttemptID,
		},
		History: history, RecentAttempts: attempts,
	}
}

func mapVersion(version domain.Version) versionResponse {
	return versionResponse{Number: version.Number, PassPercent: version.PassPercent, PublishedAt: version.PublishedAt}
}

func mapOptionalScore(score *domain.Score) *scoreResponse {
	if score == nil {
		return nil
	}
	mapped := mapScore(*score)
	return &mapped
}

func mapScore(score domain.Score) scoreResponse {
	return scoreResponse{Points: score.Points(), MaxPoints: score.MaxPoints(), Percent: score.Percent()}
}
