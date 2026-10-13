# text-correction-v1 — standalone ADK text correction and improvement

Public Go project: one Google ADK (Go v2) agent corrects a supplied text and returns a
machine-readable result — the revised draft, a reason for every change, deterministic protected-token
verification, and the ambiguities it would rather ask than guess. The tool writes nothing on disk;
the draft goes to stdout so a human reviews the diff before any source document is replaced.

This is a **rewrite from scratch**, not a port. The inherited prototype (cookbook `main.go` importing
`internal/adkhelper`, no `go.mod`) is input, not spec; S5 authorizes the reset.

Repo: `github.com/GhandyP/text-correction-improvement` (public, MIT). Work lands on `main`, one
work-unit commit per task.

## Specs

S1 — user (verbatim): "recuerda en usar google adk y este sea un repo total mente independiente".
Own Git repository initialized in this directory (never the `$HOME` repository), own Go module
`github.com/GhandyP/text-correction-improvement`, zero imports from
`github.com/GhandyP/adk-go-enterprise-cookbook`, MIT license.

S2 — user (verbatim): "usar google adk". The agent layer is `google.golang.org/adk/v2` with the
Gemini model package; no other agent framework enters the tree.

S3 — user (verbatim): "ademas vamos a crear un repo publico en github con el cli git y gh". The
public repository is created with `gh repo create --public`, wired as `origin`, and pushed. No pull
request is opened.

S4 — user (verbatim): "usar odd para la creacion de este repo de ejemplo para mi github". This
document is the spec for every task; each task closes with a work-unit commit on `main` and the
commit identity is recorded as evidence.

S5 — user (verbatim): "vamos a volver a hacerlo desde el principio para que sea su propio repo
intependiendte". Full reset: the inherited cookbook README and `main.go` carry no authority over the
new contract.

S6 — user decision (selected): the repository is named `text-correction-improvement`.

S7 — user decision (selected): v1 scope is "Contrato + agente ADK estructurado": the correction
contract with typed validation, a response-schema-constrained agent, a real CLI with flags and JSON
output, fixtures with golden model responses, and honest docs plus CI. The diff renderer, the
glossary/style-rule service, and any editor or document-store integration are later slices.

S8 — project goal, quoted verbatim from the inherited case README: "Correct supplied text while
preserving its meaning, voice, intended audience, and requested level of change."

S9 — hard rule, quoted verbatim from the inherited case README: "Source replacement never happens
without a human-approved diff." v1 writes no file: the draft and its findings go to stdout only.

S10 — case guide, quoted verbatim from `guia-80-casos.md:350-359`: elements
"`Text{body, audience, preserve_voice}`; output `{corrected, changes[], questions}`"; fast test
"Provide a paragraph with spelling, agreement, and a deliberate informal phrase; ensure the voice
survives."

S11 — design decisions, authorized by S5 and S7:
- Closed change-kind set (`spelling|grammar|punctuation|style|clarity`); an unknown kind is a
  validation failure, never rendered.
- Provenance: every `Change.Original` must be a literal substring of the submitted text and every
  `Change.Revised` a literal substring of the revised draft (whitespace-folded, case-strict), and
  every reason must be non-empty. A change nobody can locate is not a change.
- Protected tokens are verified in Go, deterministically; the model never self-certifies. A lost
  protected token blocks the draft with a typed error and a non-zero exit.
- Findings are computed by Go (`Verify`), not reported by the model; the model may only report
  `questions` for ambiguities.
- The ADK `console/web` launcher is dropped: this case is a text pipeline, so `main.go` is a real
  CLI. ADK supplies the agent and model layer, not the command surface.
- Model default `gemini-flash-latest`, overridable with `ADK_MODEL`; API key read from
  `GOOGLE_API_KEY` only and never written to disk.
- The guide's recommended models (Qwen/GLM) are not used: S2 fixes Google ADK and its Gemini
  integration for this repository.

