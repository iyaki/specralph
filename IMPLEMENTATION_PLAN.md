# Implementation Plan (commands/prompt-authoring)

**Status:** Prompt Discovery Complete; Authoring Helpers Complete (4/8 phases)
**Last Updated:** 2026-09-11
**Primary Spec:** [specs/commands/prompts-authoring.md](specs/commands/prompts-authoring.md), [specs/commands/skill.md](specs/commands/skill.md)

---

## Quick Reference

| System/Subsystem            | Spec                                                            | Module/Package                        | Artifacts                                                        | Status        |
|-----------------------------|-----------------------------------------------------------------|---------------------------------------|------------------------------------------------------------------|---------------|
| Help / run / init / version | [specs/commands/help.md](specs/commands/help.md), [run.md](specs/commands/run.md), [init.md](specs/commands/init.md), [version.md](specs/commands/version.md) | `internal/cli/cmd.go`, `run.go`, `init.go`, `version.go` | —                                                                | ✅ Complete   |
| `prompts list` / `prompts show` | [specs/commands/prompts.md](specs/commands/prompts.md)      | `internal/cli/prompts.go`             | —                                                                | ✅ Complete   |
| Prompt resolution chain     | [specs/prompts.md](specs/prompts.md)                            | `internal/prompt/prompts.go`, `frontmatter.go` | —                                                               | ✅ Complete   |
| Prompt authoring helpers    | [specs/commands/prompts-authoring.md](specs/commands/prompts-authoring.md) | `internal/prompt/prompts.go`, `frontmatter.go` | —                                                               | ✅ Complete   |
| `prompts guide`             | [specs/commands/prompts-authoring.md](specs/commands/prompts-authoring.md) | `internal/cli/prompts.go`      | guide fixture (go:embed)                                         | ❌ Missing    |
| `prompts validate`          | [specs/commands/prompts-authoring.md](specs/commands/prompts-authoring.md) | `internal/cli/prompts.go`      | —                                                                | ❌ Missing    |
| Run-time missing-signal warning | [specs/commands/prompts-authoring.md](specs/commands/prompts-authoring.md) | `internal/cli/run.go`        | —                                                                | ❌ Missing    |
| `skill install` (distribution) | [specs/commands/skill.md](specs/commands/skill.md)           | `internal/cli/skill.go` (new)         | `.agents/skills/specralph-prompts/SKILL.md` (committed ✅); build-time embed ❌ | ❌ Missing |

**Registered Commands** (verified `internal/cli/cmd.go:47-50`): `init`, `run`, `version`, `prompts` (+ Cobra `help`/`completion`). No `skill` command.

---

## Phased Plan

> TDD throughout per AGENTS.md: failing tests first for every phase. Coverage gate ≥95% (`make quality`). Mutation testing (`make mutation`) only at final stage.

### Phase 4: Prompt Package Authoring Helpers

**Goal:** Shared, testable content checks in `internal/prompt` so both `validate` and the run-time warning use one implementation.

**Status:** ✅ Complete

**Paths:**
- `internal/prompt/prompts.go`
- `internal/prompt/prompts_test.go` / `prompts_internal_test.go`

**Checklist:**
- [x] `internal/prompt.HasCompletionSignal(text string) bool` — true if `<COMPLETION_SIGNAL>` or `<promise>COMPLETE</promise>` present (placeholder-aware; substring match on prompt *content*)
- [x] `Check` type: `Name`, `Status` (`ok|warn|fail`), `Message`
- [x] `ValidationResult` type: `Path`, `Checks []Check`
- [x] `internal/prompt.ValidatePromptText(text string) []Check` — frontmatter balance check, description check, completion-signal check
- [x] Description extraction: extend `internal/cli/prompts.go:extractDescription` (body-line-only, `:137`) into prompt package supporting frontmatter `description:` OR first non-empty non-heading body line; migrate CLI caller
- [x] Keep `cli.hasCompletionSignal` (`run.go:277`, line-exact agent-output match) untouched — `RunLoop` literal detection must not change (spec constraint)

**Definition of Done:** failing tests written first; `make test` passes; helpers covered by table-driven tests (both signal forms, mixed, absent).

**Implementation notes (2026-09-11):**

- `ValidatePromptText` returns all three checks in spec order (`frontmatter`, `description`, `completion-signal`); `resolve` stays a Phase 5 CLI concern.
- Frontmatter balance reuses the delimiter scan: `ParseFrontMatter` was refactored onto new unexported `splitFrontMatter`/`opensFrontMatter` helpers (no behavior change to existing callers; `FrontMatterSettings` gained `description`).
- `ExtractDescription` returns `""` on invalid frontmatter YAML; the CLI `extractDescription` keeps the distinct `(invalid frontmatter)` sentinel by checking `ParseFrontMatter` first, so `prompts list` output is unchanged.

