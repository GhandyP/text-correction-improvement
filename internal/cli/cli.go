// Package cli turns one command line into one correction run: it reads the
// text, builds the agent, and prints either the verified draft or the reason it
// was not delivered.
//
// Exit codes are part of the contract:
//
//	0  a verified draft was delivered
//	1  the run failed (model, decode, or a blocked draft)
//	2  the invocation was wrong (bad flags, bad input, invalid request)
package cli

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/GhandyP/text-correction-improvement/internal/correction"
)

// Exit codes.
const (
	ExitOK      = 0
	ExitFailure = 1
	ExitUsage   = 2
)

// Corrector corrects one request. agent.Agent satisfies it.
type Corrector interface {
	Correct(ctx context.Context, req correction.Request) (correction.Result, error)
}

// BuildFunc builds the corrector for one run. model is the -model flag value,
// empty when the flag was not given.
type BuildFunc func(ctx context.Context, model string) (Corrector, error)

// Run executes one invocation and returns the process exit code. Everything it
// touches is a parameter, so tests exercise the whole surface without a shell.
func Run(ctx context.Context, build BuildFunc, args []string, stdin io.Reader, stdout, stderr io.Writer) int {
	flags := flag.NewFlagSet("text-correction-improvement", flag.ContinueOnError)
	// The command renders its own usage, so the flag package stays silent.
	flags.SetOutput(io.Discard)
	flags.Usage = func() {}

	var (
		file          = flags.String("file", "", `read the text from this file ("-" for stdin)`)
		asJSON        = flags.Bool("json", false, "print the contract as JSON instead of text plus report")
		language      = flags.String("language", "", "language hint, for example es-AR")
		audience      = flags.String("audience", "", "audience hint, for example clients")
		level         = flags.String("level", string(correction.LevelMinimal), "how far to go: minimal or moderate")
		preserveVoice = flags.Bool("preserve-voice", true, "keep the author's voice and register")
		model         = flags.String("model", "", "ADK model name; empty means ADK_MODEL, then the default")
		protect       stringList
	)
	flags.Var(&protect, "protect", "a token that must survive the correction verbatim (repeatable)")

	if err := flags.Parse(args); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			fmt.Fprint(stdout, usage())
			return ExitOK
		}
		fmt.Fprintf(stderr, "usage error: %v\n\n%s", err, usage())
		return ExitUsage
	}

	text, err := resolveText(flags.Args(), *file, stdin)
	if err != nil {
		fmt.Fprintf(stderr, "usage error: %v\n\n%s", err, usage())
		return ExitUsage
	}

	requestedLevel := correction.Level(*level)
	if !requestedLevel.Valid() {
		fmt.Fprintf(stderr, "usage error: unknown level %q; use %s\n", *level, levelChoices())
		return ExitUsage
	}

	corrector, err := build(ctx, strings.TrimSpace(*model))
	if err != nil {
		report(stderr, *asJSON, err)
		return ExitFailure
	}

	result, err := corrector.Correct(ctx, correction.Request{
		Text:            text,
		Language:        strings.TrimSpace(*language),
		Audience:        strings.TrimSpace(*audience),
		PreserveVoice:   *preserveVoice,
		Level:           requestedLevel,
		ProtectedTokens: protect,
	})
	if err != nil {
		report(stderr, *asJSON, err)
		if isRequestFailure(err) {
			return ExitUsage
		}
		return ExitFailure
	}

	if err := emit(stdout, stderr, *asJSON, result); err != nil {
		// The draft or the report never reached the caller: a run that cannot
		// deliver its output must not report success. The failure report itself
		// stays best-effort, because the writer that just failed is stderr.
		report(stderr, *asJSON, err)
		return ExitFailure
	}
	return ExitOK
}

