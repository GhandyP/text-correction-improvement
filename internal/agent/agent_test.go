package agent_test

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/GhandyP/text-correction-improvement/internal/agent"
	"github.com/GhandyP/text-correction-improvement/internal/correction"
)

const goodResponse = `{
	"corrected": "Mira, el nene corrió muy rápido ayer y el %s quedó intacto.",
	"changes": [
		{"kind": "spelling", "original": "corrio", "revised": "corrió", "reason": "missing accent"},
		{"kind": "spelling", "original": "rapido", "revised": "rápido", "reason": "missing accent"},
		{"kind": "spelling", "original": "quedo", "revised": "quedó", "reason": "missing accent"}
	],
	"questions": []
}`

func correctionRequest() correction.Request {
	return correction.Request{
		Text:            "Mira, el nene corrio muy rapido ayer y el %s quedo intacto.",
		Language:        "es",
		Audience:        "general",
		PreserveVoice:   true,
		Level:           correction.LevelMinimal,
		ProtectedTokens: []string{"%s"},
	}
}

func countingGenerator(raw string, calls *int) agent.GenerateFunc {
	return func(context.Context, correction.Request) (string, error) {
		*calls++
		return raw, nil
	}
}

func requireCode(t *testing.T, err error, want correction.Code) *correction.Error {
	t.Helper()

	if err == nil {
		t.Fatalf("got nil error, want code %s", want)
	}
	var typed *correction.Error
	if !errors.As(err, &typed) {
		t.Fatalf("error = %v (%T), want *correction.Error", err, err)
	}
	if typed.Code != want {
		t.Fatalf("error code = %s, want %s (detail: %s)", typed.Code, want, typed.Detail)
	}
	if strings.TrimSpace(typed.Detail) == "" {
		t.Fatal("error detail is empty; a typed error must say what failed")
	}
	return typed
}

func TestNewRejectsAnEmptyAPIKey(t *testing.T) {
	t.Setenv("GOOGLE_API_KEY", "")

	_, err := agent.New(context.Background(), agent.Config{})
	requireCode(t, err, correction.CodeMissingAPIKey)
}

func TestNewAcceptsAnInjectedGeneratorWithoutAnAPIKey(t *testing.T) {
	t.Setenv("GOOGLE_API_KEY", "")
	calls := 0

	if _, err := agent.New(context.Background(), agent.Config{Generate: countingGenerator(goodResponse, &calls)}); err != nil {
		t.Fatalf("New() = %v, want nil for an injected generator", err)
	}
	if calls != 0 {
		t.Fatalf("New() called the model %d times, want 0: construction must not generate", calls)
	}
}

func TestNewBuildsTheADKAgentWithAnAPIKey(t *testing.T) {
	a, err := agent.New(context.Background(), agent.Config{APIKey: "test-key"})
	if err != nil {
		t.Fatalf("New() = %v, want nil", err)
	}
	if got := a.Model(); got != agent.DefaultModel {
		t.Fatalf("Model() = %q, want the default %q", got, agent.DefaultModel)
	}
}

func TestModelPrecedence(t *testing.T) {
	calls := 0
	t.Setenv("ADK_MODEL", "gemini-2.5-pro")

	a, err := agent.New(context.Background(), agent.Config{Generate: countingGenerator(goodResponse, &calls)})
	if err != nil {
		t.Fatalf("New() = %v, want nil", err)
	}
	if got := a.Model(); got != "gemini-2.5-pro" {
		t.Fatalf("Model() = %q, want the ADK_MODEL override", got)
	}

	a, err = agent.New(context.Background(), agent.Config{Model: "gemini-2.5-flash", Generate: countingGenerator(goodResponse, &calls)})
	if err != nil {
		t.Fatalf("New() = %v, want nil", err)
	}
	if got := a.Model(); got != "gemini-2.5-flash" {
		t.Fatalf("Model() = %q, want the explicit config model to win over ADK_MODEL", got)
	}
}

