package httptransport

import (
	"context"
	"errors"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	domainErrors "github.com/34Minnesota/avito-antiscam-monorepo/backend/internal/domain/errors"
	"github.com/34Minnesota/avito-antiscam-monorepo/backend/internal/domain/models"
)

// TODO: все теги под snake_case

type ProgressGetter interface {
	Get(context.Context, models.UserID) (models.OverallProgress, error)
}

type ProgressHandler struct {
	service ProgressGetter
}

func NewProgressHandler(service ProgressGetter) *ProgressHandler {
	return &ProgressHandler{service: service}
}

func RegisterProgressRoutes(router gin.IRoutes, handler *ProgressHandler) {
	router.GET("/progress", handler.Get)
}

// Get godoc
//
//	@Summary		Получить прогресс пользователя
//	@Description	Возвращает общий прогресс текущего авторизованного пользователя по сценариям.
//	@Tags			progress
//	@Produce		json
//	@Security		SessionID
//	@Success		200	{object}	progressResponse
//	@Failure		401	{object}	ErrorResponse
//	@Failure		500	{object}	ErrorResponse
//	@Failure		503	{object}	ErrorResponse
//	@Router			/v1/progress [get]
func (h *ProgressHandler) Get(c *gin.Context) {
	userID, ok := CurrentUser(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, ErrorResponse{
			Code:    "unauthorized",
			Message: "authenticated user is required",
		})
		return
	}

	if h == nil || h.service == nil {
		c.JSON(http.StatusServiceUnavailable, ErrorResponse{
			Code:    "dependency_unavailable",
			Message: "progress service is unavailable",
		})
		return
	}

	result, err := h.service.Get(c.Request.Context(), userID)
	if err != nil {
		switch {
		case errors.Is(err, domainErrors.ErrDependencyUnavailable):
			c.JSON(http.StatusServiceUnavailable, ErrorResponse{
				Code:    "dependency_unavailable",
				Message: "progress dependency is unavailable",
			})
		default:
			c.JSON(http.StatusInternalServerError, ErrorResponse{
				Code:    "internal_error",
				Message: "internal server error",
			})
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

type completedAttemptResultResponse struct {
	AttemptID   uuid.UUID      `json:"attemptId"`
	Score       scoreResponse  `json:"score"`
	Outcome     models.Outcome `json:"outcome"`
	CompletedAt time.Time      `json:"completedAt"`
}

type progressTrendResponse string

type scenarioProgressResponse struct {
	ScenarioSlug             string                           `json:"scenarioSlug"`
	Title                    string                           `json:"title"`
	Completed                bool                             `json:"completed"`
	Passed                   bool                             `json:"passed"`
	AttemptsCount            int                              `json:"attemptsCount"`
	BestScore                *scoreResponse                   `json:"bestScore"`
	ActiveAttemptID          *uuid.UUID                       `json:"activeAttemptId"`
	RecentAttempts           []completedAttemptResultResponse `json:"recentAttempts"`
	InitialScore             *scoreResponse                   `json:"initialScore"`
	LatestScore              *scoreResponse                   `json:"latestScore"`
	ImprovementPercentPoints *int                             `json:"improvementPercentPoints"`
	Trend                    *progressTrendResponse           `json:"trend"`
	FirstSafeAttempt         *completedAttemptResultResponse  `json:"firstSafeAttempt"`
}

type roleProgressResponse struct {
	Role               models.Role                `json:"role"`
	TotalScenarios     int                        `json:"totalScenarios"`
	CompletedScenarios int                        `json:"completedScenarios"`
	PassedScenarios    int                        `json:"passedScenarios"`
	CompletionPercent  int                        `json:"completionPercent"`
	PassedPercent      int                        `json:"passedPercent"`
	Scenarios          []scenarioProgressResponse `json:"scenarios"`
}

type roleComparisonResponse struct {
	CompletionPercentDelta int `json:"completionPercentDelta"`
	PassedPercentDelta     int `json:"passedPercentDelta"`
}

type recommendationResponse struct {
	ScenarioSlug string `json:"scenarioSlug"`
	ReasonCode   string `json:"reasonCode"`
	ReasonText   string `json:"reasonText"`
}

type progressResponse struct {
	TotalScenarios     int                      `json:"totalScenarios"`
	CompletedScenarios int                      `json:"completedScenarios"`
	PassedScenarios    int                      `json:"passedScenarios"`
	CompletionPercent  int                      `json:"completionPercent"`
	PassedPercent      int                      `json:"passedPercent"`
	Roles              []roleProgressResponse   `json:"roles"`
	RoleComparison     roleComparisonResponse   `json:"roleComparison"`
	Recommendations    []recommendationResponse `json:"recommendations"`
}

func mapProgress(progress models.OverallProgress) progressResponse {
	roles := make([]roleProgressResponse, 0, len(progress.Roles))
	for _, role := range progress.Roles {
		roles = append(roles, mapRole(role))
	}

	recommendations := make([]recommendationResponse, 0, len(progress.Recommendations))
	for _, recommendation := range progress.Recommendations {
		recommendations = append(recommendations, mapRecommendation(recommendation))
	}

	return progressResponse{
		TotalScenarios:     progress.TotalScenarios,
		CompletedScenarios: progress.CompletedScenarios,
		PassedScenarios:    progress.PassedScenarios,
		CompletionPercent:  progress.CompletionPercent,
		PassedPercent:      progress.PassedPercent,
		Roles:              roles,
		RoleComparison: roleComparisonResponse{
			CompletionPercentDelta: progress.RoleComparison.CompletionPercentDelta,
			PassedPercentDelta:     progress.RoleComparison.PassedPercentDelta,
		},
		Recommendations: recommendations,
	}
}

func mapRecommendation(recommendation models.Recommendation) recommendationResponse {
	return recommendationResponse{
		ScenarioSlug: recommendation.ScenarioSlug,
		ReasonCode:   recommendation.ReasonCode,
		ReasonText:   recommendation.ReasonText,
	}
}

func mapRole(role models.RoleProgress) roleProgressResponse {
	scenarios := make([]scenarioProgressResponse, 0, len(role.Scenarios))
	for _, scenario := range role.Scenarios {
		scenarios = append(scenarios, mapScenario(scenario))
	}

	return roleProgressResponse{
		Role:               role.Role,
		TotalScenarios:     role.TotalScenarios,
		CompletedScenarios: role.CompletedScenarios,
		PassedScenarios:    role.PassedScenarios,
		CompletionPercent:  role.CompletionPercent,
		PassedPercent:      role.PassedPercent,
		Scenarios:          scenarios,
	}
}

func mapScenario(scenario models.ScenarioProgress) scenarioProgressResponse {
	attempts := make([]completedAttemptResultResponse, 0, len(scenario.RecentAttempts))
	for _, attempt := range scenario.RecentAttempts {
		attempts = append(attempts, completedAttemptResultResponse{
			AttemptID:   attempt.ID,
			Score:       mapScore(attempt.Score),
			Outcome:     attempt.Outcome,
			CompletedAt: attempt.CompletedAt,
		})
	}

	return scenarioProgressResponse{
		ScenarioSlug:             scenario.Slug,
		Title:                    scenario.Title,
		Completed:                scenario.Completed,
		Passed:                   scenario.Passed,
		AttemptsCount:            scenario.AttemptsCount,
		BestScore:                mapOptionalScore(scenario.BestScore),
		ActiveAttemptID:          scenario.ActiveAttemptID,
		RecentAttempts:           attempts,
		InitialScore:             mapOptionalScore(scenario.InitialScore),
		LatestScore:              mapOptionalScore(scenario.LatestScore),
		ImprovementPercentPoints: scenario.ImprovementPercentPoints,
		Trend:                    (*progressTrendResponse)(scenario.Trend),
		FirstSafeAttempt:         mapOptionalAttempt(scenario.FirstSafeAttempt),
	}
}

func mapOptionalAttempt(attempt *models.AttemptResult) *completedAttemptResultResponse {
	if attempt == nil {
		return nil
	}

	return &completedAttemptResultResponse{
		AttemptID:   attempt.ID,
		Score:       mapScore(attempt.Score),
		Outcome:     attempt.Outcome,
		CompletedAt: attempt.CompletedAt,
	}
}

func mapOptionalScore(score *models.Score) *scoreResponse {
	if score == nil {
		return nil
	}

	mapped := mapScore(*score)
	return &mapped
}

func mapScore(score models.Score) scoreResponse {
	return scoreResponse{
		Points:    score.Points(),
		MaxPoints: score.MaxPoints(),
		Percent:   score.Percent(),
	}
}
