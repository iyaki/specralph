# Implementation Plan (prompts/build-*)

**Status:** In Progress (4/6 phases) — Phase 1 (`93cb6af`): `BuildAliasPrompt` first-class config value. Phase 2 (`d1271b2`): `build-classic`/`build-subagents` registered, `build` removed from built-ins. Phase 3 (`256ff02`): `build` alias rewrite at invocation entry. Phase 4 (`c7de4b9`): `ralph init` unconditional opt-in. Phases 5–6 remain.
**Last Updated:** 2026-09-15
**Primary Spec(s):** [specs/prompts/build-subagents.md](specs/prompts/build-subagents.md), [specs/prompts/build-classic.md](specs/prompts/build-classic.md), [specs/prompts.md](specs/prompts.md)

---

## Quick Reference

| System/Subsystem | Spec | Module/Package | Artifacts | Status |
|------------------|------|----------------|-----------|--------|
| `build-classic` built-in (content contract) | [specs/prompts/build-classic.md](specs/prompts/build-classic.md) | `internal/prompt/prompts.go` (`BuildPrompt`) | — | ✅ Done (`d1271b2`: `bundledPrompt` serves `build-classic`; output byte-identical to former `build` prompt, verified against git HEAD generator) |
| `build-subagents` built-in (batch prompt) | [specs/prompts/build-subagents.md](specs/prompts/build-subagents.md) | `internal/prompt/prompts.go` (`BuildSubagentsPrompt`) | — | ✅ Done (`d1271b2`: generator matches spec Appendix outline exactly; shared `specsStudyLine` helper) |
| `build` alias rewrite at invocation entry | [specs/prompts/build-subagents.md](specs/prompts/build-subagents.md) (Workflows), [specs/prompts.md](specs/prompts.md) | `internal/cli/run.go` (`runCommandLogic` — shared entry of root cmd and `run` subcommand) | — | ✅ Done (`256ff02`: `rewriteBuildAlias` applied once after config load; shared helper reusable by Phase 5) |
| `BuildAliasPrompt` config field (flag/env/TOML/overlay) | [specs/configuration.md](specs/configuration.md) (canonical tables), [build-subagents.md](specs/prompts/build-subagents.md) | `internal/config/config.go`, `internal/cli/run.go` (`setupSharedFlags`) | — | ✅ Done (`93cb6af`: field `toml:"build-alias-prompt"` no omitempty, `--build-alias-prompt` flag, `RALPH_BUILD_ALIAS_PROMPT`, overlay merge, default `build-classic`) |
| `ralph init` unconditional opt-in | [specs/commands/init.md](specs/commands/init.md) | `internal/cli/init.go` (`buildConfigFromAnswers`), `internal/config/writer.go` | — | ✅ Done (`c7de4b9`: constant value in `buildConfigFromAnswers`; Phase 1 tag makes `WriteConfig` always emit the key; overwrite regeneration replaces prior value) |
| `prompts list/show/validate` alias awareness | [specs/commands/prompts.md](specs/commands/prompts.md) | `internal/cli/prompts.go` | — | ❌ Missing (hardcoded `build`/`plan` cases at `prompts.go:177-181, 201-212, 291-303`) |
| E2E `[prompt-overrides.build]` migration | [specs/config-by-prompt.md](specs/config-by-prompt.md) (keys use resolved name), build-subagents.md Migration notes | `test/e2e/config_by_prompt_test.go`, `test/e2e/config_local_test.go` | — | ⚠️ Partial (keys migrated to `build-classic` in `256ff02` — required by the per-commit coverage gate; escape-hatch case + new alias e2e coverage remain in Phase 6) |
| README / user docs | README.md | `README.md` | — | ❌ Documents pre-alias world (`build`/`plan` only, no alias, no `build-alias-prompt`) |

**Related domains already spec-updated (no spec work needed):** `specs/prompts.md`, `specs/configuration.md`, `specs/commands/init.md`, `specs/commands/prompts.md`, `specs/config-by-prompt.md`, `specs/config-local-overlay.md`, `specs/README.md` — all aligned to the alias model in commits `0c17456`, `2c73cd6`, `6ed845e`.

