# Implementation Plan (prompts/build-*)

**Status:** Not Started (0/6 phases) — specs complete and consistent; code has none of `build-subagents`, the `build` alias, or `build-classic` name registration. Only the `build-classic` content contract exists today (as the legacy `build` generator).
**Last Updated:** 2026-09-15
**Primary Spec(s):** [specs/prompts/build-subagents.md](specs/prompts/build-subagents.md), [specs/prompts/build-classic.md](specs/prompts/build-classic.md), [specs/prompts.md](specs/prompts.md)

---

## Quick Reference

| System/Subsystem | Spec | Module/Package | Artifacts | Status |
|------------------|------|----------------|-----------|--------|
| `build-classic` built-in (content contract) | [specs/prompts/build-classic.md](specs/prompts/build-classic.md) | `internal/prompt/prompts.go` (`BuildPrompt`) | — | ✅ Content exists as `BuildPrompt`; ❌ `build-classic` name not registered (`bundledPrompt` accepts only `build`/`plan`, `prompts.go:129-145`) |
| `build-subagents` built-in (batch prompt) | [specs/prompts/build-subagents.md](specs/prompts/build-subagents.md) | `internal/prompt/prompts.go` (new `BuildSubagentsPrompt`) | — | ❌ Missing (grep: no hits in `internal/`) |
| `build` alias rewrite at invocation entry | [specs/prompts/build-subagents.md](specs/prompts/build-subagents.md) (Workflows), [specs/prompts.md](specs/prompts.md) | `internal/cli/run.go` (`runCommandLogic` — shared entry of root cmd and `run` subcommand) | — | ❌ Missing (`GetPrompt` has no rewrite step, `prompts.go:22-46`) |
| `BuildAliasPrompt` config field (flag/env/TOML/overlay) | [specs/configuration.md](specs/configuration.md) (canonical tables), [build-subagents.md](specs/prompts/build-subagents.md) | `internal/config/config.go`, `internal/cli/run.go` (`setupSharedFlags`) | — | ❌ Missing (`Config` struct `config.go:40-59` has no field) |
| `ralph init` unconditional opt-in | [specs/commands/init.md](specs/commands/init.md) | `internal/cli/init.go` (`buildConfigFromAnswers`), `internal/config/writer.go` | — | ❌ Missing (init writes via `WriteConfig` → whole-struct encode; field tag + value suffices) |
| `prompts list/show/validate` alias awareness | [specs/commands/prompts.md](specs/commands/prompts.md) | `internal/cli/prompts.go` | — | ❌ Missing (hardcoded `build`/`plan` cases at `prompts.go:177-181, 201-212, 291-303`) |
| E2E `[prompt-overrides.build]` migration | [specs/config-by-prompt.md](specs/config-by-prompt.md) (keys use resolved name), build-subagents.md Migration notes | `test/e2e/config_by_prompt_test.go`, `test/e2e/config_local_test.go` | — | ❌ Uses `build` key; silently stops applying once alias lands |
| README / user docs | README.md | `README.md` | — | ❌ Documents pre-alias world (`build`/`plan` only, no alias, no `build-alias-prompt`) |

**Related domains already spec-updated (no spec work needed):** `specs/prompts.md`, `specs/configuration.md`, `specs/commands/init.md`, `specs/commands/prompts.md`, `specs/config-by-prompt.md`, `specs/config-local-overlay.md`, `specs/README.md` — all aligned to the alias model in commits `0c17456`, `2c73cd6`, `6ed845e`.

---

## Phased Plan

> TDD throughout per AGENTS.md: failing tests first for every phase. Coverage gate ≥95% (`make quality`). Mutation testing (`make mutation`) only at final stage.

### Phase 1: `BuildAliasPrompt` Config Field (flag / env / TOML / overlay)

**Goal:** The alias target is a first-class config value with full precedence support.

**Status:** ❌ Not started

**Paths:**
- `internal/config/config.go`
- `internal/cli/run.go` (`setupSharedFlags`, `:308-326`)
- `internal/config/config_test.go`

