package agent_test

import (
	"slices"
	"strings"
	"testing"

	"google.golang.org/genai"

	"github.com/GhandyP/text-correction-improvement/internal/agent"
	"github.com/GhandyP/text-correction-improvement/internal/correction"
)

func TestInstructionStatesTheHardRules(t *testing.T) {
	instruction := agent.Instruction()

	for _, rule := range []string{
		"meaning",
		"voice",
		"protected token",
		"questions",
		"JSON",
	} {
		if !strings.Contains(instruction, rule) {
			t.Errorf("Instruction() does not mention %q; the rules must be explicit:\n%s", rule, instruction)
		}
	}
}

func TestPromptCarriesTheWholeRequest(t *testing.T) {
	req := correctionRequest()
	req.Language = "es-AR"
	req.Audience = "small business owners"
	req.Level = correction.LevelModerate
	req.ProtectedTokens = []string{"%s", "ACME"}

	prompt := agent.Prompt(req)

	for _, want := range []string{req.Text, "es-AR", "small business owners", "moderate", "%s", "ACME", "voice and register"} {
		if !strings.Contains(prompt, want) {
			t.Errorf("Prompt() does not carry %q:\n%s", want, prompt)
		}
	}
}

func TestPromptMarksAnEmptyProtectedTokenList(t *testing.T) {
	req := correctionRequest()
	req.ProtectedTokens = nil

	if prompt := agent.Prompt(req); !strings.Contains(prompt, "No protected tokens") {
		t.Errorf("Prompt() does not say the request protects no tokens:\n%s", prompt)
	}
}

func TestPromptDropsTheVoiceClauseWhenVoiceIsNotProtected(t *testing.T) {
	req := correctionRequest()
	req.PreserveVoice = false

	prompt := agent.Prompt(req)
	if strings.Contains(prompt, "voice and register") {
		t.Errorf("Prompt() still demands the voice when PreserveVoice is false:\n%s", prompt)
	}
	if !strings.Contains(prompt, "register") {
		t.Errorf("Prompt() does not mention the register at all:\n%s", prompt)
	}
}

func TestResultSchemaIsStrict(t *testing.T) {
	schema := agent.ResultSchema()

	if schema.Type != genai.TypeObject {
		t.Errorf("schema type = %q, want OBJECT", schema.Type)
	}

	wantKeys := []string{"corrected", "changes", "questions"}
	if got := sortedKeys(schema.Properties); !slices.Equal(got, sorted(wantKeys)) {
		t.Errorf("schema properties = %v, want the set %v", got, wantKeys)
	}
	if got := slices.Clone(schema.Required); !slices.Equal(got, wantKeys) {
		t.Errorf("schema required = %v, want %v in contract order", got, wantKeys)
	}

	changeSchema := schema.Properties["changes"].Items
	if changeSchema == nil {
		t.Fatal("changes has no item schema")
	}
	wantChangeKeys := []string{"kind", "original", "revised", "reason"}
	if got := sortedKeys(changeSchema.Properties); !slices.Equal(got, sorted(wantChangeKeys)) {
		t.Errorf("change properties = %v, want the set %v", got, wantChangeKeys)
	}
	if got := slices.Clone(changeSchema.Required); !slices.Equal(got, wantChangeKeys) {
		t.Errorf("change required = %v, want %v in contract order", got, wantChangeKeys)
	}

	wantKinds := make([]string, 0, len(correction.ChangeKinds()))
	for _, kind := range correction.ChangeKinds() {
		wantKinds = append(wantKinds, string(kind))
	}
	if got := changeSchema.Properties["kind"].Enum; !slices.Equal(got, wantKinds) {
		t.Errorf("change kind enum = %v, want the closed set %v", got, wantKinds)
	}
}

func sortedKeys(properties map[string]*genai.Schema) []string {
	keys := make([]string, 0, len(properties))
	for key := range properties {
		keys = append(keys, key)
	}
	slices.Sort(keys)
	return keys
}

func sorted(values []string) []string {
	out := slices.Clone(values)
	slices.Sort(out)
	return out
}
