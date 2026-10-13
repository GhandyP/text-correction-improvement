package correction

import (
	"fmt"
	"strings"
)

// Verify reports every deterministic problem with a draft. It never trusts the
// agent: it re-checks the closed kind set, the provenance of every replacement,
// the survival of every protected token, and that the change list actually
// explains the draft. Findings are returned in a stable order — draft-level
// problems first, then per-change problems in change order, then empty
// questions — and the result is empty when the draft is clean.
func Verify(req Request, res Result) []Finding {
	var findings []Finding

	draft := foldText(res.RevisedText)
	if draft == "" {
		// An empty draft fails every other rule too; one finding is enough.
		return []Finding{{Code: CodeEmptyDraft, ChangeIndex: -1, Detail: "the revised draft is empty"}}
	}

	for _, token := range req.ProtectedTokens {
		trimmed := strings.TrimSpace(token)
		if trimmed == "" {
			continue
		}
		// Verbatim means verbatim: the token is compared byte for byte against
		// the raw draft, so a change inside the token is a lost token even when
		// it survives whitespace folding.
		if !strings.Contains(res.RevisedText, trimmed) {
			findings = append(findings, Finding{
				Code:        CodeProtectedTokenLost,
				ChangeIndex: -1,
				Detail:      fmt.Sprintf("protected token %q does not appear verbatim in the revised draft", trimmed),
			})
		}
	}

	source := foldText(req.Text)
	switch {
	case len(res.Changes) == 0 && draft != source:
		findings = append(findings, Finding{
			Code:        CodeMissingChanges,
			ChangeIndex: -1,
			Detail:      "the draft differs from the text but no change explains it",
		})
	case len(res.Changes) > 0 && draft == source:
		findings = append(findings, Finding{
			Code:        CodeStaleChanges,
			ChangeIndex: -1,
			Detail:      "changes are listed but the draft is identical to the text",
		})
	}

	for index, change := range res.Changes {
		findings = append(findings, verifyChange(index, change, source, draft)...)
	}

	for index, question := range res.Questions {
		if strings.TrimSpace(question) == "" {
			findings = append(findings, Finding{
				Code:        CodeEmptyQuestion,
				ChangeIndex: -1,
				Detail:      fmt.Sprintf("question %d is empty", index+1),
			})
		}
	}

	return findings
}

func verifyChange(index int, change Change, source, draft string) []Finding {
	var findings []Finding

	original := strings.TrimSpace(change.Original)
	revised := strings.TrimSpace(change.Revised)

	if original == "" || revised == "" {
		findings = append(findings, Finding{
			Code:        CodeReplacementSideEmpty,
			ChangeIndex: index,
			Detail:      "a change needs both an original and a revised side; anchor a deletion with surrounding context",
		})
	}
	if !change.Kind.Valid() {
		findings = append(findings, Finding{
			Code:        CodeUnknownChangeKind,
			ChangeIndex: index,
			Detail:      fmt.Sprintf("change kind %q is not one of %s", change.Kind, kindNames()),
		})
	}
	if strings.TrimSpace(change.Reason) == "" {
		findings = append(findings, Finding{
			Code:        CodeEmptyReason,
			ChangeIndex: index,
			Detail:      "a change must state why it was made",
		})
	}
	if original != "" && !strings.Contains(source, foldText(original)) {
		findings = append(findings, Finding{
			Code:        CodeOriginalNotInText,
			ChangeIndex: index,
			Detail:      fmt.Sprintf("original %q does not appear in the submitted text", original),
		})
	}
	if revised != "" && !strings.Contains(draft, foldText(revised)) {
		findings = append(findings, Finding{
			Code:        CodeRevisedNotInDraft,
			ChangeIndex: index,
			Detail:      fmt.Sprintf("revised %q does not appear in the revised draft", revised),
		})
	}

	return findings
}

func kindNames() string {
	names := make([]string, 0, len(changeKinds))
	for _, kind := range changeKinds {
		names = append(names, string(kind))
	}
	return strings.Join(names, ", ")
}

// BlockedError reports a draft that decoded but failed deterministic
// verification. The draft travels no further: an unverified revision is never
// handed back as if it had been reviewed.
type BlockedError struct {
	Findings []Finding
}

func (e *BlockedError) Error() string {
	codes := make([]string, 0, len(e.Findings))
	for _, finding := range e.Findings {
		codes = append(codes, string(finding.Code))
	}
	return fmt.Sprintf("%s: %d finding(s): %s", CodeUnverifiedDraft, len(e.Findings), strings.Join(codes, ", "))
}
