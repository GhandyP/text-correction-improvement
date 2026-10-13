# 09 — Correct and improve text

## Goal

Correct supplied text while preserving its meaning, voice, intended audience, and requested level of change.

## Run the current prototype

From this case directory:

```bash
export GOOGLE_API_KEY="your-key"
go run .
```

`internal/adkhelper.Run` uses ADK Go v2.2.0, Gemini, and `gemini-flash-latest`; set `ADK_MODEL` to override the model.

Sample prompt:

```text
Correct this paragraph while preserving its informal voice.
```

## What exists now

- `main.go (exists)` passes user text and instructions to one focused Gemini ADK agent.
- The result is free-form revised text; changed spans, preserved placeholders, language, and correction reasons are not machine-readable.
- This is intentionally text-only: it has no editor integration, glossary, document store, or automatic publish action.

## Missing elements

- **Contract:** Define `CorrectionInput{text, language, audience, style, protectedTokens[]}` -> `CorrectionResult{revisedText, changes[], preservedTokens[], warnings[]}`.
- **ADK layer:** Add a response schema and document session; keep the agent instruction explicit about meaning preservation, with no function tool for the first slice.
- **Domain boundary:** Add glossary/style-rule services and a diff renderer that can reject changes to protected tokens.
- **Fixtures and validation:** Include grammar errors, informal voice, markup, placeholders, multilingual text, and already-correct input; compare golden diffs.
- **Safety:** Keep revisions as drafts and require review before replacing a source document, especially for legal or customer-facing text.

## Target structure

```text
09-text-correction-improvement/
├── main.go (exists)
├── README.md (exists)
├── schema.go (next)
├── corrector.go (next)
├── diff.go (next)
├── testdata/text.json (next)
└── main_test.go (next)
```

## Implementation order

1. Define protected tokens, style options, and the revision contract.
2. Add schema validation and a session scoped to one document.
3. Implement deterministic token protection and diff reporting.
4. Test voice, markup, placeholders, languages, and no-op input.
5. Add glossary lookup only after the draft diff is reviewable.

## Done when

- [ ] Meaning, protected tokens, markup, and requested voice survive correction tests.
- [ ] Every change has a reason, and malformed output is rejected.
- [ ] Golden fixtures cover no-op, multilingual, and placeholder cases.
- [ ] Source replacement never happens without a human-approved diff.

## Next improvement

Introduce `CorrectionResult` and protected-token tests before adding style or editor integrations.
