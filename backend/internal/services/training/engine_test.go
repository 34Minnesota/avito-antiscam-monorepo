package training_test

import (
	"errors"
	"testing"

	"github.com/34Minnesota/avito-antiscam-monorepo/backend/internal/domain"
	domainErrors "github.com/34Minnesota/avito-antiscam-monorepo/backend/internal/domain/errors"
	"github.com/34Minnesota/avito-antiscam-monorepo/backend/internal/services/training"
)

func TestStartReturnsFirstSceneAndCleanState(t *testing.T) {
	t.Parallel()
	scene, state, err := training.Start(testDoc())
	if err != nil {
		t.Fatal(err)
	}
	if scene.SceneID != "s1" || len(scene.Options) != 3 {
		t.Fatalf("unexpected scene: %+v", scene)
	}
	if state.SceneIndex != 0 || state.Earned != 0 || len(state.Flags) != 0 {
		t.Fatalf("unexpected state: %+v", state)
	}
}

func TestStartRejectsScenarioWithoutScenes(t *testing.T) {
	t.Parallel()
	doc := testDoc()
	doc.Scenes = nil
	if _, _, err := training.Start(doc); !errors.Is(err, domainErrors.ErrInvalidScenario) {
		t.Fatalf("got %v", err)
	}
}

func TestAdvanceSafeOptionAddsWeightAndMovesOn(t *testing.T) {
	t.Parallel()
	doc := testDoc()
	tr, err := training.Advance(doc, domain.State{}, "s1", "a1")
	if err != nil {
		t.Fatal(err)
	}
	if tr.Finished || tr.NextScene == nil || tr.NextScene.SceneID != "s2" {
		t.Fatalf("unexpected transition: %+v", tr)
	}
	if tr.State.Earned != 1 || tr.State.SceneIndex != 1 || len(tr.State.Flags) != 0 {
		t.Fatalf("unexpected state: %+v", tr.State)
	}
	if tr.Verdict != domain.VerdictSafe || tr.Step.SceneID != "s1" || tr.Step.OptionID != "a1" {
		t.Fatalf("unexpected step: %+v", tr.Step)
	}
}

func TestAdvanceRiskyOptionRecordsFlagWithoutWeight(t *testing.T) {
	t.Parallel()
	tr, err := training.Advance(testDoc(), domain.State{}, "s1", "a2")
	if err != nil {
		t.Fatal(err)
	}
	if tr.State.Earned != 0 || len(tr.State.Flags) != 1 || tr.State.Flags[0] != "f1" {
		t.Fatalf("unexpected state: %+v", tr.State)
	}
}

func TestAdvanceDoesNotMutateInputState(t *testing.T) {
	t.Parallel()
	before := domain.State{SceneIndex: 0, Earned: 0, Flags: []string{"kept"}}
	if _, err := training.Advance(testDoc(), before, "s1", "a2"); err != nil {
		t.Fatal(err)
	}
	if before.SceneIndex != 0 || before.Earned != 0 || len(before.Flags) != 1 {
		t.Fatalf("input state was mutated: %+v", before)
	}
}

func TestAdvanceFatalOptionFinishesWithItsOwnEnding(t *testing.T) {
	t.Parallel()
	doc := testDoc()
	tr, err := training.Advance(doc, domain.State{}, "s1", "a3")
	if err != nil {
		t.Fatal(err)
	}
	if !tr.Finished || tr.EndingKey != "lost" || tr.Ending == nil {
		t.Fatalf("unexpected transition: %+v", tr)
	}
	if tr.Ending.Outcome != domain.OutcomeScammed {
		t.Fatalf("unexpected ending: %+v", tr.Ending)
	}
	if tr.State.SceneIndex != len(doc.Scenes) {
		t.Fatalf("fatal must jump past the last scene: %+v", tr.State)
	}
}

func TestAdvanceRejectsOutOfOrderScene(t *testing.T) {
	t.Parallel()
	_, err := training.Advance(testDoc(), domain.State{}, "s2", "b1")
	if !errors.Is(err, domainErrors.ErrOutOfOrder) {
		t.Fatalf("got %v", err)
	}
}

func TestAdvanceRejectsUnknownOption(t *testing.T) {
	t.Parallel()
	_, err := training.Advance(testDoc(), domain.State{}, "s1", "nope")
	if !errors.Is(err, domainErrors.ErrUnknownOption) {
		t.Fatalf("got %v", err)
	}
}

func TestAdvanceRejectsAlreadyFinishedAttempt(t *testing.T) {
	t.Parallel()
	doc := testDoc()
	for _, index := range []int{-1, len(doc.Scenes)} {
		_, err := training.Advance(doc, domain.State{SceneIndex: index}, "s1", "a1")
		if !errors.Is(err, domainErrors.ErrAttemptFinished) {
			t.Fatalf("index %d: got %v", index, err)
		}
	}
}

func TestAdvanceOnLastScenePicksEndingByFlags(t *testing.T) {
	t.Parallel()
	doc := testDoc()
	last := domain.State{SceneIndex: len(doc.Scenes) - 1}

	clean, err := training.Advance(doc, last, "s3", "c1")
	if err != nil {
		t.Fatal(err)
	}
	if !clean.Finished || clean.EndingKey != "safe" {
		t.Fatalf("clean run must end safe: %+v", clean)
	}

	flagged := last
	flagged.Flags = []string{"f1"}
	partial, err := training.Advance(doc, flagged, "s3", "c1")
	if err != nil {
		t.Fatal(err)
	}
	if !partial.Finished || partial.EndingKey != "partial" {
		t.Fatalf("flagged run must end partial: %+v", partial)
	}
}

func TestAdvanceRejectsMissingEnding(t *testing.T) {
	t.Parallel()

	fatalDoc := testDoc()
	delete(fatalDoc.Endings, "lost")
	if _, err := training.Advance(fatalDoc, domain.State{}, "s1", "a3"); !errors.Is(err, domainErrors.ErrInvalidScenario) {
		t.Fatalf("missing fatal ending: got %v", err)
	}

	finalDoc := testDoc()
	delete(finalDoc.Endings, "safe")
	last := domain.State{SceneIndex: len(finalDoc.Scenes) - 1}
	if _, err := training.Advance(finalDoc, last, "s3", "c1"); !errors.Is(err, domainErrors.ErrInvalidScenario) {
		t.Fatalf("missing final ending: got %v", err)
	}
}

func TestCurrentSceneID(t *testing.T) {
	t.Parallel()
	doc := testDoc()
	if got := training.CurrentSceneID(doc, domain.State{SceneIndex: 1}); got != "s2" {
		t.Fatalf("got %q", got)
	}
	for _, index := range []int{-1, len(doc.Scenes)} {
		if got := training.CurrentSceneID(doc, domain.State{SceneIndex: index}); got != "" {
			t.Fatalf("index %d: got %q", index, got)
		}
	}
}

func TestBuildSceneHidesServiceFields(t *testing.T) {
	t.Parallel()
	scene := testDoc().Scenes[0]
	payload := training.BuildScene(scene)

	if payload.SceneID != "s1" || payload.Prompt != "Что делать?" || len(payload.Intro) != 1 {
		t.Fatalf("unexpected payload: %+v", payload)
	}
	if len(payload.Options) != len(scene.Decision.Options) {
		t.Fatalf("options = %d", len(payload.Options))
	}
	for i, opt := range payload.Options {
		if opt.ID != scene.Decision.Options[i].ID || opt.Text != scene.Decision.Options[i].Text {
			t.Fatalf("option %d does not match source: %+v", i, opt)
		}
	}
}