// resolveText picks the single source of the text: one argument, one file, or
// standard input.
func resolveText(args []string, file string, stdin io.Reader) (string, error) {
	switch {
	case len(args) > 1:
		return "", fmt.Errorf("give the text as one argument, not %d", len(args))
	case len(args) == 1 && file != "":
		return "", errors.New("give the text either as an argument or with -file, not both")
	case len(args) == 1:
		return requireText(args[0])
	case file != "" && file != "-":
		content, err := os.ReadFile(file)
		if err != nil {
			return "", fmt.Errorf("read %s: %w", file, err)
		}
		return requireText(string(content))
	default:
		content, err := io.ReadAll(stdin)
		if err != nil {
			return "", fmt.Errorf("read standard input: %w", err)
		}
		return requireText(string(content))
	}
}

func requireText(text string) (string, error) {
	if strings.TrimSpace(text) == "" {
		return "", errors.New("the text to correct is empty; pass it as an argument, with -file, or on standard input")
	}
	return text, nil
}

// emit prints the verified draft. Text mode keeps the deliverable alone on
// stdout and sends the review report to stderr, so the draft stays pipeable.
// Every write is checked: a delivery that never reached the caller is a
// failure, not a success.
func emit(stdout, stderr io.Writer, asJSON bool, result correction.Result) error {
	if asJSON {
		payload := resultPayload{
			Corrected: result.RevisedText,
			Changes:   make([]changePayload, 0, len(result.Changes)),
			Questions: append([]string{}, result.Questions...),
		}
		for _, change := range result.Changes {
			payload.Changes = append(payload.Changes, changePayload{
				Kind:     string(change.Kind),
				Original: change.Original,
				Revised:  change.Revised,
				Reason:   change.Reason,
			})
		}
		return writeJSON(stdout, "the draft", payload)
	}

	if err := writeText(stdout, "the draft", result.RevisedText+"\n"); err != nil {
		return err
	}
	return writeText(stderr, "the report", reportText(result))
}

// reportText renders the review report: every change with its reason, then the
// questions the agent would rather ask than guess.
func reportText(result correction.Result) string {
	var b strings.Builder

	if len(result.Changes) == 0 {
		b.WriteString("no changes: the text was already correct.\n")
	} else {
		fmt.Fprintf(&b, "changes (%d):\n", len(result.Changes))
		for index, change := range result.Changes {
			fmt.Fprintf(&b, "  %d. [%s] %q -> %q - %s\n", index+1, change.Kind, change.Original, change.Revised, change.Reason)
		}
	}
	if len(result.Questions) > 0 {
		fmt.Fprintf(&b, "questions (%d):\n", len(result.Questions))
		for index, question := range result.Questions {
			fmt.Fprintf(&b, "  %d. %s\n", index+1, question)
		}
	}

	return b.String()
}

// report explains a failure. Text mode prints the typed error and every
// finding; JSON mode prints one error object, still on stderr, so stdout never
// carries something that is not a verified draft.
func report(stderr io.Writer, asJSON bool, err error) {
	if asJSON {
		payload := errorPayload{Code: codeOf(err), Detail: err.Error()}
		var blocked *correction.BlockedError
		if errors.As(err, &blocked) {
			payload.Findings = make([]findingPayload, 0, len(blocked.Findings))
			for _, finding := range blocked.Findings {
				payload.Findings = append(payload.Findings, findingPayload{
					Code:        string(finding.Code),
					ChangeIndex: finding.ChangeIndex,
					Detail:      finding.Detail,
				})
			}
		}
		// Best effort: the writer that just failed is stderr itself, so there is
		// nowhere left to report a second failure.
		_ = writeJSON(stderr, "the error report", payload)
		return
	}

	fmt.Fprintln(stderr, err)
	var blocked *correction.BlockedError
	if errors.As(err, &blocked) {
		for _, finding := range blocked.Findings {
			if finding.ChangeIndex >= 0 {
				fmt.Fprintf(stderr, "  - [change %d] %s: %s\n", finding.ChangeIndex, finding.Code, finding.Detail)
				continue
			}
			fmt.Fprintf(stderr, "  - %s: %s\n", finding.Code, finding.Detail)
		}
	}
}

