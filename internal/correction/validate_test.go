package correction_test

import (
	"errors"
	"strings"
	"testing"

	"github.com/GhandyP/text-correction-improvement/internal/correction"
)

func wellFormedRequest() correction.Request {
	return correction.Request{
		Text:            "Mira, el nene corrio muy rapido ayer y el %s quedo intacto.",
		Language:        "es",
		Audience:        "general",
		PreserveVoice:   true,
		Level:           correction.LevelMinimal,
		ProtectedTokens: []string{"%s"},
	}
}

func requireErrorCode(t *testing.T, err error, want correction.Code) {
	t.Helper()

	if err == nil {
		t.Fatalf("ValidateRequest() = nil, want error with code %s", want)
	}

	var typed *correction.Error
	if !errors.As(err, &typed) {
		t.Fatalf("ValidateRequest() error = %v (%T), want *correction.Error", err, err)
	}
	if typed.Code != want {
		t.Fatalf("ValidateRequest() error code = %s, want %s (detail: %s)", typed.Code, want, typed.Detail)
	}
	if strings.TrimSpace(typed.Detail) == "" {
		t.Fatal("ValidateRequest() error detail is empty; a typed error must say what failed")
	}
	if !strings.Contains(err.Error(), string(want)) {
		t.Fatalf("error %q does not mention the code %s", err, want)
	}
}

func TestValidateRequestAcceptsAWellFormedRequest(t *testing.T) {
	if err := correction.ValidateRequest(wellFormedRequest()); err != nil {
		t.Fatalf("ValidateRequest() = %v, want nil", err)
	}
}

func TestValidateRequestRejectsEmptyText(t *testing.T) {
	for _, text := range []string{"", "   ", "\n\t "} {
		req := wellFormedRequest()
		req.Text = text
		requireErrorCode(t, correction.ValidateRequest(req), correction.CodeEmptyText)
	}
}

func TestValidateRequestRejectsUnknownLevel(t *testing.T) {
	for _, level := range []correction.Level{"", "light", "Minimal"} {
		req := wellFormedRequest()
		req.Level = level
		requireErrorCode(t, correction.ValidateRequest(req), correction.CodeUnknownLevel)
	}
}

func TestValidateRequestRejectsEmptyProtectedToken(t *testing.T) {
	for _, token := range []string{"", "   ", "\n"} {
		req := wellFormedRequest()
		req.ProtectedTokens = []string{token}
		requireErrorCode(t, correction.ValidateRequest(req), correction.CodeEmptyProtectedToken)
	}
}

func TestValidateRequestRejectsDuplicateProtectedTokens(t *testing.T) {
	req := wellFormedRequest()
	req.ProtectedTokens = []string{"%s", "%s"}
	requireErrorCode(t, correction.ValidateRequest(req), correction.CodeDuplicateProtectedToken)
}

func TestValidateRequestRejectsProtectedTokenAbsentFromText(t *testing.T) {
	req := wellFormedRequest()
	req.ProtectedTokens = []string{"{{nombre}}"}
	requireErrorCode(t, correction.ValidateRequest(req), correction.CodeProtectedTokenNotInText)
}

func TestValidateRequestAcceptsProtectedTokenAcrossWhitespace(t *testing.T) {
	req := wellFormedRequest()
	req.Text = "Mira, el nene\n   corrio muy rapido."
	req.ProtectedTokens = []string{"el nene corrio"}
	if err := correction.ValidateRequest(req); err != nil {
		t.Fatalf("ValidateRequest() = %v, want nil: a protected token must survive a line-break change", err)
	}
}
