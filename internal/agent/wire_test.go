package agent_test

import (
	"errors"
	"strings"
	"testing"

	"github.com/GhandyP/text-correction-improvement/internal/agent"
	"github.com/GhandyP/text-correction-improvement/internal/correction"
)

const validResponse = `{
	"corrected": "Mira, el nene corrió muy rápido ayer y el %s quedó intacto.",
	"changes": [
		{"kind": "spelling", "original": "corrio", "revised": "corrió", "reason": "missing accent"}
	],
	"questions": []
}`

func requireDecodeError(t *testing.T, raw string, want correction.Code, detail string) {
	t.Helper()

	_, err := agent.DecodeResult([]byte(raw))
	if err == nil {
		t.Fatalf("DecodeResult(%s) = nil error, want code %s", raw, want)
	}

	var typed *correction.Error
	if !errors.As(err, &typed) {
		t.Fatalf("DecodeResult() error = %v (%T), want *correction.Error", err, err)
	}
	if typed.Code != want {
		t.Fatalf("DecodeResult() error code = %s, want %s (detail: %s)", typed.Code, want, typed.Detail)
	}
	if detail != "" && !strings.Contains(err.Error(), detail) {
		t.Fatalf("DecodeResult() error %q does not mention %q", err, detail)
	}
}

func TestDecodeResultAcceptsAWellFormedResponse(t *testing.T) {
	got, err := agent.DecodeResult([]byte(validResponse))
	if err != nil {
		t.Fatalf("DecodeResult() = %v, want nil", err)
	}

	if got.RevisedText != "Mira, el nene corrió muy rápido ayer y el %s quedó intacto." {
		t.Errorf("RevisedText = %q, want the corrected draft", got.RevisedText)
	}
	if len(got.Changes) != 1 {
		t.Fatalf("len(Changes) = %d, want 1", len(got.Changes))
	}
	want := correction.Change{
		Kind:     correction.KindSpelling,
		Original: "corrio",
		Revised:  "corrió",
		Reason:   "missing accent",
	}
	if got.Changes[0] != want {
		t.Errorf("Changes[0] = %+v, want %+v", got.Changes[0], want)
	}
	if len(got.Questions) != 0 {
		t.Errorf("Questions = %v, want empty", got.Questions)
	}
}

func TestDecodeResultAcceptsAnEmptyChangeList(t *testing.T) {
	raw := `{"corrected": "Mira, el nene corrio.", "changes": [], "questions": []}`

	got, err := agent.DecodeResult([]byte(raw))
	if err != nil {
		t.Fatalf("DecodeResult() = %v, want nil for an already-correct draft", err)
	}
	if got.RevisedText != "Mira, el nene corrio." || len(got.Changes) != 0 {
		t.Fatalf("DecodeResult() = %+v, want the draft with no changes", got)
	}
}

func TestDecodeResultMapsQuestions(t *testing.T) {
	raw := `{"corrected": "x", "changes": [], "questions": ["¿informal o formal?"]}`

	got, err := agent.DecodeResult([]byte(raw))
	if err != nil {
		t.Fatalf("DecodeResult() = %v, want nil", err)
	}
	if len(got.Questions) != 1 || got.Questions[0] != "¿informal o formal?" {
		t.Fatalf("Questions = %v, want the model's question preserved verbatim", got.Questions)
	}
}

func TestDecodeResultRejectsMalformedJSON(t *testing.T) {
	requireDecodeError(t, `{"corrected": "x", "changes": [`, correction.CodeInvalidJSON, "")
	requireDecodeError(t, ``, correction.CodeInvalidJSON, "")
	requireDecodeError(t, `["corrected"]`, correction.CodeInvalidJSON, "")
}

func TestDecodeResultRejectsTrailingContent(t *testing.T) {
	requireDecodeError(t, `{"corrected": "x", "changes": [], "questions": []} and more`, correction.CodeInvalidJSON, "")
	requireDecodeError(t, `{"corrected": "x", "changes": [], "questions": []}{}`, correction.CodeInvalidJSON, "")
}

func TestDecodeResultRejectsMarkdownFence(t *testing.T) {
	fenced := "```json\n" + validResponse + "\n```"
	requireDecodeError(t, fenced, correction.CodeInvalidJSON, "")
}

