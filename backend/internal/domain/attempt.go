package domain

import (
	"time"

	"github.com/google/uuid"
)

// AttemptStatus — состояние попытки прохождения.
type AttemptStatus string

const (
	AttemptInProgress AttemptStatus = "in_progress"
	AttemptFinished   AttemptStatus = "finished"
)

// State — позиция пользователя внутри сценария.
// Хранится в attempts.state как JSONB.
type State struct {
	SceneIndex int      `json:"scene_index"`
	Earned     float64  `json:"earned"`
	Flags      []string `json:"flags"`
}

// Attempt — одно прохождение сценария пользователем.
type Attempt struct {
	ID           uuid.UUID
	UserID       uuid.UUID
	ScenarioID   uuid.UUID
	Status       AttemptStatus
	CurrentScene string
	State        State
	Score        *int
	Outcome      *Outcome
	StartedAt    time.Time
	FinishedAt   *time.Time
}

// AttemptStep — зафиксированный выбор пользователя.
// Полный журнал шагов — источник правды и для оценки, и для разбора.
type AttemptStep struct {
	AttemptID uuid.UUID
	SceneID   string
	OptionID  string
	Verdict   Verdict
	Flag      string
	Weight    float64
	CreatedAt time.Time
}