**Risks/Dependencies:** None — stdlib only. Do not conflate the new content check with the existing output-line check.

---

### Phase 5: `prompts validate <target>` Subcommand

**Goal:** Static validation with per-check report and exit code 1 on any `fail`.

**Status:** ❌ Not started

**Paths:**
- `internal/cli/prompts.go` (new `NewPromptsValidateCommand`, registered next to list/show at `:24-25`)
- `internal/cli/prompts_test.go`
- `internal/prompt/prompts.go` (Phase 4 helpers)

**Checklist:**
- [ ] Target resolution: `.md` suffix or path separator → file path (must exist); else resolve as prompt name via same chain as `prompts show` (built-in first, then `PromptsDir` via `findFileUpwards`)
- [ ] Checks in order, all reported (first `fail` does not stop later checks): `resolve`, `frontmatter`, `description`, `completion-signal`
- [ ] `frontmatter`: unbalanced `---` delimiters → `warn`; empty `description:` in frontmatter → `warn`
- [ ] `completion-signal`: neither signal form → `fail` (loop can never complete)
- [ ] Output format: one line per check (`ok|warn|fail  name: message`), then summary (`N failed, M warnings`), path header first
- [ ] Exit codes: 0 if no `fail`, 1 otherwise; unknown target → `prompt "<target>" not found`, exit 1
- [ ] Check messages never include prompt content (spec security constraint)

**Definition of Done:** all 7 spec verifications from prompts-authoring.md pass; tests cover file target, named target, built-in, missing target, unbalanced frontmatter, signal-less file.

**Risks/Dependencies:** Depends on Phase 4. Named-prompt resolution should reuse `prompts show` resolution, not duplicate it.

**Reference pattern:** `internal/cli/prompts.go` `NewPromptsShowCommand` (`:48`) for subcommand shape and resolution chain.

---

### Phase 6: `prompts guide` Subcommand

**Goal:** Print the embedded authoring contract (static text, stdout, exit 0).

**Status:** ❌ Not started

**Paths:**
- `internal/cli/prompts.go` (new `NewPromptsGuideCommand`)
- `internal/cli/guide.md` (fixture, `go:embed` — spec default)
- `internal/cli/prompts_test.go`

**Checklist:**
- [ ] Guide fixture covering all spec-mandated content: file location (`PromptsDir`, default `$HOME/.ralph`, `<name>.md`, invoked as `ralph run <name>` / `ralph <name>`), frontmatter `description` convention, completion signal contract (placeholder + literal), recommended structure (objective → study inputs → single-task steps → validation → stop condition), `scope` argument behavior (not substituted — stated explicitly)
- [ ] Distribution line pointing at `ralph skill install [dir]` (default `.agents/skills`) — depends on Phase 8 for the command to exist
- [ ] `go:embed` the fixture; print to stdout, no banners, pipe-friendly
- [ ] Register subcommand; appears in `ralph prompts --help`
- [ ] Open question resolved per spec default: fixture file, not inline string

**Definition of Done:** `ralph prompts guide` prints contract, exit 0; test asserts content markers (each spec-required topic present).

**Risks/Dependencies:** Distribution line references Phase 8's command; implement Phase 8 before releasing, or the guide advertises a nonexistent command.

---

### Phase 7: Run-Time Missing-Signal Warning

**Goal:** Warn once on stderr when an executed prompt file cannot ever complete the loop.

**Status:** ❌ Not started

**Paths:**
- `internal/cli/run.go` (hook after `prompt.GetPrompt` at `:77-80`)
- `internal/cli/run_test.go`

**Checklist:**
- [ ] After resolution, if source was a prompt file (`cfg.PromptFile` explicit or `PromptsDir/<name>.md`) and text lacks both signal forms → stderr: `warning: prompt file has no completion signal (<COMPLETION_SIGNAL>); the loop will only stop at max iterations`
- [ ] Inline prompts (`--prompt`) and stdin exempt
- [ ] Built-ins not special-cased (they contain the signal; check is a natural no-op)
- [ ] Warning printed exactly once, before loop start; run proceeds normally

**Definition of Done:** spec verification 7 reproduced as failing test first (signal-less `prompts/review.md` → warning on stderr, run continues); inline/stdin/built-in cases assert no warning.