---

## Phased Plan

> TDD throughout per AGENTS.md: failing tests first for every phase. Coverage gate ≥95% (`make quality`). Mutation testing (`make mutation`) only at final stage.

### Phase 1: `BuildAliasPrompt` Config Field (flag / env / TOML / overlay)

**Goal:** The alias target is a first-class config value with full precedence support.

**Status:** ✅ Complete (`93cb6af`)

**Paths:**
- `internal/config/config.go`
- `internal/cli/run.go` (`setupSharedFlags`, `:308-326`)
- `internal/config/config_test.go`

**Checklist:**
- [x] `Config.BuildAliasPrompt string` with `toml:"build-alias-prompt"` (no `omitempty` — `WriteConfig` must always emit it for init)
- [x] Flag `--build-alias-prompt` bound to the field in `setupSharedFlags`
- [x] Env `RALPH_BUILD_ALIAS_PROMPT`: extend `envValues` (`:21-31`), `readEnv` (`:181-195`), `applyConfigValues` (`:197+`) — default `build-classic`, empty resolves to default (spec: "empty value resolves to the default")
- [x] Local-overlay merge: extend `mergePromptAndLogScalars` (`:334-344`) with `meta.IsDefined("build-alias-prompt")`
- [x] Precedence tests: flag > env > base TOML > local overlay > default

**Definition of Done:** failing precedence tests written first; `go test ./internal/config/ ./internal/cli/` passes; `make lint` clean.

**Risks/Dependencies:** None — stdlib only. No validation at config load (spec: validated at prompt resolution time, Phase 3).

---

### Phase 2: Built-in Prompt Registry (`build-classic`, `build-subagents`)

**Goal:** `bundledPrompt` serves the three spec built-ins; the batch generator exists per the spec outline.

**Status:** ✅ Complete (`d1271b2`)

**Paths:**
- `internal/prompt/prompts.go` (`bundledPrompt` `:129-145`, new generator)
- `internal/prompt/prompts_test.go`

**Checklist:**
- [x] `BuildSubagentsPrompt(cfg *config.Config) string` — generated text matches the spec Appendix outline exactly: 10-task cap, at-least-one-subagent-per-task rule, orchestrate-only constraint, per-task validate/commit/plan-update, failed-task record-and-continue, stop-after-batch, signal only when ALL tasks complete; `<SpecsDir>`/`<SpecsIndexFile>`/`<ImplementationPlanName>` substitutions shared with `BuildPrompt` (reuse the index-reference computation, do not duplicate)
- [x] `bundledPrompt` cases: `build-classic` → `BuildPrompt` + banner naming the resolved built-in; `build-subagents` → `BuildSubagentsPrompt` + banner; `plan` unchanged
- [x] `build` removed from built-ins: error message lists `build-classic`, `build-subagents`, `plan` (alias rewrite in Phase 3 is what makes plain `build` work)
- [x] `build-classic` output byte-identical to current `BuildPrompt` output (backwards-compat reference, spec verification: "`ralph build-classic` output equals the former `build` prompt text")
- [x] Table tests: `build-subagents` contains every spec-mandated marker; `build-classic` markers per build-classic.md verifications

**Definition of Done:** failing tests first; `go test ./internal/prompt/` passes; `make lint` clean.

**Risks/Dependencies:** None — string generation only. Banner text change for `build` (`USING DEFAULT 'BUILD' PROMPT`) is superseded by named banners; e2e `run_command_test.go` asserts `[build]` loop banner (RunLoop prompt name), not this banner — verify no e2e assertion pins the old banner text.

---

### Phase 3: `build` Alias Rewrite at Invocation Entry

**Goal:** `build` rewrites to `BuildAliasPrompt` before resolution, in every entry point, exactly once.

**Status:** ✅ Complete (`256ff02`)