**Checklist:**
- [ ] `Config.BuildAliasPrompt string` with `toml:"build-alias-prompt"` (no `omitempty` — `WriteConfig` must always emit it for init)
- [ ] Flag `--build-alias-prompt` bound to the field in `setupSharedFlags`
- [ ] Env `RALPH_BUILD_ALIAS_PROMPT`: extend `envValues` (`:21-31`), `readEnv` (`:181-195`), `applyConfigValues` (`:197+`) — default `build-classic`, empty resolves to default (spec: "empty value resolves to the default")
- [ ] Local-overlay merge: extend `mergePromptAndLogScalars` (`:334-344`) with `meta.IsDefined("build-alias-prompt")`
- [ ] Precedence tests: flag > env > base TOML > local overlay > default

**Definition of Done:** failing precedence tests written first; `go test ./internal/config/ ./internal/cli/` passes; `make lint` clean.

**Risks/Dependencies:** None — stdlib only. No validation at config load (spec: validated at prompt resolution time, Phase 3).

---

### Phase 2: Built-in Prompt Registry (`build-classic`, `build-subagents`)

**Goal:** `bundledPrompt` serves the three spec built-ins; the batch generator exists per the spec outline.

**Status:** ❌ Not started

**Paths:**
- `internal/prompt/prompts.go` (`bundledPrompt` `:129-145`, new generator)
- `internal/prompt/prompts_test.go`

**Checklist:**
- [ ] `BuildSubagentsPrompt(cfg *config.Config) string` — generated text matches the spec Appendix outline exactly: 10-task cap, at-least-one-subagent-per-task rule, orchestrate-only constraint, per-task validate/commit/plan-update, failed-task record-and-continue, stop-after-batch, signal only when ALL tasks complete; `<SpecsDir>`/`<SpecsIndexFile>`/`<ImplementationPlanName>` substitutions shared with `BuildPrompt` (reuse the index-reference computation, do not duplicate)
- [ ] `bundledPrompt` cases: `build-classic` → `BuildPrompt` + banner naming the resolved built-in; `build-subagents` → `BuildSubagentsPrompt` + banner; `plan` unchanged
- [ ] `build` removed from built-ins: error message lists `build-classic`, `build-subagents`, `plan` (alias rewrite in Phase 3 is what makes plain `build` work)
- [ ] `build-classic` output byte-identical to current `BuildPrompt` output (backwards-compat reference, spec verification: "`ralph build-classic` output equals the former `build` prompt text")
- [ ] Table tests: `build-subagents` contains every spec-mandated marker; `build-classic` markers per build-classic.md verifications

**Definition of Done:** failing tests first; `go test ./internal/prompt/` passes; `make lint` clean.

**Risks/Dependencies:** None — string generation only. Banner text change for `build` (`USING DEFAULT 'BUILD' PROMPT`) is superseded by named banners; e2e `run_command_test.go` asserts `[build]` loop banner (RunLoop prompt name), not this banner — verify no e2e assertion pins the old banner text.

---

### Phase 3: `build` Alias Rewrite at Invocation Entry

**Goal:** `build` rewrites to `BuildAliasPrompt` before resolution, in every entry point, exactly once.

**Status:** ❌ Not started

**Paths:**
- `internal/cli/run.go` (`runCommandLogic` `:40-94` — the single shared entry for bare `ralph`, `ralph build`, `ralph run build`; spec architecture calls this "cmd.go invocation entry" but both Cobra entrypoints converge here)
- `internal/cli/run_test.go`

