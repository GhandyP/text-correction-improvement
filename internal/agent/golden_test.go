package agent_test

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"slices"
	"testing"

	"github.com/GhandyP/text-correction-improvement/internal/agent"
	"github.com/GhandyP/text-correction-improvement/internal/correction"
)

// caseRequest is the per-case part of a correction request; the text itself
// lives in input.txt so the fixture is the text a person would paste.
type caseRequest struct {
	Language        string   `json:"language"`
	Audience        string   `json:"audience"`
	PreserveVoice   bool     `json:"preserveVoice"`
	Level           string   `json:"level"`
	ProtectedTokens []string `json:"protectedTokens"`
}

type wantError struct {
	Code     string   `json:"code"`
	Findings []string `json:"findings"`
}

// TestGoldenCases replays every recorded model response through the real
// pipeline — validate, generate, decode, verify — with a fake generator, so the
// whole contract is exercised without a network, a credential, or a live model.
//
// Each case directory holds input.txt (the text as written), request.json (the
// caller's constraints), model.json (the recorded response shape), and either
// want.json (the expected verified draft) or want-error.json (the expected
// block). want.json is decoded by the same strict decoder, so a fixture cannot
// describe a draft the pipeline could not produce.
func TestGoldenCases(t *testing.T) {
	entries, err := os.ReadDir("testdata/cases")
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) < 6 {
		t.Fatalf("found %d cases, want at least 6", len(entries))
	}

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		t.Run(entry.Name(), func(t *testing.T) {
			base := filepath.Join("testdata/cases", entry.Name())

			text, err := os.ReadFile(filepath.Join(base, "input.txt"))
			if err != nil {
				t.Fatal(err)
			}
			var meta caseRequest
			readJSON(t, filepath.Join(base, "request.json"), &meta)
			modelResponse, err := os.ReadFile(filepath.Join(base, "model.json"))
			if err != nil {
				t.Fatal(err)
			}

			requested := correction.Request{
				Text:            string(text),
				Language:        meta.Language,
				Audience:        meta.Audience,
				PreserveVoice:   meta.PreserveVoice,
				Level:           correction.Level(meta.Level),
				ProtectedTokens: meta.ProtectedTokens,
			}

			var sent correction.Request
			corrector, err := agent.New(context.Background(), agent.Config{
				Generate: func(_ context.Context, req correction.Request) (string, error) {
					sent = req
					return string(modelResponse), nil
				},
			})
			if err != nil {
				t.Fatal(err)
			}

			result, err := corrector.Correct(context.Background(), requested)
			if !sameValue(t, sent, requested) {
				t.Fatalf("the agent sent %+v, want the request verbatim %+v", sent, requested)
			}

			if path := filepath.Join(base, "want-error.json"); fileExists(path) {
				assertBlocked(t, err, path)
				if result.RevisedText != "" || len(result.Changes) != 0 || len(result.Questions) != 0 {
					t.Fatalf("Correct() returned %+v, want no draft for a blocked case", result)
				}
				return
			}

			if err != nil {
				t.Fatalf("Correct() = %v, want the verified draft", err)
			}

			wantBytes, err := os.ReadFile(filepath.Join(base, "want.json"))
			if err != nil {
				t.Fatal(err)
			}
			want, err := agent.DecodeResult(wantBytes)
			if err != nil {
				t.Fatalf("want.json does not match the contract: %v", err)
			}
			if !sameValue(t, result, want) {
				t.Fatalf("Correct() = %+v, want %+v", result, want)
			}
		})
	}
}

func assertBlocked(t *testing.T, err error, path string) {
	t.Helper()

	var want wantError
	readJSON(t, path, &want)

	var blocked *correction.BlockedError
	if !errors.As(err, &blocked) {
		t.Fatalf("Correct() error = %v (%T), want *correction.BlockedError", err, err)
	}
	if want.Code != string(correction.CodeUnverifiedDraft) {
		t.Fatalf("want-error code = %q, want %s", want.Code, correction.CodeUnverifiedDraft)
	}

	codes := make([]string, 0, len(blocked.Findings))
	for _, finding := range blocked.Findings {
		codes = append(codes, string(finding.Code))
	}
	if !slices.Equal(codes, want.Findings) {
		t.Fatalf("blocked findings = %v, want %v", codes, want.Findings)
	}
}

func readJSON(t *testing.T, path string, target any) {
	t.Helper()

	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if err := json.Unmarshal(raw, target); err != nil {
		t.Fatalf("decode %s: %v", path, err)
	}
}

func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

// sameValue compares two values by their JSON encoding, so slices compare by
// content instead of by identity.
func sameValue(t *testing.T, got, want any) bool {
	t.Helper()

	gotJSON, err := json.Marshal(got)
	if err != nil {
		t.Fatalf("encode the produced value: %v", err)
	}
	wantJSON, err := json.Marshal(want)
	if err != nil {
		t.Fatalf("encode the expected value: %v", err)
	}
	return string(gotJSON) == string(wantJSON)
}
