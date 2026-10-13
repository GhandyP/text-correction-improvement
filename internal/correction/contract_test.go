package correction_test

import (
	"slices"
	"testing"

	"github.com/GhandyP/text-correction-improvement/internal/correction"
)

func TestChangeKindsIsClosedAndFresh(t *testing.T) {
	want := []correction.ChangeKind{
		correction.KindSpelling,
		correction.KindGrammar,
		correction.KindPunctuation,
		correction.KindStyle,
		correction.KindClarity,
	}

	got := correction.ChangeKinds()
	if !slices.Equal(got, want) {
		t.Fatalf("ChangeKinds() = %v, want %v", got, want)
	}

	got[0] = "mutated"
	if again := correction.ChangeKinds(); !slices.Equal(again, want) {
		t.Fatalf("ChangeKinds() after caller mutation = %v, want a fresh %v", again, want)
	}
}

func TestLevelsIsClosedAndFresh(t *testing.T) {
	want := []correction.Level{correction.LevelMinimal, correction.LevelModerate}

	got := correction.Levels()
	if !slices.Equal(got, want) {
		t.Fatalf("Levels() = %v, want %v", got, want)
	}

	got[0] = "mutated"
	if again := correction.Levels(); !slices.Equal(again, want) {
		t.Fatalf("Levels() after caller mutation = %v, want a fresh %v", again, want)
	}
}

func TestChangeKindValidAcceptsOnlyTheClosedSet(t *testing.T) {
	for _, kind := range correction.ChangeKinds() {
		if !kind.Valid() {
			t.Errorf("ChangeKind(%q).Valid() = false, want true", kind)
		}
	}

	for _, kind := range []correction.ChangeKind{"", "typo", "Spelling", "rewrite"} {
		if kind.Valid() {
			t.Errorf("ChangeKind(%q).Valid() = true, want false", kind)
		}
	}
}

func TestLevelValidAcceptsOnlyTheClosedSet(t *testing.T) {
	for _, level := range correction.Levels() {
		if !level.Valid() {
			t.Errorf("Level(%q).Valid() = false, want true", level)
		}
	}

	for _, level := range []correction.Level{"", "light", "Minimal", "aggressive"} {
		if level.Valid() {
			t.Errorf("Level(%q).Valid() = true, want false", level)
		}
	}
}
