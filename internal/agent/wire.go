// Package agent runs the correction agent behind one seam: a generate
// function. The default implementation talks to Google ADK Go v2 and Gemini;
// tests inject a fake, so no test needs a network or a credential.
package agent

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"slices"
	"sort"
	"strings"

	"github.com/GhandyP/text-correction-improvement/internal/correction"
)

// The exact keys a model response must carry. The decoder and the response
// schema share them, so the prompt contract and the parser cannot drift apart.
var (
	resultKeys = []string{"corrected", "changes", "questions"}
	changeKeys = []string{"kind", "original", "revised", "reason"}
)

// DecodeResult decodes one model response into the correction contract. The
// shape is strict on purpose: the payload must be a single JSON object, the
// keys must be spelled exactly as the schema declares them, every declared key
// is required, and unknown keys are rejected. A drifted response is an error,
// never a silently mangled draft.
//
// Decoding does not verify the draft: Verify owns the semantic rules, and the
// two boundaries stay separate so each is testable on its own.
func DecodeResult(raw []byte) (correction.Result, error) {
	var result correction.Result

	obj, err := decodeObject(raw, "")
	if err != nil {
		return correction.Result{}, err
	}
	if err := requireExactKeys(obj, "", resultKeys); err != nil {
		return correction.Result{}, err
	}

	if err := decodeField(obj, "", "corrected", &result.RevisedText); err != nil {
		return correction.Result{}, err
	}

	var changes []json.RawMessage
	if err := decodeField(obj, "", "changes", &changes); err != nil {
		return correction.Result{}, err
	}
	for index, rawChange := range changes {
		change, err := decodeChange(rawChange, index)
		if err != nil {
			return correction.Result{}, err
		}
		result.Changes = append(result.Changes, change)
	}

	if err := decodeField(obj, "", "questions", &result.Questions); err != nil {
		return correction.Result{}, err
	}

	return result, nil
}

func decodeChange(raw json.RawMessage, index int) (correction.Change, error) {
	path := fmt.Sprintf("changes[%d]", index)

	obj, err := decodeObject(raw, path)
	if err != nil {
		return correction.Change{}, err
	}
	if err := requireExactKeys(obj, path, changeKeys); err != nil {
		return correction.Change{}, err
	}

	var change correction.Change
	if err := decodeField(obj, path, "kind", &change.Kind); err != nil {
		return correction.Change{}, err
	}
	if err := decodeField(obj, path, "original", &change.Original); err != nil {
		return correction.Change{}, err
	}
	if err := decodeField(obj, path, "revised", &change.Revised); err != nil {
		return correction.Change{}, err
	}
	if err := decodeField(obj, path, "reason", &change.Reason); err != nil {
		return correction.Change{}, err
	}
	return change, nil
}

// decodeObject decodes exactly one JSON object and rejects trailing content.
func decodeObject(raw []byte, path string) (map[string]json.RawMessage, error) {
	trimmed := bytes.TrimSpace(raw)
	if len(trimmed) == 0 {
		return nil, invalidJSON(path, "the payload is empty")
	}

	decoder := json.NewDecoder(bytes.NewReader(trimmed))
	var obj map[string]json.RawMessage
	if err := decoder.Decode(&obj); err != nil {
		return nil, invalidJSON(path, err.Error())
	}
	if _, err := decoder.Token(); err != io.EOF {
		return nil, invalidJSON(path, "content follows the JSON object, so the payload is not one object")
	}
	if obj == nil {
		return nil, invalidJSON(path, "the payload is not a JSON object")
	}
	return obj, nil
}

// requireExactKeys enforces the declared key set: every required key is present
// and no other key is. Missing keys are reported before unknown ones, and the
// unknown key reported is the first in sorted order, so the error is stable.
func requireExactKeys(obj map[string]json.RawMessage, path string, required []string) error {
	for _, key := range required {
		if _, present := obj[key]; !present {
			return missingField(path, key, caseNeighbour(obj, key))
		}
	}

	var unknown []string
	for key := range obj {
		if !slices.Contains(required, key) {
			unknown = append(unknown, key)
		}
	}
	if len(unknown) > 0 {
		sort.Strings(unknown)
		return &correction.Error{
			Code:   correction.CodeUnknownField,
			Detail: fmt.Sprintf("%s carries the unknown field %q; the allowed fields are %s", where(path), unknown[0], strings.Join(required, ", ")),
		}
	}
	return nil
}

// caseNeighbour returns the sorted first key of obj that differs from key only
// in case, so a case-drifted response names both the missing field and the one
// the model actually sent.
func caseNeighbour(obj map[string]json.RawMessage, key string) string {
	var matches []string
	for candidate := range obj {
		if candidate != key && strings.EqualFold(candidate, key) {
			matches = append(matches, candidate)
		}
	}
	if len(matches) == 0 {
		return ""
	}
	sort.Strings(matches)
	return matches[0]
}

// decodeField decodes one declared key into target. An explicit null is treated
// as a missing field: this contract has no nullable members.
func decodeField(obj map[string]json.RawMessage, path, key string, target any) error {
	raw := bytes.TrimSpace(obj[key])
	if bytes.Equal(raw, []byte("null")) {
		return missingField(path, key, "")
	}
	if err := json.Unmarshal(raw, target); err != nil {
		return invalidJSON(fieldPath(path, key), err.Error())
	}
	return nil
}

func missingField(path, key, neighbour string) error {
	detail := fmt.Sprintf("%s is missing the required field %q", where(path), key)
	if neighbour != "" {
		detail += fmt.Sprintf("; it carries %q instead, and the contract is case-sensitive", neighbour)
	}
	return &correction.Error{Code: correction.CodeMissingField, Detail: detail}
}

func invalidJSON(path, detail string) error {
	return &correction.Error{
		Code:   correction.CodeInvalidJSON,
		Detail: fmt.Sprintf("%s is not valid JSON: %s", where(path), detail),
	}
}

func fieldPath(path, key string) string {
	if path == "" {
		return key
	}
	return path + "." + key
}

func where(path string) string {
	if path == "" {
		return "the model response"
	}
	return path
}