**Paths:**
- `internal/cli/run.go` (`runCommandLogic` `:40-94` — the single shared entry for bare `ralph`, `ralph build`, `ralph run build`; spec architecture calls this "cmd.go invocation entry" but both Cobra entrypoints converge here)
- `internal/cli/run_test.go`

**Checklist:**
- [x] Rewrite after `cfg.LoadConfig()` (`:54`), before `prompt.GetPrompt` (`:77`): if `promptName == "build"` and `cfg.BuildAliasPrompt != "build"`, set `promptName = cfg.BuildAliasPrompt`
- [x] Downstream layers observe only the rewritten name: `prompt-overrides` lookup (`applyEffectiveSettings` `:90`), loop banners, `RunLoop` prompt name — no further plumbing needed, verify via tests
- [x] Self-alias escape hatch: `build-alias-prompt = "build"` skips rewrite; with `PromptsDir/build.md` present the file is used (pre-alias behavior); without it, resolution fails (`build` is not a built-in)
- [x] Unknown target `build-alias-prompt = "nope"` fails with the existing prompt-not-found error before loop start; `ralph plan` unaffected
- [x] Empty/absent value resolves to `build-classic` (Phase 1 default)
- [x] Inline/stdin/explicit `--prompt-file` sources are unchanged by the rewrite (rewrite only affects name-based resolution steps)

**Definition of Done:** failing tests first (default alias, opt-in target, escape hatch with and without file, unknown target error, `plan` unaffected); `make lint` clean.

**Risks/Dependencies:** Depends on Phases 1–2. Exactly one rewrite — no double application on re-entry paths.

---

### Phase 4: `ralph init` Unconditional Opt-In

**Goal:** Every generated config targets `build-subagents`; key never asked, never omitted, overwritten on re-init.

**Status:** ✅ Complete (`c7de4b9`)

**Paths:**
- `internal/cli/init.go` (`buildConfigFromAnswers` — called by `writeInitConfig` `:373-378`)
- `internal/cli/init_test.go`

- [x] `buildConfigFromAnswers` sets `BuildAliasPrompt: "build-subagents"` unconditionally (not in `InitAnswers` — spec: "not part of `InitAnswers`"; not a questionnaire question)
- [x] Generated TOML contains `build-alias-prompt = "build-subagents"` (whole-struct encode via `config.WriteConfig`/toml encoder — covered by Phase 1 field tag)
- [x] Re-init overwrite path replaces any previous `build-alias-prompt` value (regeneration always writes the constant)
- [x] Preview lines unchanged (spec mandates the TOML key, not a preview line)
- [x] Test: generated config loads through existing config resolution and `ralph build` resolves to `build-subagents` (unit-level: `WriteConfig` output contains the key)

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
- [x] Migrate `[prompt-overrides.build]` e2e keys to `build-classic` (spec migration note: overrides keyed by resolved name) — done in `256ff02` (per-commit coverage gate required green e2e); alternatively cover the escape hatch (`build-alias-prompt = "build"`) in one case to pin legacy-key behavior
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

### 2026-09-15: Phase 1 — `BuildAliasPrompt` Config Field

- 2026-09-15: `go test ./internal/config/ -run 'TestLoadConfigDefaults|TestLoadConfigBuildAliasPromptPrecedence|TestLoadConfigWithOverlayScalars|TestWriteConfigAlwaysEmitsBuildAliasPrompt'` and `go test ./internal/cli/ -run TestSetupSharedFlagsRegistersBuildAliasPrompt` (pre-implementation) — RED confirmed: `BuildAliasPrompt undefined` compile errors and `unknown flag: --build-alias-prompt`; failures are the missing feature, not test typos.
- 2026-09-15: implementation — `defaultBuildAliasPrompt = "build-classic"` const; `Config.BuildAliasPrompt` (`toml:"build-alias-prompt"`, no `omitempty`); `envValues.buildAliasPrompt` + `readEnv` `RALPH_BUILD_ALIAS_PROMPT`; `applyConfigValues` `resolveString` with default; `mergePromptAndLogScalars` overlay key; `setupSharedFlags` `--build-alias-prompt` binding. Test support: `clearConfigEnv` clears the new var; `assertDefaultCoreFields` asserts the `build-classic` default.
- 2026-09-15: `go test ./internal/config/ ./internal/cli/` — PASS. New coverage: `TestLoadConfigBuildAliasPromptPrecedence` (file < env < flag), overlay-wins assertion in `TestLoadConfigWithOverlayScalars`, `TestWriteConfigAlwaysEmitsBuildAliasPrompt` (key emitted even when empty — init contract), `TestSetupSharedFlagsRegistersBuildAliasPrompt` (flag on both `run` and root commands).
- 2026-09-15: `make lint` — 0 issues after wrapping the two over-limit lines to match existing repo style (multi-line `resolveString`, continuation-string flag usage).
- 2026-09-15: commit `93cb6af` — `feat(config): add BuildAliasPrompt field with full precedence support` (6 files, +129).