func TestCorrectReturnsAVerifiedDraft(t *testing.T) {
	calls := 0
	a, err := agent.New(context.Background(), agent.Config{Generate: countingGenerator(goodResponse, &calls)})
	if err != nil {
		t.Fatalf("New() = %v, want nil", err)
	}

	got, err := a.Correct(context.Background(), correctionRequest())
	if err != nil {
		t.Fatalf("Correct() = %v, want nil", err)
	}
	if calls != 1 {
		t.Fatalf("the model was called %d times, want 1", calls)
	}
	if got.RevisedText != "Mira, el nene corrió muy rápido ayer y el %s quedó intacto." {
		t.Errorf("RevisedText = %q, want the corrected draft", got.RevisedText)
	}
	if len(got.Changes) != 3 {
		t.Errorf("len(Changes) = %d, want 3", len(got.Changes))
	}
}

func TestCorrectRejectsAnInvalidRequestBeforeCallingTheModel(t *testing.T) {
	calls := 0
	a, err := agent.New(context.Background(), agent.Config{Generate: countingGenerator(goodResponse, &calls)})
	if err != nil {
		t.Fatalf("New() = %v, want nil", err)
	}

	empty := correctionRequest()
	empty.Text = "   "
	requireCode(t, mustCorrect(t, a, empty), correction.CodeEmptyText)

	unprotectable := correctionRequest()
	unprotectable.ProtectedTokens = []string{"{{nombre}}"}
	requireCode(t, mustCorrect(t, a, unprotectable), correction.CodeProtectedTokenNotInText)

	if calls != 0 {
		t.Fatalf("the model was called %d times for invalid requests, want 0", calls)
	}
}

func TestCorrectReportsAModelFailure(t *testing.T) {
	boom := errors.New("model exploded")
	a, err := agent.New(context.Background(), agent.Config{
		Generate: func(context.Context, correction.Request) (string, error) { return "", boom },
	})
	if err != nil {
		t.Fatalf("New() = %v, want nil", err)
	}

	_, err = a.Correct(context.Background(), correctionRequest())
	requireCode(t, err, correction.CodeModelCallFailed)
	if !errors.Is(err, boom) {
		t.Fatalf("Correct() error = %v, want it to wrap the model failure", err)
	}
}

func TestCorrectReportsAMalformedResponse(t *testing.T) {
	calls := 0
	a, err := agent.New(context.Background(), agent.Config{Generate: countingGenerator("no json here", &calls)})
	if err != nil {
		t.Fatalf("New() = %v, want nil", err)
	}

	requireCode(t, mustCorrect(t, a, correctionRequest()), correction.CodeInvalidJSON)
}

func TestCorrectBlocksAnUnverifiedDraft(t *testing.T) {
	// The model drops the protected placeholder, so Go must refuse the draft.
	calls := 0
	raw := `{"corrected": "Mira, el nene corrió muy rápido ayer y el quedó intacto.", "changes": [], "questions": []}`
	a, err := agent.New(context.Background(), agent.Config{Generate: countingGenerator(raw, &calls)})
	if err != nil {
		t.Fatalf("New() = %v, want nil", err)
	}

	got, err := a.Correct(context.Background(), correctionRequest())

	var blocked *correction.BlockedError
	if !errors.As(err, &blocked) {
		t.Fatalf("Correct() error = %v (%T), want *correction.BlockedError", err, err)
	}
	if len(blocked.Findings) != 2 {
		t.Fatalf("BlockedError findings = %v, want both the lost token and the unexplained draft", blocked.Findings)
	}
	if blocked.Findings[0].Code != correction.CodeProtectedTokenLost || blocked.Findings[1].Code != correction.CodeMissingChanges {
		t.Fatalf("BlockedError findings = %v, want protected_token_lost then missing_changes in stable order", blocked.Findings)
	}
	if !strings.Contains(err.Error(), string(correction.CodeProtectedTokenLost)) {
		t.Fatalf("BlockedError message %q does not name the finding code", err)
	}
	if got.RevisedText != "" || len(got.Changes) != 0 || len(got.Questions) != 0 {
		t.Fatalf("Correct() returned %+v, want no result: an unverified draft must not escape", got)
	}
}

func mustCorrect(t *testing.T, a *agent.Agent, req correction.Request) error {
	t.Helper()

	if _, err := a.Correct(context.Background(), req); err != nil {
		return err
	}
	t.Fatal("Correct() = nil error, want a failure")
	return nil
}
