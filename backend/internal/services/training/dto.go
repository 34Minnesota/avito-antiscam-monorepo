package training

import (
	"github.com/34Minnesota/avito-antiscam-monorepo/backend/internal/domain"
	"github.com/google/uuid"
)

// Структуры прохождения тренажера

// ScenarioStat — персональная статистика пользователя по сценарию для каталога.
type ScenarioStats struct {
	BestScore     int `json:"best_score"`
	AttemptsCount int `json:"attempts_count"`
}

// ScenarioCard — карточка каталога.
type ScenarioCard struct {
	ID          uuid.UUID      `json:"id"`
	Slug        string         `json:"slug"`
	Role        domain.Role    `json:"role"`
	Category    string         `json:"category"`
	Difficulty  int            `json:"difficulty"`
	Title       string         `json:"title"`
	Description string         `json:"description"`
	Stats       *ScenarioStats `json:"stats"`
}

// StartResult — ответ на старт прохождения.
type StartResult struct {
	AttemptID   uuid.UUID          `json:"attempt_id"`
	Listing     domain.Listing     `json:"listing"`
	Counterpart domain.Counterpart `json:"counterpart"`
	Role        domain.Role        `json:"role"`
	Title       string             `json:"title"`
	Scene       ScenePayload       `json:"scene"`
	ScenesTotal int                `json:"scenes_total"`
}

// Feedback — объяснение сделанного выбора.
type Feedback struct {
	Verdict domain.Verdict `json:"verdict"`
	Text    string         `json:"text"`
}

// ChoiceResult — ответ на выбор пользователя.
type ChoiceResult struct {
	Feedback  Feedback         `json:"feedback"`
	Reaction  []domain.Message `json:"reaction"`
	NextScene *ScenePayload    `json:"next_scene"`
	Finished  bool             `json:"finished"`
	//Summary   *scoring.Result  `json:"summary,omitempty"`
}

// Структуры клиентского представления

// OptionPayload — вариант ответа без служебных полей.
type OptionPayload struct {
	ID   string `json:"id"`
	Text string `json:"text"`
}

// ScenePayload — сцена в том виде, в каком её видит клиент.
type ScenePayload struct {
	SceneID string           `json:"scene_id"`
	Intro   []domain.Message `json:"intro"`
	Prompt  string           `json:"prompt"`
	Options []OptionPayload  `json:"options"`
}