**Checklist:**
- [ ] Rewrite after `cfg.LoadConfig()` (`:54`), before `prompt.GetPrompt` (`:77`): if `promptName == "build"` and `cfg.BuildAliasPrompt != "build"`, set `promptName = cfg.BuildAliasPrompt`
- [ ] Downstream layers observe only the rewritten name: `prompt-overrides` lookup (`applyEffectiveSettings` `:90`), loop banners, `RunLoop` prompt name — no further plumbing needed, verify via tests
- [ ] Self-alias escape hatch: `build-alias-prompt = "build"` skips rewrite; with `PromptsDir/build.md` present the file is used (pre-alias behavior); without it, resolution fails (`build` is not a built-in)
- [ ] Unknown target `build-alias-prompt = "nope"` fails with the existing prompt-not-found error before loop start; `ralph plan` unaffected
- [ ] Empty/absent value resolves to `build-classic` (Phase 1 default)
- [ ] Inline/stdin/explicit `--prompt-file` sources are unchanged by the rewrite (rewrite only affects name-based resolution steps)

**Definition of Done:** failing tests first (default alias, opt-in target, escape hatch with and without file, unknown target error, `plan` unaffected); `make lint` clean.

**Risks/Dependencies:** Depends on Phases 1–2. Exactly one rewrite — no double application on re-entry paths.

---

### Phase 4: `ralph init` Unconditional Opt-In

**Goal:** Every generated config targets `build-subagents`; key never asked, never omitted, overwritten on re-init.

**Status:** ❌ Not started

**Paths:**
- `internal/cli/init.go` (`buildConfigFromAnswers` — called by `writeInitConfig` `:373-378`)
- `internal/cli/init_test.go`

**Checklist:**
- [ ] `buildConfigFromAnswers` sets `BuildAliasPrompt: "build-subagents"` unconditionally (not in `InitAnswers` — spec: "not part of `InitAnswers`"; not a questionnaire question)
- [ ] Generated TOML contains `build-alias-prompt = "build-subagents"` (whole-struct encode via `config.WriteConfig`/toml encoder — covered by Phase 1 field tag)
- [ ] Re-init overwrite path replaces any previous `build-alias-prompt` value (regeneration always writes the constant)
- [ ] Preview lines unchanged (spec mandates the TOML key, not a preview line)
- [ ] Test: generated config loads through existing config resolution and `ralph build` resolves to `build-subagents` (unit-level: `WriteConfig` output contains the key)

**Definition of Done:** failing tests first; `go test ./internal/cli/` passes.

**Risks/Dependencies:** Depends on Phase 1 (field + tag). Note: existing `ralph.toml` files written before this change lack the key — that is the intended default (`build-classic`), not a migration case.

---

### Phase 5: `prompts` CLI Alias Awareness (list / show / validate)

**Goal:** The prompts surface matches `specs/commands/prompts.md`: three built-ins, an Aliases section, and `build` following the alias.

**Status:** ❌ Not started

**Paths:**
- `internal/cli/prompts.go` (`runPromptsList` `:201-229`, `runPromptsShow` `:291-326`, `resolveValidationTarget` `:163-193`)
- `internal/cli/prompts_test.go`

**Checklist:**
- [ ] `prompts list` built-ins: `build-classic` (single-task description), `build-subagents` (batch description per spec sample), `plan` — replacing the hardcoded two-line block; `Aliases:` section with `build -> <target> (configurable via build-alias-prompt)`
- [ ] `prompts show build` prints the target prompt content (rewrites via the same rule; `prompts show build-classic` / `build-subagents` print their built-ins)
- [ ] `resolveValidationTarget` built-in set extended to the three names, alias-aware for `build` (so `prompts validate build-subagents` reports all checks ok per spec verification)
- [ ] List description text matches spec sample (`build-classic`: "Implement a single task…"; `build-subagents`: "Pick up to 10 tasks… at least one subagent per task")
- [ ] Existing `prompts_test.go` assertions updated: list contains the three built-ins + alias line; show-build still yields "Agent Instructions (Build Mode)" via alias→classic

**Definition of Done:** failing tests first; `make lint` clean; manual `./bin/ralph prompts list` / `show build` / `validate build-subagents` match spec sample outputs.

**Risks/Dependencies:** Depends on Phases 2–3 (shared alias-rewrite helper reused here — single implementation of the rewrite rule, not a duplicate).

