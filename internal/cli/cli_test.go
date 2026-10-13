package cli_test

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/GhandyP/text-correction-improvement/internal/cli"
	"github.com/GhandyP/text-correction-improvement/internal/correction"
)

type fakeCorrector struct {
	calls  int
	got    correction.Request
	result correction.Result
	err    error
}

func (f *fakeCorrector) Correct(_ context.Context, req correction.Request) (correction.Result, error) {
	f.calls++
	f.got = req
	return f.result, f.err
}

// builder records the model name the CLI asked for.
func builder(f *fakeCorrector, model *string, buildErr error) cli.BuildFunc {
	return func(_ context.Context, name string) (cli.Corrector, error) {
		if model != nil {
			*model = name
		}
		if buildErr != nil {
			return nil, buildErr
		}
		return f, nil
	}
}

func verifiedResult() correction.Result {
	return correction.Result{
		RevisedText: "Mira, el nene corrió muy rápido.",
		Changes: []correction.Change{
			{Kind: correction.KindSpelling, Original: "corrio", Revised: "corrió", Reason: "missing accent"},
			{Kind: correction.KindSpelling, Original: "rapido", Revised: "rápido", Reason: "missing accent"},
		},
		Questions: []string{"¿Debe mantenerse el tono informal?"},
	}
}

type runResult struct {
	code   int
	stdout string
	stderr string
}

func run(t *testing.T, f *fakeCorrector, stdin string, buildErr error, args ...string) runResult {
	t.Helper()

	var stdout, stderr strings.Builder
	code := cli.Run(context.Background(), builder(f, nil, buildErr), args, strings.NewReader(stdin), &stdout, &stderr)
	return runResult{code: code, stdout: stdout.String(), stderr: stderr.String()}
}

func TestRunDeliversTheDraftToStdoutAndTheReportToStderr(t *testing.T) {
	f := &fakeCorrector{result: verifiedResult()}

	got := run(t, f, "", nil, "Mira, el nene corrio muy rapido.")

	if got.code != 0 {
		t.Fatalf("exit code = %d, want 0 (stderr: %s)", got.code, got.stderr)
	}
	if got.stdout != "Mira, el nene corrió muy rápido.\n" {
		t.Errorf("stdout = %q, want the draft and nothing else", got.stdout)
	}
	if !strings.Contains(got.stderr, "spelling") || !strings.Contains(got.stderr, "missing accent") {
		t.Errorf("stderr = %q, want the change report", got.stderr)
	}
	if !strings.Contains(got.stderr, "¿Debe mantenerse el tono informal?") {
		t.Errorf("stderr = %q, want the open questions", got.stderr)
	}
	if f.calls != 1 {
		t.Errorf("the corrector was called %d times, want 1", f.calls)
	}
	if f.got.Text != "Mira, el nene corrio muy rapido." {
		t.Errorf("request text = %q, want the positional argument", f.got.Text)
	}
}

func TestRunReportsAnAlreadyCorrectText(t *testing.T) {
	f := &fakeCorrector{result: correction.Result{RevisedText: "Mira, el nene corrio muy rapido."}}

	got := run(t, f, "", nil, "Mira, el nene corrio muy rapido.")

	if got.code != 0 {
		t.Fatalf("exit code = %d, want 0", got.code)
	}
	if !strings.Contains(got.stderr, "no changes") {
		t.Errorf("stderr = %q, want it to say the text needed no change", got.stderr)
	}
}

func TestRunReadsTheTextFromAFile(t *testing.T) {
	f := &fakeCorrector{result: verifiedResult()}
	path := filepath.Join(t.TempDir(), "draft.txt")
	if err := os.WriteFile(path, []byte("Mira, el nene corrio.\n"), 0o600); err != nil {
		t.Fatal(err)
	}

	got := run(t, f, "", nil, "-file", path)

	if got.code != 0 {
		t.Fatalf("exit code = %d, want 0 (stderr: %s)", got.code, got.stderr)
	}
	if f.got.Text != "Mira, el nene corrio.\n" {
		t.Errorf("request text = %q, want the file content verbatim", f.got.Text)
	}
}