func TestDecodeResultRejectsUnknownTopLevelField(t *testing.T) {
	raw := `{"corrected": "x", "changes": [], "questions": [], "notes": []}`
	requireDecodeError(t, raw, correction.CodeUnknownField, "notes")
}

func TestDecodeResultRejectsUnknownChangeField(t *testing.T) {
	raw := `{"corrected": "x", "changes": [{"kind": "spelling", "original": "a", "revised": "b", "reason": "r", "confidence": 0.9}], "questions": []}`
	requireDecodeError(t, raw, correction.CodeUnknownField, "confidence")
}

func TestDecodeResultIsCaseSensitiveAboutKeys(t *testing.T) {
	// The key is missing because the response spelled it in another case; the
	// error must say so instead of silently accepting a drifted contract.
	raw := `{"Corrected": "x", "changes": [], "questions": []}`

	_, err := agent.DecodeResult([]byte(raw))
	var typed *correction.Error
	if !errors.As(err, &typed) || typed.Code != correction.CodeMissingField {
		t.Fatalf("DecodeResult() error = %v, want a missing_field error", err)
	}
	if !strings.Contains(typed.Detail, "corrected") || !strings.Contains(typed.Detail, "Corrected") {
		t.Fatalf("DecodeResult() detail = %q, want it to name both the required and the received key", typed.Detail)
	}
}

func TestDecodeResultRejectsMissingTopLevelFields(t *testing.T) {
	cases := map[string]string{
		"corrected": `{"changes": [], "questions": []}`,
		"changes":   `{"corrected": "x", "questions": []}`,
		"questions": `{"corrected": "x", "changes": []}`,
	}
	for field, raw := range cases {
		requireDecodeError(t, raw, correction.CodeMissingField, field)
	}
}

func TestDecodeResultRejectsNullTopLevelField(t *testing.T) {
	for _, body := range []string{
		`{"corrected": null, "changes": [], "questions": []}`,
		`{"corrected": "x", "changes": null, "questions": []}`,
		`{"corrected": "x", "changes": [], "questions": null}`,
	} {
		requireDecodeError(t, body, correction.CodeMissingField, "")
	}
}

func TestDecodeResultRejectsMissingChangeField(t *testing.T) {
	for _, body := range []string{
		`{"kind": "spelling", "original": "a", "revised": "b"}`,
		`{"kind": "spelling", "original": "a", "reason": "r"}`,
		`{"kind": "spelling", "revised": "b", "reason": "r"}`,
		`{"original": "a", "revised": "b", "reason": "r"}`,
	} {
		raw := `{"corrected": "x", "changes": [` + body + `], "questions": []}`
		requireDecodeError(t, raw, correction.CodeMissingField, "changes[0]")
	}
}

func TestDecodeResultRejectsWrongJSONType(t *testing.T) {
	requireDecodeError(t, `{"corrected": 7, "changes": [], "questions": []}`, correction.CodeInvalidJSON, "corrected")
	requireDecodeError(t, `{"corrected": "x", "changes": {}, "questions": []}`, correction.CodeInvalidJSON, "changes")
	requireDecodeError(t, `{"corrected": "x", "changes": [], "questions": [7]}`, correction.CodeInvalidJSON, "questions")
}

func TestDecodeResultReportsTheFirstProblemInDocumentedOrder(t *testing.T) {
	requireDecodeError(t, `{"questions": []}`, correction.CodeMissingField, "corrected")
	requireDecodeError(t, `{"corrected": "x"}`, correction.CodeMissingField, "changes")
}

func TestDecodeResultDoesNotVerifyTheDraft(t *testing.T) {
	// Verification is Verify's job: a decodable response with a broken reason
	// must still decode, so the two boundaries stay separate.
	raw := `{"corrected": "x", "changes": [{"kind": "spelling", "original": "a", "revised": "b", "reason": "  "}], "questions": []}`

	if _, err := agent.DecodeResult([]byte(raw)); err != nil {
		t.Fatalf("DecodeResult() = %v, want nil: decoding must not verify", err)
	}
}
