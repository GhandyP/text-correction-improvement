package agent

import (
	"fmt"
	"strings"

	"github.com/GhandyP/text-correction-improvement/internal/correction"
)

// instruction is the agent's system instruction. The model edits, it never
// rewrites, and every rule here is one the Go side verifies afterwards.
const instruction = `You correct and improve one supplied text. You edit; you never rewrite.

Hard rules:
- Preserve the meaning, the facts, and the author's voice; never add or remove information.
- Return the whole corrected text in "corrected", even when nothing needed changing.
- Never alter, translate, translate away, or drop a protected token; reproduce each one exactly.
- List every change in "changes" with its kind, the exact original fragment, the exact revised fragment, and a reason.
- Replacements only: both fragments must appear literally, the original in the submitted text and the revised in the corrected text.
- If something is ambiguous, do not guess: put the question in "questions" instead.
- Reply with one JSON object that matches the response schema and nothing else: no prose, no markdown fence.`

// Instruction returns the instruction the agent runs under.
func Instruction() string { return instruction }

// Prompt renders one correction request as the user message: the text to
// correct plus every constraint the contract later verifies.
func Prompt(req correction.Request) string {
	var b strings.Builder

	b.WriteString("Correct the text between the markers and return only the JSON object.\n\n")
	fmt.Fprintf(&b, "Level: %s\n", req.Level)

	if language := strings.TrimSpace(req.Language); language != "" {
		fmt.Fprintf(&b, "Language: %s\n", language)
	} else {
		b.WriteString("Language: not specified; keep the text in its own language.\n")
	}

	if audience := strings.TrimSpace(req.Audience); audience != "" {
		fmt.Fprintf(&b, "Audience: %s\n", audience)
	} else {
		b.WriteString("Audience: not specified.\n")
	}

	if req.PreserveVoice {
		b.WriteString("Preserve the author's voice and register; the result must still sound like its author.\n")
	} else {
		b.WriteString("You may adjust the register, but keep the meaning and the facts intact.\n")
	}

	if len(req.ProtectedTokens) == 0 {
		b.WriteString("No protected tokens.\n")
	} else {
		b.WriteString("Protected tokens, each reproduced exactly:\n")
		for _, token := range req.ProtectedTokens {
			fmt.Fprintf(&b, "- %q\n", token)
		}
	}

	b.WriteString("\n--- BEGIN TEXT ---\n")
	b.WriteString(req.Text)
	b.WriteString("\n--- END TEXT ---\n")

	return b.String()
}
