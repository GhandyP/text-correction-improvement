// Package correction defines the text correction contract: what the caller
// submits, what the agent returns, and the deterministic rules that decide
// whether a draft may leave the process.
//
// The package is pure: it performs no I/O, calls no model, and holds no state,
// so every rule is testable without a network, a credential, or a live agent.
package correction

import "slices"

// ChangeKind is the closed set of change categories a draft may report. A kind
// outside the set is a validation failure, never a rendered change.
type ChangeKind string

// The closed set of change kinds.
const (
	KindSpelling    ChangeKind = "spelling"
	KindGrammar     ChangeKind = "grammar"
	KindPunctuation ChangeKind = "punctuation"
	KindStyle       ChangeKind = "style"
	KindClarity     ChangeKind = "clarity"
)

var changeKinds = []ChangeKind{KindSpelling, KindGrammar, KindPunctuation, KindStyle, KindClarity}

// ChangeKinds returns a fresh copy of the closed change-kind set in its
// canonical order, so no caller can mutate it.
func ChangeKinds() []ChangeKind { return slices.Clone(changeKinds) }

// Valid reports whether k belongs to the closed set.
func (k ChangeKind) Valid() bool { return slices.Contains(changeKinds, k) }

// Level is the closed set of requested change intensities. It is required: an
// empty level is a validation failure, not a default.
type Level string

const (
	// LevelMinimal fixes errors and keeps the author's own phrasing.
	LevelMinimal Level = "minimal"
	// LevelModerate also smooths unclear phrasing, still preserving meaning and voice.
	LevelModerate Level = "moderate"
)

var levels = []Level{LevelMinimal, LevelModerate}

// Levels returns a fresh copy of the closed level set in its canonical order.
func Levels() []Level { return slices.Clone(levels) }

// Valid reports whether l belongs to the closed set.
func (l Level) Valid() bool { return slices.Contains(levels, l) }

// Change is one replacement the agent claims it made, with its reason.
//
// Both sides are replacements: Original must be a literal substring of the
// submitted text and Revised a literal substring of the revised draft. A pure
// deletion or insertion is out of contract; the agent anchors it with
// surrounding context so Verify can locate it.
type Change struct {
	Kind     ChangeKind
	Original string
	Revised  string
	Reason   string
}

// Request is one correction job for one document.
type Request struct {
	Text            string
	Language        string
	Audience        string
	PreserveVoice   bool
	Level           Level
	ProtectedTokens []string
}

// Result is the agent's answer: the revised draft, the changes it claims, and
// the ambiguities it would rather ask than guess.
type Result struct {
	RevisedText string
	Changes     []Change
	Questions   []string
}

// Code identifies a validation failure or a verification finding.
type Code string

// Request validation codes.
const (
	CodeEmptyText               Code = "empty_text"
	CodeUnknownLevel            Code = "unknown_level"
	CodeEmptyProtectedToken     Code = "empty_protected_token"
	CodeDuplicateProtectedToken Code = "duplicate_protected_token"
	CodeProtectedTokenNotInText Code = "protected_token_not_in_text"
)

// Draft verification codes.
const (
	CodeEmptyDraft           Code = "empty_draft"
	CodeProtectedTokenLost   Code = "protected_token_lost"
	CodeMissingChanges       Code = "missing_changes"
	CodeStaleChanges         Code = "stale_changes"
	CodeReplacementSideEmpty Code = "replacement_side_empty"
	CodeUnknownChangeKind    Code = "unknown_change_kind"
	CodeEmptyReason          Code = "empty_reason"
	CodeOriginalNotInText    Code = "original_not_in_text"
	CodeRevisedNotInDraft    Code = "revised_not_in_draft"
	CodeEmptyQuestion        Code = "empty_question"
)

// Error is a typed validation failure. Every error carries a Code for callers
// and a Detail for humans, and the rendered message always names the code.
type Error struct {
	Code   Code
	Detail string
}

func (e *Error) Error() string { return string(e.Code) + ": " + e.Detail }

// Finding is one deterministic verification problem. ChangeIndex is the index
// of the offending change, or -1 for findings about the request or the draft as
// a whole. Every finding blocks the draft.
type Finding struct {
	Code        Code
	ChangeIndex int
	Detail      string
}