S12 — non-goals for v1: no writing, replacing, or publishing documents; no glossary or style-rule
service; no editor integration; no document store; no multi-document batch; no streaming UI; no
ADK launcher subcommands. Adjacent cookbook cases stay out of scope: `10-mass-translation` (batch
translation), `03-very-long-document-summary` (summarization), `19-message-urgency-detection`.

## Tasks

Commit identities are recorded as evidence once each task lands; they are never predicted.

| ID | Task | Specs | Route | Commit |
| --- | --- | --- | --- | --- |
| T0 | ODD tracking: independent `git init` here, inherited-state commit, this document | S1, S4, S5 | inline | `f54f1d5` |
| T1 | Reset: drop the inherited prototype, MIT LICENSE, `.gitignore`, own `go.mod`, `main.go` that compiles without `internal/adkhelper`, honest README | S1, S5, S6 | inline | `cbdd0a3` |
| T2 | `internal/correction`: contract types, closed change-kind set, typed validation, deterministic `Verify` (provenance, protected tokens, empty draft) with RED tests | S8, S10, S11 | inline | `4f32cf1` |
| T3 | Response schema plus strict JSON decode (unknown and missing fields rejected) with RED tests | S7, S11 | inline | `7f0bdd7` |
| T4 | `internal/agent`: ADK Go v2 Gemini agent, one document per session, no tools, injectable generate seam, instruction contract tests, `adk/v2` pinned | S2, S7, S11 | inline | `681bce8` |
| T5 | `internal/cli`: flags (`-file`, `-json`, `-language`, `-audience`, `-level`, `-protect`, `-model`), stdin/file/argument input, text and JSON output, exit codes, RED tests | S7, S8, S10 | inline | `444fecc` |
| T6 | Fixtures and golden model responses: already-correct no-op, informal voice, markup and placeholders, multilingual, lost protected token | S8, S10, S11 | inline | `d0aa640` |
| T7 | Honest README rewrite, Makefile, CI workflow (gofmt, vet, build, test), CHANGELOG | S4, S8, S9 | inline | pending |
| T8 | Publish: `gh repo create --public`, `origin`, push, verify visibility and that the remote tree equals local `HEAD` | S3 | inline | `75360e8` |
| T9 | Close: RDD assess over the candidate, execute its plan, final report with the risk line | S1-S12 | assess | `review-03d8840dc12df998` |
| T10 | Bounded correction of the native finding `R3-001`: protected tokens verified byte for byte | S8, S11 | inline | `f2649d7` |

Every task closes with at least one work-unit commit on `main`. Tests and docs land alongside the
behavior they describe. The push in T8 is authorized by S3; opening pull requests is not.

## Log

L1 — user (verbatim): "mira explora este repo , vamos a volver a hacerlo desde el principio para
que sea su propio repo intependiendte , ademas vamos a crear un repo publico en github con el cli
git y gh , y usar odd para la creacion de este repo de ejemplo para mi github"

L2 — user (verbatim): "recuerda en usar google adk y este sea un repo total mente independiente"

L3 — exploration evidence (incident, 2026-10-16). `~/Documents/09-text-correction-improvement` had no
`.git`; `git rev-parse --show-toplevel` resolved to `/home/ghandy`, where the `$HOME` repository's
index held unrelated staged work (`Desktop/informe mami/Septiembre/INFORME_DE_GESTION_SEPTIEMBRE_2026_V1_BORRADOR.docx`,
`.pdf`, `odd/tasks/inidi-september-2026-report.md`). No `git add` was run there. Fix: `git init -b
main` in the case directory (T0, commit `f54f1d5`); the `$HOME` index was verified unchanged
afterwards. Siblings `06-inbox-prioritization` and `07-task-extraction` already own their
repositories; `08-automated-email-follow-ups` is still inside `$HOME`.