---

### Phase 6: Migration, E2E, Docs

**Goal:** External surfaces consistent with the alias model; suite green under `make quality`.

**Status:** ❌ Not started

**Paths:**
- `test/e2e/config_by_prompt_test.go`, `test/e2e/config_local_test.go` (`[prompt-overrides.build]` → `[prompt-overrides.build-classic]`)
- `test/e2e/COVERAGE_MATRIX.md` (routing rows)
- `README.md` (feature bullets `:37`, built-in prompts `:245`, completion-signal section `:407`, quickstart `:440`)
- New e2e coverage for alias behavior

**Checklist:**
- [ ] Migrate `[prompt-overrides.build]` e2e keys to `build-classic` (spec migration note: overrides keyed by resolved name); alternatively cover the escape hatch (`build-alias-prompt = "build"`) in one case to pin legacy-key behavior
- [ ] New e2e: default `ralph build` == `build-classic` output; `build-alias-prompt = "build-subagents"` in TOML → batch prompt emitted; unknown alias target fails before agent execution
- [ ] E2E init coverage: generated TOML contains `build-alias-prompt = "build-subagents"` (init is TTY-gated; use existing non-TTY init test pattern for the config-write portion)
- [ ] README: built-in prompts are `build-classic`/`build-subagents`/`plan`, `build` alias + `build-alias-prompt` documented, `prompts list` sample updated
- [ ] `make quality` full gate (lint, gosec, arch, coverage ≥95%); `make test` (unit + e2e)
- [ ] `make mutation ARGS="internal/prompt internal/cli"` at final stage only

**Definition of Done:** `make quality` and `make test` pass; spec verifications from build-subagents.md "Verifications" section each executed manually and recorded in the log below.

**Risks/Dependencies:** Depends on all prior phases. E2E harness builds the real binary — alias behavior is exercised end-to-end.

---

## Verification Log

### 2026-09-15: Scope Study for prompts/build-*

- 2026-09-15: `git log --oneline -- specs/` — spec-side alias work landed in `0c17456` (build-subagents prompt, build alias, init opt-in specs + cross-spec updates), `2c73cd6` (content contracts), `6ed845e` (moved into `specs/prompts/`). All code-side work outstanding.
- 2026-09-15: read `specs/prompts/build-subagents.md`, `specs/prompts/build-classic.md`, `specs/prompts.md` — resolution chain step 1 is the alias rewrite; built-ins are `build-classic`/`build-subagents`/`plan`; `build-subagents.md` Status: Proposed, `build-classic.md` Status: Implemented (content contract only).
- 2026-09-15: read `internal/prompt/prompts.go` — `GetPrompt` (`:22-46`) has no rewrite step; `bundledPrompt` (`:129-145`) accepts only `build`/`plan` and errors with "(build, plan)"; `BuildPrompt` (`:168-212`) output verified line-by-line against build-classic.md Appendix — identical (content contract satisfied under the old name; `INDEX_REFERENCE` computed at `:169-177`).
- 2026-09-15: grep `internal/` for `BuildAliasPrompt|build-alias-prompt|build-subagents|build-classic|BuiltInPrompts` — zero source hits; only test files reference `build`.
- 2026-09-15: read `internal/cli/cmd.go`, `internal/cli/run.go` — root cmd and `run` subcommand both route into `runCommandLogic` (`run.go:40`); `parsePositionalArgs` defaults promptName to `build` (`:295`); `setupSharedFlags` (`:308-326`) has no alias flag; `applyEffectiveSettings` (`:90`) keys prompt-overrides by promptName (rewritten name flows through automatically).
- 2026-09-15: read `internal/cli/prompts.go` — `runPromptsList` (`:201-229`) hardcodes two built-ins (`build`, `plan`), no Aliases section; `runPromptsShow` (`:291-303`) hardcoded built-in switch, no alias; `resolveValidationTarget` built-in cases at `:177-181`, no alias.
- 2026-09-15: read `internal/config/config.go`, `internal/config/writer.go` — `Config` (`:40-59`) lacks the field; `readEnv`/`applyConfigValues`/`mergePromptAndLogScalars` all need the new key; `WriteConfig` (`writer.go:13`) encodes the whole struct, so the TOML tag alone makes init output complete once the value is set.
- 2026-09-15: read `internal/cli/init.go` — `writeInitConfig` (`:373`) → `buildConfigFromAnswers` → `config.WriteConfig`; `InitAnswers` (`:166-173`) has no alias field (spec says it must not); questionnaire keys (`:58-71`) contain none.
- 2026-09-15: grep `test/e2e` — `config_by_prompt_test.go:51`, `config_local_test.go:85-88` use `[prompt-overrides.build]`; `run_command_test.go:15` asserts loop banner `[build]` (RunLoop name, not the bundled banner — survives alias rewrite only until Phase 6 review; keep an eye on it). `COVERAGE_MATRIX.md` documents `build` routing.
- 2026-09-15: read `IMPLEMENTATION_PLAN.md` (prior) — scoped to `commands/prompt-authoring`, 8/8 complete, dated 2026-09-11; correct for its scope but wrong scope for `prompts/build-*` → regenerated.
- 2026-09-15: grep `README.md` — documents pre-alias built-ins (`build` and `plan`) at `:37`, `:245`, `:407`; no alias/`build-alias-prompt` mention anywhere.
- 2026-09-15: baseline untouched — no source modified in this planning pass; `make quality` to be run as gate for Phase 1 TDD start.

