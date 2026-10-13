package correction_test

import (
	"slices"
	"strings"
	"testing"

	"github.com/GhandyP/text-correction-improvement/internal/correction"
)

const (
	sourceText = "Mira, el nene corrio muy rapido ayer y el %s quedo intacto."
	draftText  = "Mira, el nene corrió muy rápido ayer y el %s quedó intacto."
)

func verifiedRequest() correction.Request {
	return correction.Request{
		Text:            sourceText,
		Language:        "es",
		Audience:        "general",
		PreserveVoice:   true,
		Level:           correction.LevelMinimal,
		ProtectedTokens: []string{"%s"},
	}
}

func verifiedResult() correction.Result {
	return correction.Result{
		RevisedText: draftText,
		Changes: []correction.Change{
			{Kind: correction.KindSpelling, Original: "corrio", Revised: "corrió", Reason: "missing accent"},
			{Kind: correction.KindSpelling, Original: "rapido", Revised: "rápido", Reason: "missing accent"},
			{Kind: correction.KindSpelling, Original: "quedo", Revised: "quedó", Reason: "missing accent"},
		},
	}
}

type finding struct {
	Code  correction.Code
	Index int
}

func findingsOf(t *testing.T, result correction.Result) []finding {
	t.Helper()

	got := correction.Verify(verifiedRequest(), result)
	out := make([]finding, 0, len(got))
	for _, f := range got {
		out = append(out, finding{Code: f.Code, Index: f.ChangeIndex})
	}
	return out
}

func requireFindings(t *testing.T, result correction.Result, want ...finding) {
	t.Helper()

	if got := findingsOf(t, result); !slices.Equal(got, want) {
		t.Fatalf("Verify() findings = %v, want %v", got, want)
	}
}

func TestVerifyAcceptsAVerifiedDraft(t *testing.T) {
	if got := findingsOf(t, verifiedResult()); len(got) != 0 {
		t.Fatalf("Verify() findings = %v, want none", got)
	}
}

func TestVerifyAcceptsAnAlreadyCorrectNoOp(t *testing.T) {
	if got := findingsOf(t, correction.Result{RevisedText: sourceText}); len(got) != 0 {
		t.Fatalf("Verify() findings = %v, want none for an already-correct draft", got)
	}
}

func TestVerifyAcceptsWhitespaceOnlyDifferencesAsNoOp(t *testing.T) {
	result := correction.Result{RevisedText: "Mira, el nene\n   corrio muy rapido ayer y el %s quedo intacto."}
	if got := findingsOf(t, result); len(got) != 0 {
		t.Fatalf("Verify() findings = %v, want none when only whitespace changed", got)
	}
}

func TestVerifyReportsEmptyDraft(t *testing.T) {
	for _, draft := range []string{"", "   ", "\n\t"} {
		requireFindings(t, correction.Result{RevisedText: draft}, finding{Code: correction.CodeEmptyDraft, Index: -1})
	}
}

func TestVerifyReportsLostProtectedToken(t *testing.T) {
	result := verifiedResult()
	result.RevisedText = "Mira, el nene corrió muy rápido ayer y el quedó intacto."

	got := correction.Verify(verifiedRequest(), result)
	if len(got) != 1 {
		t.Fatalf("Verify() findings = %v, want exactly one protected_token_lost", got)
	}
	if got[0].Code != correction.CodeProtectedTokenLost || got[0].ChangeIndex != -1 {
		t.Fatalf("Verify() finding = %+v, want protected_token_lost at index -1", got[0])
	}
	if !strings.Contains(got[0].Detail, "%s") {
		t.Fatalf("finding detail %q does not name the lost token", got[0].Detail)
	}
}

func TestVerifyAcceptsProtectedTokenAcrossWhitespace(t *testing.T) {
	result := verifiedResult()
	result.RevisedText = "Mira,   el nene corrió muy rápido ayer y el\n%s quedó intacto."

	if got := findingsOf(t, result); len(got) != 0 {
		t.Fatalf("Verify() findings = %v, want none: the protected token survived with different whitespace", got)
	}
}