L4 — exploration evidence. The tree held only `README.md` and `main.go` (12K), identical to
`adk-go-enterprise-cookbook/casos/09-text-correction-improvement` (`diff -rq` clean). `main.go`
imported `github.com/GhandyP/adk-go-enterprise-cookbook/internal/adkhelper`, which resolves to the
private cookbook at `/home/ghandy/Documents/P1_proyecticos/adk-go-enterprise-cookbook` (remote
`GhandyP/proyecticos`); `gh repo view GhandyP/adk-go-enterprise-cookbook` does not resolve publicly.
There was no `go.mod`, so the README section "Run the current prototype" was false: nothing built.

L5 — exploration evidence. `internal/adkhelper/runner.go` was read: `gemini.NewModel` with the model
name from `ADK_MODEL` else `gemini-flash-latest`, API key from `GOOGLE_API_KEY`; a single
`llmagent.New` agent with no tools; `launcher.Config` plus `full.NewLauncher()`, which exposes the
ADK console/web launcher rather than a text pipeline. Cookbook `go.mod` pins `adk/v2 v2.2.0` and
`genai v1.66.0`.

L6 — exploration evidence. Sibling conventions: `06-inbox-prioritization` (own repo, module
`github.com/GhandyP/06-inbox-prioritization`, MIT, `.gitignore`, `odd/tasks/`, `internal/priority`
built through RED commits on `main`) is the closest precedent; `04-meeting-transcription-summary`
(`adk-meeting-summary`) pins `adk/v2 v2.6.0` with `go 1.26.6` and builds cleanly on this machine
(probe exit 0), and the go1.26.6 toolchain is present in the module cache. Remotes for the series
are descriptive names: `grounded-email-summary`, `professional-email-drafting`, `adk-meeting-summary`.

L7 — user answers (verbatim): repository name "text-correction-improvement"; scope "Contrato +
agente ADK estructurado (recomendado)".

L8 — decision (T1). `go.mod` declares `go 1.26.6` and T4 pins `google.golang.org/adk/v2 v2.6.0` to
match the newest sibling (04); `GOTOOLCHAIN=auto` was verified to build that module locally, so the
1.26.4 system toolchain is not a blocker. Dependencies arrive with the code that first imports them,
so every commit stays `go mod tidy`-clean: T1 is dependency-free, `genai` and `adk/v2` land at T4.

L9 — decision (T1). Work lands directly on `main`, matching the recorded preference for this series
(`06/odd/tasks/inbox-prioritization-v1.md` L9, `04/odd/tasks/adk-meeting-summary-v1.md` L3). The
push is authorized here by S3 because the user explicitly asked for the public repository; pull
requests were not requested.

L10 — decision (T2/T3). The model output contract is the wire shape of S10 (`corrected`, `changes[]`,
`questions[]`), decoded strictly: unknown fields and missing required fields are errors, so a schema
drift cannot silently pass as a valid draft. `Verify` is a pure function over `(Request, Result)` so
its rules are testable without network, credentials, or a model.

L11 — decision (T4). The agent is constructed behind one seam — the ADK call is a function value
injected into `agent.Agent` — so tests exercise instruction, schema, decode, and verification with a
fake generator and no network. A live Gemini call needs `GOOGLE_API_KEY` and stays out of CI;
`testdata/model/*.json` carries the recorded-shape responses the golden tests replay.

L12 — evidence (T1). Commit `cbdd0a3` on `main`. `go build ./...`, `go vet ./...` and `gofmt -l .`
were clean; `go list -m` reports `github.com/GhandyP/text-correction-improvement`; tracked files are
`.gitignore`, `LICENSE`, `README.md`, `go.mod`, `main.go` and this document. `go run .` prints
"text-correction-improvement: not wired yet; see odd/tasks/text-correction-v1.md" and exits 2, so no
commit claims a working CLI before T5.

L13 — evidence (T2). Commit `4f32cf1`. RED was observed as `no non-test Go files in
.../internal/correction [build failed]`, exit 1. GREEN as
`ok github.com/GhandyP/text-correction-improvement/internal/correction` with 26 test functions;
`go build ./...`, `go vet ./...` and `gofmt -l .` all clean. One fixture was corrected before GREEN
rather than after: the stale-changes test listed accented revisions while handing `Verify` the
unaccented source as the draft, so it was asserting three extra findings of its own making; the
fixture now lists one locatable no-op change, leaving staleness as the only problem.