**Risks/Dependencies:** Depends on Phase 4 (`prompt.HasCompletionSignal`). Resolution currently returns text without provenance — may need the file-source branch to report provenance or re-derive it (smallest change: check in the two file-sourced branches or compare resolved path).

---

### Phase 8: `ralph skill install [dir]` (Related Domain: skill.md)

**Goal:** Distribute the embedded prompt-authoring skill so agents discover these conventions — the guide's distribution pointer and the workflow's discoverability story.

**Status:** ❌ Not started (spec Proposed; skill source committed)

**Paths:**
- `internal/cli/skill.go` (new)
- `.agents/skills/specralph-prompts/SKILL.md` (✅ already committed, canonical source — commit `1a7512f`)
- `Makefile` / build step (dot-directories cannot be `go:embed`-ed directly; spec allows copy/relocate into embedding package at build time)
- `internal/cli/cmd.go:46-51` (register `skill` command)

**Checklist:**
- [ ] Build-time embed of the skill content into the binary
- [ ] `skill install [dir]`: default `.agents/skills`; target `<dir>/specralph-prompts/SKILL.md`; dirs created `0755`, file written `0644`
- [ ] Existing target without `--force` → error `skill already exists at <target> (use --force to overwrite)`, original untouched
- [ ] `--force` overwrites; no backup
- [ ] Success message to stdout: `Installed specralph-prompts skill to <target>`; exit 0
- [ ] Config-independent (does not load `ralph.toml`)
- [ ] Installed file passes `skill-creator` `quick_validate.py` (frontmatter valid, name `specralph-prompts`, description without angle brackets)

**Definition of Done:** all 6 skill.md verifications pass; tests cover happy path, custom dir, collision without/with `--force`.

**Risks/Dependencies:** Embed mechanism is the only non-trivial part (Go `embed` cannot read `.agents/...`). Independent of Phases 4–7; only Phase 6's guide text references it.

---

## Verification Log

### 2026-06-19: Help and Prompt Discovery (from prior plan cycle)

- 2026-06-19: `./bin/ralph --help`, `ralph help prompts`, `ralph prompts list` - help wiring and `prompts list` verified functional.
- 2026-06-19: `ralph prompts show build` - recorded as gap in prior plan (Phase 3).
- (closed later) `feat: add prompts show subcommand` commit `255a808` - `prompts show` implemented; prior plan never updated to reflect this.

### 2026-09-11: Scope Study for prompt-authoring

- 2026-09-11: `git log --oneline -- specs/` - confirmed `1a7512f docs(specs): add prompt authoring and skill install specs` added `specs/commands/prompts-authoring.md`, `specs/commands/skill.md`, `.agents/skills/specralph-prompts/SKILL.md`, plus cross-links into `specs/README.md` and `specs/commands/prompts.md`. Both new specs are `Status: Proposed`.
- 2026-09-11: read `internal/cli/prompts.go` - only `list` + `show` subcommands registered (`:24-25`); no `guide`, no `validate`.

- 2026-09-11: grep `internal/` for `HasCompletionSignal|COMPLETION_SIGNAL` - `cli.HasCompletionSignal` exists only as literal agent-output check wrapper (`run.go:336-339` over `:277`); no placeholder-aware prompt-content check; no `ValidatePromptText`; no run-time warning anywhere.
- 2026-09-11: read `internal/prompt/prompts.go` - resolution chain verified: `customPrompt` (inline) → `stdinPrompt` → `explicitPromptFile` (`cfg.PromptFile`) → `promptFromDir` (`PromptsDir/<name>.md`, `findFileUpwards`) → `bundledPrompt`; `ParseFrontMatter` exists in `internal/prompt/frontmatter.go`.
- 2026-09-11: read `internal/cli/cmd.go:46-51` - registered commands: `init`, `run`, `version`, `prompts`; no `skill` command; `internal/cli/skill.go` absent (glob confirmed).
- 2026-09-11: read `internal/cli/prompts.go:137-164` - `extractDescription` is body-line-only (no frontmatter `description:` support), lives in CLI layer → Phase 4 migration needed.
- 2026-09-11: read `IMPLEMENTATION_PLAN.md` (prior) - stale: scoped to `commands/help`, dated 2026-06-19, contradicts code on `prompts show` → regenerated for current scope.

### 2026-09-11: Phase 4 - Prompt Package Authoring Helpers

