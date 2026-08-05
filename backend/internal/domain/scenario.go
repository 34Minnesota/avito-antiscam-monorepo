package domain

import "github.com/google/uuid"

// Role — за кого играет пользователь в сценарии.
type Role string

const (
	RoleBuyer  Role = "buyer"
	RoleSeller Role = "seller"
)

func (r Role) IsValid() bool { return r == RoleBuyer || r == RoleSeller }

// Verdict — оценка отдельного выбора пользователя.
//
//	risky — рискованное, но не фатальное: диалог продолжается, признак засчитан как пропущенный;
//	fatal — точка невозврата, сценарий немедленно завершается провальной концовкой.
type Verdict string

const (
	VerdictSafe  Verdict = "safe"  // safe  — безопасное действие;
	VerdictRisky Verdict = "risky" // risky — рискованное, но не фатальное
	VerdictFatal Verdict = "fatal" // fatal — точка невозврата
)

// Outcome — итог всего прохождения.
type Outcome string

const (
	OutcomeSafe    Outcome = "safe"
	OutcomePartial Outcome = "partial"
	OutcomeScammed Outcome = "scammed"
)

// Author - от чьего лица показано сообщение в чате.
type Author string

const (
	AuthorCounterpart Author = "counterpart" // собеседник (потенциальный мошенник)
	AuthorSystem      Author = "system"      // служебная вставка
	AuthorUser        Author = "user"        // реплика самого пользователя
)

// Scenario — строка каталога: метаданные в колонках + документ в JSONB.
type Scenario struct {
	ID       uuid.UUID
	IsActive bool
	Doc      ScenarioDoc
}

// ScenarioDoc — содержимое сценария целиком. Хранится в БД одним JSONB-документом
type ScenarioDoc struct {
	Version     int               `json:"version"`
	Slug        string            `json:"slug"`
	Role        Role              `json:"role"`
	Category    string            `json:"category"`
	Difficulty  int               `json:"difficulty"`
	Title       string            `json:"title"`
	Description string            `json:"description"`
	Listing     Listing           `json:"listing"`
	Counterpart Counterpart       `json:"counterpart"`
	Scenes      []Scene           `json:"scenes"`
	Endings     map[string]Ending `json:"endings"`
	Debrief     Debrief           `json:"debrief"`
}

// Attachment — вложение к сообщению: ссылка, картинка, чек.
type Attachment struct {
	Type  string `json:"type"`
	Title string `json:"title"`
}

// Message — одна реплика в диалоге.
type Message struct {
	Author     Author      `json:"author"`
	Text       string      `json:"text"`
	Attachment *Attachment `json:"attachment,omitempty"`
}

// Listing — карточка объявления, вокруг которого идёт сделка.
type Listing struct {
	Title    string `json:"title"`
	Price    int    `json:"price"`
	Location string `json:"location"`
	Image    string `json:"image"` // TODO: решить вопрос с image
}

// Counterpart — профиль собеседника.
type Counterpart struct {
	Name       string  `json:"name"`
	Rating     float64 `json:"rating"`
	Reviews    int     `json:"reviews"`
	Registered string  `json:"registered"`
}

// Option — вариант действия на развилке.
//
// Поля Verdict, Flag, Feedback и Ending — не видны юзеру
type Option struct {
	ID       string    `json:"id"`
	Text     string    `json:"text"`
	Verdict  Verdict   `json:"verdict"`
	Flag     string    `json:"flag,omitempty"`
	Reaction []Message `json:"reaction,omitempty"`
	Feedback string    `json:"feedback"`
	Ending   string    `json:"ending,omitempty"` // только для VerdictFatal
}

// Decision — развилка сцены.
type Decision struct {
	Prompt  string   `json:"prompt"`
	Options []Option `json:"options"`
}

// Scene — единица сценария: несколько реплик и один выбор.
// Weight задаёт вклад сцены в итоговый балл: чем важнее навык, тем больше вес.
type Scene struct {
	ID       string    `json:"id"`
	Weight   float64   `json:"weight"`
	Intro    []Message `json:"intro"`
	Decision Decision  `json:"decision"`
}

// Ending — вариант финала сценария.
type Ending struct {
	Outcome Outcome `json:"outcome"`
	Title   string  `json:"title"`
	Text    string  `json:"text"`
}

// FlagInfo — описание признака мошенничества для разбора.
type FlagInfo struct {
	ID    string `json:"id"`
	Title string `json:"title"`
	Text  string `json:"text"`
}

// Debrief — обучающая часть, показывается после прохождения.
type Debrief struct {
	KeyFlags []FlagInfo `json:"key_flags"`
	Takeaway string     `json:"takeaway"`
}

//------------------------ Вспомогательные методы ------------------------------

// SceneByID возвращает сцену по идентификатору.
func (d ScenarioDoc) SceneByID(id string) (Scene, bool) {
	for _, s := range d.Scenes {
		if s.ID == id {
			return s, true
		}
	}
	return Scene{}, false
}

// OptionByID возвращает вариант ответа внутри сцены.
func (s Scene) OptionByID(id string) (Option, bool) {
	for _, o := range s.Decision.Options {
		if o.ID == id {
			return o, true
		}
	}
	return Option{}, false
}

// TotalWeight — сумма весов всех сцен, знаменатель итогового балла.
func (d ScenarioDoc) TotalWeight() float64 {
	var total float64
	for _, s := range d.Scenes {
		total += s.Weight
	}
	return total
}

// FlagByID ищет описание признака в разборе сценария.
func (d ScenarioDoc) FlagByID(id string) (FlagInfo, bool) {
	for _, f := range d.Debrief.KeyFlags {
		if f.ID == id {
			return f, true
		}
	}
	return FlagInfo{}, false
}