// isRequestFailure reports whether the failure is the caller's input rather
// than the model's answer.
func isRequestFailure(err error) bool {
	var typed *correction.Error
	if !errors.As(err, &typed) {
		return false
	}
	switch typed.Code {
	case correction.CodeEmptyText,
		correction.CodeUnknownLevel,
		correction.CodeEmptyProtectedToken,
		correction.CodeDuplicateProtectedToken,
		correction.CodeProtectedTokenNotInText:
		return true
	default:
		return false
	}
}

func codeOf(err error) string {
	var typed *correction.Error
	if errors.As(err, &typed) {
		return string(typed.Code)
	}
	var blocked *correction.BlockedError
	if errors.As(err, &blocked) {
		return string(correction.CodeUnverifiedDraft)
	}
	return "error"
}

type resultPayload struct {
	Corrected string          `json:"corrected"`
	Changes   []changePayload `json:"changes"`
	Questions []string        `json:"questions"`
}

type changePayload struct {
	Kind     string `json:"kind"`
	Original string `json:"original"`
	Revised  string `json:"revised"`
	Reason   string `json:"reason"`
}

type errorPayload struct {
	Code     string           `json:"code"`
	Detail   string           `json:"detail"`
	Findings []findingPayload `json:"findings,omitempty"`
}

type findingPayload struct {
	Code        string `json:"code"`
	ChangeIndex int    `json:"changeIndex"`
	Detail      string `json:"detail"`
}

// writeJSON encodes payload and writes it, turning a write failure into a
// typed error instead of dropping it.
func writeJSON(w io.Writer, what string, payload any) error {
	var buf bytes.Buffer
	encoder := json.NewEncoder(&buf)
	encoder.SetIndent("", "  ")
	encoder.SetEscapeHTML(false)
	if err := encoder.Encode(payload); err != nil {
		return &correction.Error{
			Code:   correction.CodeOutputWriteFailed,
			Detail: fmt.Sprintf("encoding %s failed: %v", what, err),
			Cause:  err,
		}
	}
	return writeText(w, what, buf.String())
}

// writeText writes text and turns any failure into a typed error, so no caller
// can claim success for output that never arrived.
func writeText(w io.Writer, what, text string) error {
	if _, err := io.WriteString(w, text); err != nil {
		return &correction.Error{
			Code:   correction.CodeOutputWriteFailed,
			Detail: fmt.Sprintf("writing %s failed: %v", what, err),
			Cause:  err,
		}
	}
	return nil
}

type stringList []string

func (l *stringList) String() string { return strings.Join(*l, ",") }

func (l *stringList) Set(value string) error {
	*l = append(*l, value)
	return nil
}

func levelChoices() string {
	levels := correction.Levels()
	names := make([]string, 0, len(levels))
	for _, level := range levels {
		names = append(names, string(level))
	}
	return strings.Join(names, " or ")
}

func usage() string {
	return `Usage: text-correction-improvement [flags] [text]

Corrects one text and prints the verified draft to stdout. The report — every
change with its reason and the questions the agent would rather ask than guess —
goes to stderr, so the draft stays pipeable. Reads the text from the argument,
from -file, or from standard input.

Flags:
  -file string      read the text from this file ("-" for stdin)
  -json             print the contract as JSON instead of text plus report
  -language string  language hint, for example es-AR
  -audience string  audience hint, for example clients
  -level string     how far to go: ` + levelChoices() + ` (default minimal)
  -preserve-voice   keep the author's voice and register (default true)
  -protect string   a token that must survive the correction verbatim (repeatable)
  -model string     ADK model name; empty means ADK_MODEL, then the default
  -h, -help         print this help

Environment:
  GOOGLE_API_KEY    Gemini credential (required unless a model is injected)
  ADK_MODEL         model override, for example gemini-2.5-pro

Exit codes:
  0  a verified draft was delivered
  1  the run failed: model error, undecodable response, or a blocked draft
  2  the invocation was wrong: bad flags, bad input, or an invalid request
`
}