- 2026-09-11: TDD RED - `go test ./internal/prompt/` - failed on undefined `prompt.HasCompletionSignal` / `ExtractDescription` / `Check` / `ValidatePromptText` (expected missing-feature failure).
- 2026-09-11: `go test ./internal/prompt/ ./internal/cli/` - all pass after implementation; legacy frontmatter tests unchanged.
- 2026-09-11: `make lint` - 0 issues (fixed funlen/lll/mnd/nlreturn in new code: table hoisted to package var, long messages wrapped, capacity hint dropped).
- 2026-09-11: `make test-coverage` - total 95.8% (gate ≥95%); `internal/prompt` at 97.0%, all new helpers at 100%. The only failing package is `cmd/ralph` — pre-existing `TestSkillsLockPointsToExternalRepos` (verified failing on clean tree via `git stash`).
- 2026-09-11: `make test-race` - no data races in `internal/...` (same pre-existing `cmd/ralph` failure, unrelated).
- 2026-09-11: `make security` - 0 issues; `make arch` - no warnings.
- 2026-09-11: `make build` + `RALPH_PROMPTS_DIR=... ralph prompts list` / `prompts show review` - custom prompt with frontmatter `description` shows the frontmatter description in `list` and frontmatter-stripped body in `show` (CLI migration verified end-to-end).
- 2026-09-11: Fixed pre-existing commit blocker: commit `8458895` removed the `create-readme` skill (only `github/awesome-copilot` entry) without updating `TestSkillsLockPointsToExternalRepos`; the stale expectation failed the pre-commit gate on every commit. Fixed in commit `d19cb4f`.


### 2026-09-11: Test Baseline

- 2026-09-11: baseline unchanged — no source modified in this planning pass; `make quality` to be run as gate for Phase 4 TDD start.

---

## Summary

| Phase | Description                                  | Status      | Completion |
|-------|----------------------------------------------|-------------|------------|
| 1     | Help Command Verification                    | ✅ Complete | 100%       |
| 2     | Prompts Command - List Subcommand            | ✅ Complete | 100%       |
| 3     | Prompts Command - Show Subcommand            | ✅ Complete | 100%       |
| 4     | Prompt Package Authoring Helpers             | ✅ Complete | 100%       |
| 5     | `prompts validate <target>`                  | ❌ Missing  | 0%         |
| 6     | `prompts guide`                              | ❌ Missing  | 0%         |
| 7     | Run-Time Missing-Signal Warning              | ❌ Missing  | 0%         |
| 8     | `ralph skill install` (skill.md)             | ❌ Missing  | 0%         |

**Remaining Effort:** Phases 5–8 (four phases). Recommended order: 5 → 7 → 6+8 (guide and skill install ship together so the guide's distribution line references a real command). All new code TDD; final gate `make quality`, then `make mutation`.

---

## Known Existing Work

- `prompts list` / `prompts show` fully implemented in `internal/cli/prompts.go` (commits `255a808`, lint fix `097de17`); registered in `cmd.go:50`. Frontmatter stripped on show.
- `internal/prompt` authoring helpers (Phase 4, commit `15ca16a`): `HasCompletionSignal` (placeholder-aware content substring), `ValidatePromptText` (returns `[]Check` in order frontmatter/description/completion-signal), `ExtractDescription` (frontmatter `description:` first, else first non-empty non-heading body line; `""` on invalid frontmatter YAML), `Check`/`ValidationResult` types, `StatusOK/Warn/Fail` constants. `cli.hasCompletionSignal` (`run.go:277`) untouched.
- `internal/prompt.FrontMatterSettings` gained `Description`; delimiter scan refactored into `splitFrontMatter`/`opensFrontMatter` (unexported) — `ParseFrontMatter` behavior unchanged.
- `internal/prompt.Prompt` resolution chain (`GetPrompt`): inline → stdin → explicit file → `PromptsDir` file (upwards search via `findFileUpwards`) → bundled build/plan. Banner written for file-sourced prompts.
- `internal/prompt.ParseFrontMatter` (`frontmatter.go`) parses `model`/`agentMode` overrides — reusable for the `frontmatter`/`description` checks.
- `cli.hasCompletionSignal` (`run.go:277`): literal, line-exact check of *agent output* inside `RunLoop` — protected by spec; do not repurpose for content checking.
- `.agents/skills/specralph-prompts/SKILL.md` committed (canonical skill source, commit `1a7512f`); only the embedding/CLI distribution is missing.
- Cobra command registration pattern established in `cmd.go` (`NewPromptsCommand` composition, `cmd.AddCommand`).

## Manual Deployment Tasks

None.