L14 — decision (T3). The response decoder is deliberately narrower than a tolerant JSON reader:
one object, exact key spellings, every key required, no trailing content, no markdown fence, and
explicit `null` treated as missing. `corrected`, `changes` and `questions` and the four change keys
live in one place shared with the response schema so the two cannot drift; a case-drifted key
reports both the required name and the one received.

L15 — evidence (T3). RED was observed as `no non-test Go files in .../internal/agent [build failed]`,
exit 1. GREEN as `ok .../internal/agent` plus `ok .../internal/correction`, 41 test functions across
both packages; `go build ./...`, `go vet ./...` and `gofmt -l .` all clean. One test was rewritten
before GREEN: the case-drift case expected `unknown_field`, but completeness is checked first by
design, so it now expects `missing_field` and asserts that the detail names both spellings.

L16 — decision (T4). The ADK wiring lives entirely inside one closure built by `newGeminiGenerator`:
the Gemini model, a single `llmagent` with the response schema and MIME type set, an in-memory ADK
runner with session auto-creation, and a fresh session id per request so one document never sees
another's state. `New` fails fast with `missing_api_key` when no credential is configured and no
generator is injected, so a missing key is a construction error rather than a failed first call.
`Correct` runs one direction only — validate, generate, decode, verify — and returns `BlockedError`
instead of a draft when verification finds anything. The ADK `console`/`web` launcher is not used:
the runner is driven directly, and thought parts are excluded when the answer text is collected.

L17 — evidence (T4). RED was observed as a build failure listing the undefined symbols
(`agent.GenerateFunc`, `agent.New`, `agent.Config`, `agent.DefaultModel`, `agent.Agent`,
`correction.CodeMissingAPIKey`), exit 1. GREEN as `ok .../internal/agent` with 31 test functions
passing; `go build ./...`, `go vet ./...` and `gofmt -l .` clean. The first GREEN run surfaced three
real defects, all fixed before the commit: `BlockedError` was defined in the agent package but
belongs to the contract vocabulary, so it moved to `internal/correction`; the blocked-draft fixture
legitimately produces two findings (lost protected token and unexplained draft) and the test now
asserts both in stable order; and the schema test compared an ordered key list against a sorted one.
`go.mod` now pins `google.golang.org/adk/v2 v2.6.0` and `google.golang.org/genai v1.74.0` as direct
dependencies. Local builds run with `GOPROXY=off` because this machine cannot reach
`proxy.golang.org` reliably and the module cache is complete; CI uses the normal proxy.

L18 — decision (T5). The CLI never imports the agent: it depends on a two-method seam (`Corrector`,
`BuildFunc`), so its tests run against a fake with no ADK in the test binary, and `main.go` is the
only place that wires ADK to the command. Exit codes are contract: 0 verified draft, 1 failed run or
blocked draft, 2 usage error or invalid request. In text mode stdout carries only the deliverable and
the report goes to stderr; in JSON mode the success payload is the same contract shape the model
returns, and failures print one error object on stderr, so stdout never carries something that is not
a verified draft.

L19 — evidence (T5). Commit `444fecc`. RED was observed as `no non-test Go files in
.../internal/cli`. GREEN as `ok .../internal/cli` with 17 test functions, and the full suite reports
74 test functions across the three packages with `go build ./...`, `go vet ./...` and `gofmt -l .`
clean. Smoke tests on the built binary: no credential → `missing_api_key` and exit 1; `-h` → usage on
stdout and exit 0; `-level bogus` → `usage error: unknown level "bogus"; use minimal or moderate`
and exit 2.

