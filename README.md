# text-correction-improvement

Correct supplied text while preserving its meaning, voice, intended audience, and requested level of
change.

This repository is an independent Go project built with **Organic Driven Development**: the task
plan in [`odd/tasks/text-correction-v1.md`](odd/tasks/text-correction-v1.md) is the spec, and each
task lands as one work-unit commit on `main`.

## Status

The command works end to end when a Gemini key is available: `internal/correction` holds the
contract and the deterministic checks, `internal/agent` runs the Google ADK Go v2 Gemini agent, and
`internal/cli` reads the text and prints the draft. Verifying against a live model is a manual step;
the automated suite covers everything except the model call itself.

```bash
export GOOGLE_API_KEY="your-key"
go run . -file draft.txt          # the draft goes to stdout, the report to stderr
echo "Mira, el nene corrio." | go run . -json
```

| Task | Deliverable | State |
| --- | --- | --- |
| T1 | Standalone module (`go.mod`, LICENSE, `.gitignore`) | done |
| T2 | `internal/correction`: contract and deterministic verification | done |
| T3 | Response schema and strict JSON decode | done |
| T4 | `internal/agent`: Google ADK Go v2 Gemini agent | done |
| T5 | `internal/cli`: flags, JSON output, exit codes | done |
| T6 | Fixtures and golden model responses | pending |
| T7 | README, Makefile, CI, CHANGELOG | pending |

## Environment

The default model is `gemini-flash-latest` through Google ADK Go v2; set `ADK_MODEL` to override it
and `GOOGLE_API_KEY` to authenticate. The key is read from the environment and never written to
disk. Do not commit `.env` files; `.gitignore` already excludes them.

## Scope

The tool corrects text and reports what changed and why. It never writes, replaces, or publishes a
document: the draft and its findings go to stdout so a human reviews the diff before using it. See
the non-goals in the task plan.

## License

MIT — see [LICENSE](LICENSE).