func TestRunReadsTheTextFromStdin(t *testing.T) {
	f := &fakeCorrector{result: verifiedResult()}

	got := run(t, f, "Mira, el nene corrio.", nil, "-file", "-")

	if got.code != 0 {
		t.Fatalf("exit code = %d, want 0 (stderr: %s)", got.code, got.stderr)
	}
	if f.got.Text != "Mira, el nene corrio." {
		t.Errorf("request text = %q, want the stdin content", f.got.Text)
	}
}

func TestRunReadsTheTextFromStdinWhenNoArgumentIsGiven(t *testing.T) {
	f := &fakeCorrector{result: verifiedResult()}

	got := run(t, f, "Mira, el nene corrio.", nil)

	if got.code != 0 {
		t.Fatalf("exit code = %d, want 0 (stderr: %s)", got.code, got.stderr)
	}
	if f.got.Text != "Mira, el nene corrio." {
		t.Errorf("request text = %q, want the piped text", f.got.Text)
	}
}

func TestRunCarriesTheContractFlags(t *testing.T) {
	f := &fakeCorrector{result: verifiedResult()}

	got := run(t, f, "", nil, "-language", "es-AR", "-audience", "clientes", "-level", "moderate",
		"-preserve-voice=false", "-protect", "%s", "-protect", "ACME", "Mira, el %s de ACME corrio.")

	if got.code != 0 {
		t.Fatalf("exit code = %d, want 0 (stderr: %s)", got.code, got.stderr)
	}
	want := correction.Request{
		Text:            "Mira, el %s de ACME corrio.",
		Language:        "es-AR",
		Audience:        "clientes",
		Level:           correction.LevelModerate,
		ProtectedTokens: []string{"%s", "ACME"},
	}
	if !reflect.DeepEqual(f.got, want) {
		t.Errorf("request = %+v, want %+v", f.got, want)
	}
}

func TestRunDefaultsToMinimalAndPreservedVoice(t *testing.T) {
	f := &fakeCorrector{result: verifiedResult()}

	if got := run(t, f, "", nil, "Mira, el nene corrio."); got.code != 0 {
		t.Fatalf("exit code = %d, want 0 (stderr: %s)", got.code, got.stderr)
	}
	if f.got.Level != correction.LevelMinimal || !f.got.PreserveVoice {
		t.Errorf("request = %+v, want level minimal with the voice preserved", f.got)
	}
}

func TestRunPassesTheModelToTheBuilder(t *testing.T) {
	f := &fakeCorrector{result: verifiedResult()}
	var model string

	var stdout, stderr strings.Builder
	code := cli.Run(context.Background(), builder(f, &model, nil), []string{"-model", "gemini-2.5-pro", "Mira, el nene corrio."},
		strings.NewReader(""), &stdout, &stderr)

	if code != 0 {
		t.Fatalf("exit code = %d, want 0 (stderr: %s)", code, stderr.String())
	}
	if model != "gemini-2.5-pro" {
		t.Fatalf("builder model = %q, want the -model flag value", model)
	}
}

func TestRunRejectsUsageErrors(t *testing.T) {
	cases := map[string][]string{
		"no input":              {},
		"argument and file":     {"-file", "x.txt", "some text"},
		"unknown level":         {"-level", "aggressive", "some text"},
		"unknown flag":          {"-bogus", "some text"},
		"empty positional text": {"   "},
		"too many arguments":    {"one", "two"},
		"missing file":          {"-file", filepath.Join(t.TempDir(), "absent.txt")},
	}

	for name, args := range cases {
		t.Run(name, func(t *testing.T) {
			f := &fakeCorrector{result: verifiedResult()}

			got := run(t, f, "", nil, args...)

			if got.code != 2 {
				t.Fatalf("exit code = %d, want 2 (stdout: %q, stderr: %q)", got.code, got.stdout, got.stderr)
			}
			if got.stdout != "" {
				t.Errorf("stdout = %q, want nothing on a usage error", got.stdout)
			}
			if strings.TrimSpace(got.stderr) == "" {
				t.Error("stderr is empty, want a usage message")
			}
			if f.calls != 0 {
				t.Errorf("the corrector was called %d times, want 0", f.calls)
			}
		})
	}
}