L20 — evidence (T6). Commit `d0aa640`. Six cases live in `internal/agent/testdata/cases`: already-correct,
informal-voice (the guide's fast test: spelling, agreement, and a deliberate informal phrase),
placeholders-markup, multilingual, ambiguous-tone, and protected-token-lost as the one blocked case.
Each replay builds the request from `input.txt` and `request.json`, feeds `model.json` through the
real pipeline with an injected generator, and compares the verified draft against `want.json`,
which the strict decoder re-reads so no fixture can describe a draft the pipeline could not have
produced. The first run failed on placeholders-markup because `printf` had eaten the `%s` while the
fixture was being written; the pipeline rejected the request with `protected_token_not_in_text`,
which is the test doing its job on its own fixture. Fixed, then the whole suite is green.

L21 — decision (T7). The README states what works and what does not: the automated suite covers
everything except the live model call, which needs a key and stays a manual step. The
flags-after-the-text limitation is documented instead of hidden, because the Go `flag` package stops
at the first positional argument. CI mirrors `make check` — gofmt, vet, build, test — so a local pass
is the same gate the repository runs.

L22 — evidence (T7). `make check` exited 0 with gofmt clean, vet clean, build clean and the suite
green; `.github/workflows/ci.yml` parses as YAML; the Makefile has 12 real tab-indented recipe lines.
The first README draft showed a flag-after-text invocation that the Go `flag` package would reject as
extra positional arguments; both the example and a new sentence now say flags come before the text.

L23 — evidence (T8). The public repository was created with
`gh repo create text-correction-improvement --public --source=. --remote=origin --push
--description "Correct and improve text with a Google ADK Go (v2) Gemini agent, and verify every
change deterministically in Go before printing the draft."`. `gh repo view` reports `visibility:
PUBLIC`, default branch `main`, and the URL `https://github.com/GhandyP/text-correction-improvement`;
`git remote -v` shows `origin` on that URL; `git ls-remote origin refs/heads/main` returns the local
`HEAD` at the moment of the push, `75360e8`; the branch tracks `origin/main` and the tree is clean.
No pull request was opened, because S3 asked for the repository, not for a review flow.

L24 — evidence (T9, native review). RDD assess over the committed range `f54f1d5..91cd5e2` reported
high risk — native signal `shell_source` in `.github/workflows/ci.yml` — and its plan required writer
self-verification plus a separate independent verifier. The `gentle-ai-verify` subagent stalled and is
reported unavailable, not retried. The native ordinary review of lineage `review-03d8840dc12df998`
ran four lenses over the same frozen range (risk, resilience, readability, reliability); the refuter
found one candidate-caused severe finding, `R3-001` (reliability, CRITICAL): protected tokens were
checked with folded whitespace, so a token could change from the inside and still pass while the CLI
reported success and delivered a draft that broke the preservation promise.

L25 — evidence (T10). RED was observed as `TestValidateRequestRejectsAProtectedTokenBrokenByWhitespace`
and `TestVerifyRejectsAProtectedTokenChangedInternally` failing against the folded implementation;
GREEN after `ValidateRequest` requires the token to appear verbatim in the text and `Verify` requires
it to appear verbatim in the draft, so only whitespace around a token may differ. Commit `f2649d7`,
`make check` exit 0, pushed with the remote at the same revision. The declared correction plan was 80
diff lines and the realized diff is 68 additions plus 21 deletions, 89 lines: an 11% overrun on the
estimate, recorded because the plan was a plan and the difference is real.

L26 — evidence (T9 close). The targeted validator approved the corrected candidate, the exact
provider-issued acknowledgement was executed, and the authority is burned
(`gentle-ai.review-acknowledged/v1`, consumed revision `sha256:b6533c9d15d6e98be5b545acfbd5d9ff4965dc0c994742e1252995eeccaa648b`).
One advisory, non-blocking finding remains for later work: `R4-001` (resilience, WARNING,
informational, `internal/cli/cli.go:168-169`); the closure states it must never reopen this candidate.
Delivery follows ordinary repository policy; the public repository is at the corrected revision.
