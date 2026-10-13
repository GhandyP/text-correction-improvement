package agent

import (
	"google.golang.org/genai"
	"slices"

	"github.com/GhandyP/text-correction-improvement/internal/correction"
)

// ResultSchema describes the exact JSON object the model must return. It is
// built from the same key lists the decoder enforces, so the schema the model
// receives and the parser that reads its answer cannot drift apart.
func ResultSchema() *genai.Schema {
	return &genai.Schema{
		Type: genai.TypeObject,
		Properties: map[string]*genai.Schema{
			"corrected": {
				Type:        genai.TypeString,
				Description: "The whole corrected text.",
			},
			"changes": {
				Type:        genai.TypeArray,
				Description: "Every change made, one entry per replacement.",
				Items: &genai.Schema{
					Type: genai.TypeObject,
					Properties: map[string]*genai.Schema{
						"kind": {
							Type:        genai.TypeString,
							Enum:        kindEnum(),
							Description: "The category of the change.",
						},
						"original": {
							Type:        genai.TypeString,
							Description: "The exact fragment of the submitted text that was replaced.",
						},
						"revised": {
							Type:        genai.TypeString,
							Description: "The exact fragment that replaced it in the corrected text.",
						},
						"reason": {
							Type:        genai.TypeString,
							Description: "Why the change was made.",
						},
					},
					Required: slices.Clone(changeKeys),
				},
			},
			"questions": {
				Type:        genai.TypeArray,
				Description: "Ambiguities to ask the author about instead of guessing.",
				Items:       &genai.Schema{Type: genai.TypeString},
			},
		},
		Required: slices.Clone(resultKeys),
	}
}

func kindEnum() []string {
	kinds := correction.ChangeKinds()
	names := make([]string, 0, len(kinds))
	for _, kind := range kinds {
		names = append(names, string(kind))
	}
	return names
}