func TestRunPrintsHelpOnRequest(t *testing.T) {
	f := &fakeCorrector{result: verifiedResult()}

	got := run(t, f, "", nil, "-h")

	if got.code != 0 {
		t.Fatalf("exit code = %d, want 0 for -h", got.code)
	}
	if !strings.Contains(got.stdout, "Usage") {
		t.Errorf("stdout = %q, want the usage message", got.stdout)
	}
	if f.calls != 0 {
		t.Errorf("the corrector was called %d times, want 0 for -h", f.calls)
	}
}

func TestRunMapsRequestFailuresToUsageErrors(t *testing.T) {
	f := &fakeCorrector{err: &correction.Error{Code: correction.CodeProtectedTokenNotInText, Detail: "protected token %q does not appear in the text"}}

	got := run(t, f, "", nil, "-protect", "{{nombre}}", "Mira, el nene corrio.")

	if got.code != 2 {
		t.Fatalf("exit code = %d, want 2 for an invalid request", got.code)
	}
	if !strings.Contains(got.stderr, string(correction.CodeProtectedTokenNotInText)) {
		t.Errorf("stderr = %q, want the error code", got.stderr)
	}
	if got.stdout != "" {
		t.Errorf("stdout = %q, want nothing on a rejected request", got.stdout)
	}
}

func TestRunReportsABlockedDraft(t *testing.T) {
	findings := []correction.Finding{{
		Code:        correction.CodeProtectedTokenLost,
		ChangeIndex: -1,
		Detail:      `protected token "%s" does not appear in the revised draft`,
	}}
	f := &fakeCorrector{err: &correction.BlockedError{Findings: findings}}

	got := run(t, f, "", nil, "-protect", "%s", "Mira, el %s corrio.")

	if got.code != 1 {
		t.Fatalf("exit code = %d, want 1 for a blocked draft", got.code)
	}
	if got.stdout != "" {
		t.Errorf("stdout = %q, want no draft: a blocked draft must not be delivered", got.stdout)
	}
	if !strings.Contains(got.stderr, string(correction.CodeProtectedTokenLost)) {
		t.Errorf("stderr = %q, want the finding code", got.stderr)
	}
}

func TestRunReportsAModelFailure(t *testing.T) {
	f := &fakeCorrector{err: &correction.Error{Code: correction.CodeModelCallFailed, Detail: "the model call failed: boom"}}

	got := run(t, f, "", nil, "Mira, el nene corrio.")

	if got.code != 1 {
		t.Fatalf("exit code = %d, want 1 for a model failure", got.code)
	}
	if !strings.Contains(got.stderr, string(correction.CodeModelCallFailed)) {
		t.Errorf("stderr = %q, want the error code", got.stderr)
	}
	if got.stdout != "" {
		t.Errorf("stdout = %q, want nothing on a failed run", got.stdout)
	}
}

func TestRunReportsABuildFailure(t *testing.T) {
	f := &fakeCorrector{result: verifiedResult()}
	buildErr := &correction.Error{Code: correction.CodeMissingAPIKey, Detail: "set GOOGLE_API_KEY"}

	got := run(t, f, "", buildErr, "Mira, el nene corrio.")

	if got.code != 1 {
		t.Fatalf("exit code = %d, want 1 when the agent cannot be built", got.code)
	}
	if !strings.Contains(got.stderr, string(correction.CodeMissingAPIKey)) {
		t.Errorf("stderr = %q, want the error code", got.stderr)
	}
	if f.calls != 0 {
		t.Errorf("the corrector was called %d times, want 0", f.calls)
	}
}

