# text-correction-improvement

Correct supplied text while preserving its meaning, voice, intended audience, and requested level of
change — and show the author what changed before anyone uses it.

One Google ADK (Go v2) agent does the editing. **Go, not the model, decides whether the answer is
acceptable**: every change must be locatable in the source and in the draft, every protected token
must survive verbatim, and a draft that fails those checks is blocked instead of printed.

## Requirements

- Go 1.26.6 or newer
- A Gemini API key in `GOOGLE_API_KEY` for a real run. The test suite needs no credential and no
  network.

## Quick start

```bash
export GOOGLE_API_KEY="your-key"
go build -o text-correction-improvement .

# from a file
./text-correction-improvement -file draft.txt

# from stdin, machine-readable
echo "Mira, el nene corrio rapidísimo." | ./text-correction-improvement -json

# protect placeholders: they come out exactly as they went in, or nothing is printed
./text-correction-improvement -protect '%s' -protect '{{nombre}}' -file template.txt
```

The draft is the only thing on stdout, so it pipes. The report — every change with its reason, and
the questions the agent would rather ask than guess — goes to stderr. Output of a real run looks like
this:

```console
$ ./text-correction-improvement -protect '%s' "Mira, el nene corrio muy rapido y el %s quedo intacto."
Mira, el nene corrió muy rápido y el %s quedó intacto.
changes (3):
  1. [spelling] "corrio" -> "corrió" - missing accent
  2. [spelling] "rapido" -> "rápido" - missing accent
  3. [spelling] "quedo" -> "quedó" - missing accent
```

## Flags

| Flag | Meaning |
| --- | --- |
| `-file path` | Read the text from `path`; `-` means standard input |
| `-json` | Print the contract as JSON instead of draft plus text report |
| `-language tag` | Language hint, for example `es-AR` |
| `-audience text` | Audience hint, for example `clients` |
| `-level minimal\|moderate` | How far to go; `minimal` fixes errors and keeps the phrasing, `moderate` also smooths unclear phrasing (default `minimal`) |
| `-preserve-voice` | Keep the author's voice and register (default `true`) |
| `-protect token` | A token that must survive verbatim; repeat the flag for each token |
| `-model name` | ADK model name; empty means `ADK_MODEL`, then `gemini-flash-latest` |
| `-h`, `-help` | Print the usage message |

Text comes from the single argument, from `-file`, or from standard input — exactly one of them.
Flags come before the text, following the Go `flag` package's convention.

## Exit codes

| Code | Meaning |
| --- | --- |
| `0` | A verified draft was delivered |
| `1` | The run failed: model error, undecodable response, or a blocked draft |
| `2` | The invocation was wrong: bad flags, bad input, or an invalid request |

## JSON output

Success, on stdout — the same shape the model is asked for:

```json
{
  "corrected": "Mira, el nene corrió muy rápido.",
  "changes": [
    {"kind": "spelling", "original": "corrio", "revised": "corrió", "reason": "missing accent"}
  ],
  "questions": []
}
```

Failure, on stderr, so stdout never carries something that is not a verified draft:

```json
{
  "code": "unverified_draft",
  "detail": "unverified_draft: 1 finding(s): protected_token_lost",
  "findings": [
    {"code": "protected_token_lost", "changeIndex": -1, "detail": "protected token \"%s\" does not appear in the revised draft"}
  ]
}
```

## What Go verifies

A draft is delivered only when all of these hold. Anything else is a block, not a warning:

- Every change names a kind from the closed set `spelling|grammar|punctuation|style|clarity`, and
  states a reason.
- Both sides of a change appear literally — the original in the submitted text, the revised in the
  draft — compared with whitespace folded and case kept. A change nobody can locate is not a change.
- Every protected token is present in the draft. A token missing from the submitted text is rejected
  before the model is even called, because it protects nothing.
- The change list explains the draft: a draft that differs with no changes listed is blocked, and so
  is a change list whose draft is identical to the source.
- The model's answer is one JSON object with the exact keys, all of them present; unknown keys,
  key-case drift, trailing content, and markdown fences are rejected rather than repaired.

## Configuration

| Variable | Meaning |
| --- | --- |
| `GOOGLE_API_KEY` | Gemini credential, read from the environment only and never written to disk |
| `ADK_MODEL` | Model override, for example `gemini-2.5-pro`; the default is `gemini-flash-latest` |

## Development

```bash
make check                 # gofmt check, vet, build, test
go test ./... -count=1
```

The suite needs no network and no credential: the agent is built with an injected generator, so the
fixtures in `internal/agent/testdata/cases` replay recorded model responses through the real pipeline
— validate, generate, decode, verify. **A live Gemini call is not covered by CI**; that step is manual
and requires a key.

- Task plan, specs, and commit-by-commit evidence: [`odd/tasks/text-correction-v1.md`](odd/tasks/text-correction-v1.md)
- Fixture corpus: [`internal/agent/testdata/cases`](internal/agent/testdata/cases)

## Limitations

- Only the model call itself is untested automatically, by design.
- Replacements only: a pure deletion has to be anchored with surrounding context, so both sides of a
  change stay locatable.
- No glossary, no style rules, no editor integration, no document store: v1 corrects one text per run.

## Scope

The tool never writes, replaces, or publishes a document. The draft and its findings go to stdout and
stderr, and a human decides what to do with them. A blocked draft is discarded rather than shown.

## License

MIT — see [LICENSE](LICENSE).