---

## Summary

| Phase | Description | Status | Completion |
|-------|-------------|--------|------------|
| 1 | `BuildAliasPrompt` config field (flag/env/TOML/overlay) | ❌ Not started | 0% |
| 2 | Built-in prompt registry + `build-subagents` generator | ❌ Not started | 0% |
| 3 | `build` alias rewrite at invocation entry | ❌ Not started | 0% |
| 4 | `ralph init` unconditional opt-in | ❌ Not started | 0% |
| 5 | `prompts` CLI alias awareness (list/show/validate) | ❌ Not started | 0% |
| 6 | Migration, e2e, README/docs | ❌ Not started | 0% |

**Remaining Effort:** All 6 phases. Largest single item is the `build-subagents` generator + its content-marker tests (Phase 2); highest-blast-radius item is the e2e `[prompt-overrides.build]` migration (Phase 6). `build-classic` content work is already done (exists as `BuildPrompt`).

---

## Known Existing Work

- `internal/prompt.BuildPrompt` (`prompts.go:168-212`) already produces the exact `build-classic` content contract (verified against spec Appendix, including the awkward `Manual Deployment Tasks` wrapping at `:203-204`). Phase 2 must not touch its output.
- `GetPrompt` resolution chain (`prompts.go:22-46`): inline → stdin → explicit file → `PromptsDir` file (`findFileUpwards`) → bundled. The alias rewrite slots in before this chain at the CLI entry; the chain itself needs no change.
- `applyEffectiveSettings` (`run.go:90`) keys `[prompt-overrides.*]` by promptName — after the rewrite, resolved-name keys apply with zero extra plumbing.
- Run-time missing-signal warning (`run.go:84-87`) and `prompt.HasCompletionSignal` operate on resolved text — unaffected by the alias.
- `prompts list/show/validate` infrastructure (`internal/cli/prompts.go`) is complete from the prior plan cycle; Phase 5 only extends the built-in set and adds alias resolution.
- `config.WriteConfig` (atomic temp-file + rename, `writer.go:13`) encodes the full struct — Phase 4 needs only a field tag plus the constant value in `buildConfigFromAnswers`.
- E2E harness (`test/e2e/harness_test.go`) builds the real binary and a stub agent (`complete_once` mode) — adequate for alias behavior verification without new infrastructure.
- Cobra command registration pattern established in `cmd.go` (`:47-51`); no new commands required.

## Manual Deployment Tasks

None.