func TestRunEmitsTheContractAsJSON(t *testing.T) {
	f := &fakeCorrector{result: verifiedResult()}

	got := run(t, f, "", nil, "-json", "Mira, el nene corrio muy rapido.")

	if got.code != 0 {
		t.Fatalf("exit code = %d, want 0 (stderr: %s)", got.code, got.stderr)
	}

	var payload struct {
		Corrected string `json:"corrected"`
		Changes   []struct {
			Kind     string `json:"kind"`
			Original string `json:"original"`
			Revised  string `json:"revised"`
			Reason   string `json:"reason"`
		} `json:"changes"`
		Questions []string `json:"questions"`
	}
	if err := json.Unmarshal([]byte(got.stdout), &payload); err != nil {
		t.Fatalf("stdout is not JSON: %v (%q)", err, got.stdout)
	}
	if payload.Corrected != "Mira, el nene corrió muy rápido." {
		t.Errorf("corrected = %q, want the draft", payload.Corrected)
	}
	if len(payload.Changes) != 2 || payload.Changes[0].Kind != "spelling" {
		t.Errorf("changes = %+v, want the two spelling changes", payload.Changes)
	}
	if len(payload.Questions) != 1 {
		t.Errorf("questions = %v, want the open question", payload.Questions)
	}
	if got.stderr != "" {
		t.Errorf("stderr = %q, want nothing: in JSON mode the report travels in the payload", got.stderr)
	}
}

func TestRunEmitsFailuresAsJSONOnStderr(t *testing.T) {
	f := &fakeCorrector{err: &correction.BlockedError{Findings: []correction.Finding{
		{Code: correction.CodeEmptyReason, ChangeIndex: 1, Detail: "a change must state why it was made"},
	}}}

	got := run(t, f, "", nil, "-json", "Mira, el nene corrio.")

	if got.code != 1 {
		t.Fatalf("exit code = %d, want 1", got.code)
	}
	if got.stdout != "" {
		t.Errorf("stdout = %q, want nothing on failure", got.stdout)
	}

	var payload struct {
		Code     string `json:"code"`
		Detail   string `json:"detail"`
		Findings []struct {
			Code        string `json:"code"`
			ChangeIndex int    `json:"changeIndex"`
			Detail      string `json:"detail"`
		} `json:"findings"`
	}
	if err := json.Unmarshal([]byte(got.stderr), &payload); err != nil {
		t.Fatalf("stderr is not JSON: %v (%q)", err, got.stderr)
	}
	if payload.Code != string(correction.CodeUnverifiedDraft) {
		t.Errorf("code = %q, want %s", payload.Code, correction.CodeUnverifiedDraft)
	}
	if len(payload.Findings) != 1 || payload.Findings[0].Code != string(correction.CodeEmptyReason) || payload.Findings[0].ChangeIndex != 1 {
		t.Errorf("findings = %+v, want the blocked finding with its change index", payload.Findings)
	}
}

func TestRunReportsABlockedDraftAsJSONWithItsIndex(t *testing.T) {
	f := &fakeCorrector{err: &correction.BlockedError{Findings: []correction.Finding{
		{Code: correction.CodeRevisedNotInDraft, ChangeIndex: 2, Detail: `revised "veloz" does not appear in the revised draft`},
	}}}

	got := run(t, f, "", nil, "-json", "Mira, el nene corrio.")

	if got.code != 1 {
		t.Fatalf("exit code = %d, want 1", got.code)
	}
	if !strings.Contains(got.stderr, string(correction.CodeRevisedNotInDraft)) {
		t.Errorf("stderr = %q, want the finding code", got.stderr)
	}
	if !strings.Contains(got.stderr, "2") {
		t.Errorf("stderr = %q, want the change index of the finding", got.stderr)
	}
}