func TestVerifyReportsChangeOriginalAbsentFromText(t *testing.T) {
	result := verifiedResult()
	result.Changes[1].Original = "rápido"

	requireFindings(t, result, finding{Code: correction.CodeOriginalNotInText, Index: 1})
}

func TestVerifyReportsChangeRevisedAbsentFromDraft(t *testing.T) {
	result := verifiedResult()
	result.Changes[1].Revised = "veloz"

	requireFindings(t, result, finding{Code: correction.CodeRevisedNotInDraft, Index: 1})
}

func TestVerifyReportsUnknownChangeKind(t *testing.T) {
	result := verifiedResult()
	result.Changes[0].Kind = "rewrite"

	requireFindings(t, result, finding{Code: correction.CodeUnknownChangeKind, Index: 0})
}

func TestVerifyReportsEmptyReason(t *testing.T) {
	result := verifiedResult()
	result.Changes[2].Reason = "   "

	requireFindings(t, result, finding{Code: correction.CodeEmptyReason, Index: 2})
}

func TestVerifyReportsEmptyReplacementSide(t *testing.T) {
	result := verifiedResult()
	result.Changes[0].Original = ""

	requireFindings(t, result, finding{Code: correction.CodeReplacementSideEmpty, Index: 0})

	result = verifiedResult()
	result.Changes[1].Revised = ""

	requireFindings(t, result, finding{Code: correction.CodeReplacementSideEmpty, Index: 1})
}

func TestVerifyReportsDraftChangedWithoutChanges(t *testing.T) {
	requireFindings(t, correction.Result{RevisedText: draftText}, finding{Code: correction.CodeMissingChanges, Index: -1})
}

func TestVerifyReportsStaleChanges(t *testing.T) {
	// The draft is identical to the source and every listed change is locatable
	// in both, so staleness is the only problem left to report.
	result := correction.Result{
		RevisedText: sourceText,
		Changes: []correction.Change{
			{Kind: correction.KindSpelling, Original: "rapido", Revised: "rapido", Reason: "no-op"},
		},
	}

	requireFindings(t, result, finding{Code: correction.CodeStaleChanges, Index: -1})
}

func TestVerifyReportsEmptyQuestion(t *testing.T) {
	result := verifiedResult()
	result.Questions = []string{"¿Debe mantenerse el tono informal?", "   "}

	got := correction.Verify(verifiedRequest(), result)
	if len(got) != 1 {
		t.Fatalf("Verify() findings = %v, want exactly one empty_question", got)
	}
	if got[0].Code != correction.CodeEmptyQuestion || got[0].ChangeIndex != -1 {
		t.Fatalf("Verify() finding = %+v, want empty_question at index -1", got[0])
	}
	if !strings.Contains(got[0].Detail, "2") {
		t.Fatalf("finding detail %q does not name the empty question position", got[0].Detail)
	}
}

func TestVerifyReportsAllFindingsInStableOrder(t *testing.T) {
	result := correction.Result{
		RevisedText: "Mira, el nene corrió muy rápido ayer y el quedó intacto.",
		Changes: []correction.Change{
			{Kind: correction.KindSpelling, Original: "corrio", Revised: "corrió", Reason: ""},
			{Kind: "rewrite", Original: "rapido", Revised: "rápido", Reason: "missing accent"},
		},
		Questions: []string{"¿Debe mantenerse el tono informal?", ""},
	}

	requireFindings(t, result,
		finding{Code: correction.CodeProtectedTokenLost, Index: -1},
		finding{Code: correction.CodeEmptyReason, Index: 0},
		finding{Code: correction.CodeUnknownChangeKind, Index: 1},
		finding{Code: correction.CodeEmptyQuestion, Index: -1},
	)
}
