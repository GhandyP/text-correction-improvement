# Changelog

All notable changes to this project are documented here. The format follows
[Keep a Changelog](https://keepachangelog.com/en/1.1.0/), and the project uses
[Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added

- Correction contract in `internal/correction`: `Request{text, language, audience, preserveVoice,
  level, protectedTokens}` to `Result{revisedText, changes, questions}`, a closed change-kind set
  (`spelling|grammar|punctuation|style|clarity`), typed validation errors, and deterministic `Verify`
  findings for provenance, protected tokens, unexplained drafts, empty drafts, and empty reasons.
- Strict model-response decoder in `internal/agent`: exact key set, every key required, `null` treated
  as missing, unknown keys and markdown fences rejected, and a diagnostic that names both spellings
  when a key drifts only in case.
- Google ADK Go v2 Gemini agent (`google.golang.org/adk/v2`), one session per document, no tools,
  response schema and JSON MIME type, `gemini-flash-latest` by default with `ADK_MODEL` override, and
  an injected generate seam so the whole pipeline is tested without a network or a credential.
- CLI in `internal/cli`: `-file`, `-json`, `-language`, `-audience`, `-level`, `-preserve-voice`,
  `-protect`, `-model`; exit codes `0` verified draft, `1` failed or blocked run, `2` usage error;
  draft on stdout and report on stderr.
- Fixture corpus with six recorded-response cases (`already-correct`, `informal-voice`,
  `placeholders-markup`, `multilingual`, `ambiguous-tone`, `protected-token-lost`) replayed through
  the real pipeline.
- Makefile with `check`, and a GitHub Actions workflow running gofmt, vet, build, and test.
- MIT license, README, and the ODD task plan that documents specs, decisions, and evidence.

### Security

- `GOOGLE_API_KEY` is read from the environment only, is never written to disk, and `.gitignore`
  excludes `.env` and token files.
- A draft that fails verification is discarded: an unverified revision is never printed as if it had
  been reviewed.
