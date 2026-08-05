package training

import (
	"fmt"

	"github.com/34Minnesota/avito-antiscam-monorepo/backend/internal/domain"
	domainErrors "github.com/34Minnesota/avito-antiscam-monorepo/backend/internal/domain/errors"
)

// Проигрыватель сценариев тренажера

const (
	endingKeySafe    = "safe"
	endingKeyPartial = "partial"
)

// Transition — результат одного шага прохождения.
type Transition struct {
	Verdict   domain.Verdict
	Feedback  string
	Reaction  []domain.Message
	NextScene *ScenePayload
	Ending    *domain.Ending
	EndingKey string
	State     domain.State
	Step      domain.AttemptStep
	Finished  bool
}

// Start возвращает первую сцену сценария и начальное состояние.
func Start(doc domain.ScenarioDoc) (ScenePayload, domain.State, error) {
	if len(doc.Scenes) == 0 {
		return ScenePayload{}, domain.State{}, fmt.Errorf("%w: scenario has no scenes", domainErrors.ErrInvalidScenario)
	}
	state := domain.State{SceneIndex: 0, Earned: 0, Flags: []string{}}
	return BuildScene(doc.Scenes[0]), state, nil
}

// Advance обрабатывает выбор пользователя и возвращает следующее состояние диалога.
//
// Правила:
//  1. sceneID обязан совпадать с текущей сценой — иначе ErrOutOfOrder.
//     Это защита от двойного клика, кнопки «назад» и попыток отправить
//     выбор для сцены, до которой пользователь ещё не дошёл.
//  2. safe  — начисляем вес сцены и идём дальше;
//     risky — веса нет, признак записан как пропущенный, диалог продолжается;
//     fatal — сценарий обрывается концовкой, указанной в самой опции.
//  3. Когда сцены закончились, концовка выбирается по накопленным флагам:
//     нет пропущенных признаков — "safe", есть — "partial".
func Advance(doc domain.ScenarioDoc, st domain.State, sceneID, optionID string) (Transition, error) {
	if st.SceneIndex < 0 || st.SceneIndex >= len(doc.Scenes) {
		return Transition{}, domainErrors.ErrAttemptFinished
	}

	scene := doc.Scenes[st.SceneIndex]
	if scene.ID != sceneID {
		return Transition{}, fmt.Errorf("%w: wanted %q, received %q", domainErrors.ErrOutOfOrder, scene.ID, sceneID)
	}

	opt, ok := scene.OptionByID(optionID)
	if !ok {
		return Transition{}, fmt.Errorf("%w: %q in scene %q", domainErrors.ErrUnknownOption, optionID, sceneID)
	}

	next := domain.State{
		SceneIndex: st.SceneIndex,
		Earned:     st.Earned,
		Flags:      append([]string(nil), st.Flags...),
	}

	if opt.Verdict == domain.VerdictSafe {
		next.Earned += scene.Weight
	} else if opt.Flag != "" {
		next.Flags = append(next.Flags, opt.Flag)
	}

	tr := Transition{
		Verdict:  opt.Verdict,
		Feedback: opt.Feedback,
		Reaction: opt.Reaction,
		Step: domain.AttemptStep{
			SceneID:  scene.ID,
			OptionID: opt.ID,
			Verdict:  opt.Verdict,
			Flag:     opt.Flag,
			Weight:   scene.Weight,
		},
	}

	if opt.Verdict == domain.VerdictFatal {
		ending, found := doc.Endings[opt.Ending]
		if !found {
			return Transition{}, fmt.Errorf("%w: ending %q not found", domainErrors.ErrInvalidScenario, opt.Ending)
		}
		next.SceneIndex = len(doc.Scenes)
		tr.Finished = true
		tr.Ending = &ending
		tr.EndingKey = opt.Ending
		tr.State = next
		return tr, nil
	}

	next.SceneIndex++

	// Подбор концовки
	if next.SceneIndex >= len(doc.Scenes) {
		key := endingKeySafe
		if len(next.Flags) > 0 {
			key = endingKeyPartial
		}
		ending, found := doc.Endings[key]
		if !found {
			return Transition{}, fmt.Errorf("%w: ending %q not found", domainErrors.ErrInvalidScenario, key)
		}
		tr.Finished = true
		tr.Ending = &ending
		tr.EndingKey = key
		tr.State = next
		return tr, nil
	}

	payload := BuildScene(doc.Scenes[next.SceneIndex])
	tr.NextScene = &payload
	tr.State = next
	return tr, nil
}

// CurrentSceneID возвращает идентификатор сцены, на которой стоит пользователь
func CurrentSceneID(doc domain.ScenarioDoc, st domain.State) string {
	if st.SceneIndex < 0 || st.SceneIndex >= len(doc.Scenes) {
		return ""
	}
	return doc.Scenes[st.SceneIndex].ID
}