### 2026-09-15: Phase 2 — Built-in Prompt Registry (`build-classic`, `build-subagents`)

- 2026-09-15: `go test ./internal/prompt/` (pre-implementation) — RED confirmed: `BuildSubagentsPrompt undefined` (missing feature, not test typos). New tests: exact spec-outline match, `NoSpecsIndex` parenthetical omission, spec-marker table (6 classic + 9 subagents markers), bundled resolution byte-equality + named banners, `build` no longer a built-in with error listing all three names.
- 2026-09-15: implementation — `bundledPrompt` cases `build-classic`/`build-subagents`/`plan` with named banners (`USING DEFAULT 'BUILD-CLASSIC' PROMPT` etc.); index-reference computation extracted into `specsStudyLine` shared by `BuildPrompt` and `BuildSubagentsPrompt` (no duplication); not-found error lists the three built-ins.
- 2026-09-15: byte-identical check — throwaway test compared refactored `BuildPrompt` against the git-HEAD generator across 4 configs (with index, `NoSpecsIndex`, no index file, empty) — PASS, throwaway deleted. build-classic.md content contract unchanged.
- 2026-09-15: grep for `USING DEFAULT|pre-bundled prompts` in test sources — no e2e/unit assertion pins the old banner or old error text (only `.ralph/logs` runtime logs). Phase 2 risk note cleared.
- 2026-09-15: migrated `internal/cli` unit tests that invoke bare `build` with no prompt source (would fail until Phase 3 restores it): `cmd_test.go` debug happy path and default-name routing (now backed by a `~/.ralph/build.md` prompts-dir file), `run_test.go` debug happy path (`[build-classic]` banner) and built-in-silence case. `cmd_config_test.go` untouched (explicit `--prompt-file` short-circuits before bundled resolution).
- 2026-09-15: `go test ./internal/...` — PASS (all packages). Coverage over `./internal/...`: 95.7% (gate ≥95%).
- 2026-09-15: `make lint` — 0 issues (wrapped 5 over-length lines per repo style).
- 2026-09-15: e2e not run — expected red on the 5 cases resolving bundled `build` (`config_precedence` NoSpecsIndex, `plan_flags`, `specs_flags` ×2) and the 2 `[prompt-overrides.build]` cases; Phase 3's rewrite heals the first group, Phase 6's migration the second. This is the documented interim state between phases.
- 2026-09-15: commit `d1271b2` — `feat(prompt): register build-classic and build-subagents built-in prompts` (4 files, +228/−26).

### 2026-09-15: Phase 3 — `build` Alias Rewrite at Invocation Entry

