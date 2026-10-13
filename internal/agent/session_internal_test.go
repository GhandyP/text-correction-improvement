package agent

import (
	"strings"
	"testing"

	"google.golang.org/genai"
)

func TestNewSessionIDIsUnique(t *testing.T) {
	seen := make(map[string]struct{}, 128)
	for i := 0; i < 128; i++ {
		id, err := newSessionID()
		if err != nil {
			t.Fatalf("newSessionID() = %v, want nil", err)
		}
		if !strings.HasPrefix(id, "doc-") || len(id) <= len("doc-") {
			t.Fatalf("newSessionID() = %q, want a doc- prefixed id", id)
		}
		if _, duplicate := seen[id]; duplicate {
			t.Fatalf("newSessionID() returned the duplicate %q; one document must not share session state", id)
		}
		seen[id] = struct{}{}
	}
}

func TestTextOfConcatenatesTextAndSkipsThoughts(t *testing.T) {
	content := &genai.Content{Parts: []*genai.Part{
		{Text: "the answer"},
		{Text: "internal reasoning", Thought: true},
		nil,
		{Text: " continues"},
	}}

	if got := textOf(content); got != "the answer continues" {
		t.Fatalf("textOf() = %q, want the text parts only", got)
	}
	if got := textOf(nil); got != "" {
		t.Fatalf("textOf(nil) = %q, want an empty string", got)
	}
}
