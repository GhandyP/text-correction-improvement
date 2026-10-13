package correction

import (
	"fmt"
	"strings"
)

// ValidateRequest rejects a request that cannot be corrected or verified,
// returning the first failure as a typed *Error. A protected token must be
// present in the text it is supposed to protect; otherwise it protects nothing.
func ValidateRequest(req Request) error {
	if strings.TrimSpace(req.Text) == "" {
		return &Error{Code: CodeEmptyText, Detail: "the text to correct is empty"}
	}
	if !req.Level.Valid() {
		return &Error{
			Code:   CodeUnknownLevel,
			Detail: fmt.Sprintf("level %q is not one of %s", req.Level, levelNames()),
		}
	}

	folded := foldText(req.Text)
	seen := make(map[string]struct{}, len(req.ProtectedTokens))
	for _, token := range req.ProtectedTokens {
		trimmed := strings.TrimSpace(token)
		if trimmed == "" {
			return &Error{Code: CodeEmptyProtectedToken, Detail: "a protected token is empty"}
		}

		key := foldText(trimmed)
		if _, duplicate := seen[key]; duplicate {
			return &Error{
				Code:   CodeDuplicateProtectedToken,
				Detail: fmt.Sprintf("protected token %q is repeated", trimmed),
			}
		}
		seen[key] = struct{}{}

		if !strings.Contains(folded, key) {
			return &Error{
				Code:   CodeProtectedTokenNotInText,
				Detail: fmt.Sprintf("protected token %q does not appear in the text, so it protects nothing", trimmed),
			}
		}
	}
	return nil
}

// foldText collapses every run of whitespace into one space and trims the ends,
// so the rules tolerate a reflowed line but not reworded, re-cased, or
// re-accented text.
func foldText(text string) string { return strings.Join(strings.Fields(text), " ") }

func levelNames() string {
	names := make([]string, 0, len(levels))
	for _, level := range levels {
		names = append(names, string(level))
	}
	return strings.Join(names, ", ")
}