- 2026-09-15: `go test ./internal/cli/ -run TestRunBuildAliasRewrite` (pre-implementation) — RED confirmed: 4/7 cases fail with `prompt file not found for 'build'` (default alias, config target, flag override, unknown target — none observe a rewrite); escape-hatch and plan-unaffected guards green as expected.
- 2026-09-15: implementation — `rewriteBuildAlias(promptName, aliasTarget)` applied once in `runCommandLogic` after `applyEnvFlagOverrides`, before `GetPrompt`; escape hatch (`target == "build"`) and non-`build` names pass through unchanged. Single shared helper in `run.go`, reusable by Phase 5.
- 2026-09-15: same tests (post-implementation) — 7/7 PASS: default `build` → `BUILD-CLASSIC` banner + `[build-classic]` loop banner; `build-alias-prompt = "build-subagents"` → batch prompt markers; `--build-alias-prompt build-classic` overrides TOML; escape hatch uses `$HOME/.ralph/build.md` (`USING PROMPT FILE`, `[build]` banner) and fails resolution without it; unknown target `nope` fails before `Starting Specralph`; `plan` unaffected.
- 2026-09-15: migrated pre-alias expectations — `TestNewRalphCommandDefaultToBuild` (dropped the Phase 2 `build.md` crutch; asserts `[build-classic]`), 3× `[prompt-overrides.build]` unit keys → `build-classic` (`cmd_config_test.go`), e2e `RunDefaultsToBuildPrompt` banner → `[build-classic]`.
- 2026-09-15: `make test-e2e` — 2 failures (`TestE2EConfigByPromptOverrideFromConfigApplies`, `TestE2EConfigLocalOverlay_PromptOverridesDeepMerge`): `[prompt-overrides.build]` keys stopped applying — exactly the Phase 6 risk, but the per-commit coverage gate blocks red e2e. Key migration pulled forward into this commit: both e2e files key overrides by `build-classic` while still invoking positional `build` (proving invocation `build` → resolved-name keys end-to-end). E2E fully green afterwards.
- 2026-09-15: `go test ./internal/...` — PASS; coverage 95.8% (gate ≥95%). `make lint` — 0 issues (after extracting the alias test table into a package var for `funlen`).
- 2026-09-15: commit `256ff02` — `feat(cli): rewrite build alias at invocation entry` (7 files, +184/−20).

### 2026-09-15: Phase 4 — `ralph init` Unconditional Opt-In

- 2026-09-15: `go test ./internal/cli/ -run 'TestInitCommandWritesBuildAliasPrompt|TestInitCommandOverwriteReplacesBuildAliasPrompt'` (pre-implementation) — RED confirmed: both fail with `build-alias-prompt = ""` in the generated TOML (missing feature, not test typos). New tests: default-init TOML contains the key and decodes through config resolution to `build-subagents`; confirmed overwrite of a `build-alias-prompt = "build-classic"` file regenerates with `build-subagents`.
- 2026-09-15: implementation — single line: `buildConfigFromAnswers` sets `BuildAliasPrompt: "build-subagents"`. Key never asked (not in `InitAnswers`/questionnaire), never omitted (Phase 1 tag without `omitempty`), overwrite always rewrites the constant; seed path deliberately ignores any existing value.
- 2026-09-15: same tests (post-implementation) — PASS; `go test ./internal/...` — PASS; coverage 95.8% (gate ≥95%); `make lint` — 0 issues.
- 2026-09-15: commit `c7de4b9` — `feat(cli): opt generated configs into build-subagents unconditionally` (2 files, +43).

---

## Summary

| Phase | Description | Status | Completion |
|-------|-------------|--------|------------|
| 1 | `BuildAliasPrompt` config field (flag/env/TOML/overlay) | ✅ Complete | 100% |
| 2 | Built-in prompt registry + `build-subagents` generator | ✅ Complete | 100% |
| 3 | `build` alias rewrite at invocation entry | ✅ Complete | 100% |
| 4 | `ralph init` unconditional opt-in | ✅ Complete | 100% |
| 5 | `prompts` CLI alias awareness (list/show/validate) | ❌ Not started | 0% |
| 6 | Migration, e2e, README/docs | ❌ Not started | 0% |

**Remaining Effort:** Phases 5–6 (`prompts` CLI alias awareness, docs/mutation). Phase 5 reuses `rewriteBuildAlias` (`internal/cli/run.go`) — single implementation of the rewrite rule. Phase 6 remainder: escape-hatch e2e case, new alias e2e coverage, init e2e, README.

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
