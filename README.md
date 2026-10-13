# text-correction-improvement

Correct supplied text while preserving its meaning, voice, intended audience, and requested level of
change.

This repository is an independent Go project built with **Organic Driven Development**: the task
plan in [`odd/tasks/text-correction-v1.md`](odd/tasks/text-correction-v1.md) is the spec, and each
task lands as one work-unit commit on `main`.

## Status

Early development. The module is standalone; the correction contract, the ADK agent, and the CLI
land task by task. Until T5 lands there is no working command, and this README says so instead of
pretending otherwise.

| Task | Deliverable | State |
| --- | --- | --- |
| T1 | Standalone module (`go.mod`, LICENSE, `.gitignore`) | this commit |
| T2 | `internal/correction`: contract and deterministic verification | pending |
| T3 | Response schema and strict JSON decode | pending |
| T4 | `internal/agent`: Google ADK Go v2 Gemini agent | pending |
| T5 | `internal/cli`: flags, JSON output, exit codes | pending |
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
