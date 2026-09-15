# Build with Subagents Prompt

Status: Proposed

## Overview

### Purpose

- Define the built-in `build-subagents` prompt: batch execution of up to 10 tasks from the implementation plan, with at least one subagent per task.
- Define the `build` prompt alias and the optional `build-alias-prompt` configuration that controls which prompt the alias resolves to.
- Have `ralph init` opt new projects into `build-subagents` by always writing `build-alias-prompt` in the generated config.
- Preserve the previous single-task build prompt as `build-classic`.

### Goals

- Keep `build` resolving to the legacy single-task prompt by default for backwards compatibility.
- Keep `build-classic` behaviorally identical to the former `build` prompt.
- Opt new projects into `build-subagents` via `ralph init`, which always writes `build-alias-prompt = "build-subagents"`.
- Specify the alias rewrite rule, its entry points, and failure behavior.
- Specify the `build-subagents` prompt content contract: batch cap, subagent delegation, per-task validation/commit/plan update, and completion signal gating.

### Non-Goals

- Making Ralph spawn or manage subagents itself; subagents are dispatched by the underlying agent CLI within a single loop iteration.
- Introducing a generic alias map or additional aliases (future work).
- Making the batch size (10) configurable (future work).
- Changing prompt source precedence for non-`build` names (see [prompts.md](prompts.md)).
- Changing command routing (see [commands/run.md](commands/run.md)).

### Scope

- In scope: the `build-subagents` and `build-classic` built-in prompts, the `build` alias rewrite, the `BuildAliasPrompt` config field, the `ralph init` always-written key, and built-in prompt listing.
- Out of scope: agent adapters, per-prompt model/mode override mechanics (see [config-by-prompt.md](config-by-prompt.md)), loop execution internals.

## Architecture

### Module/package layout (tree format)

```
internal/
  prompt/
    prompts.go    (BuildPrompt backs build-classic; new BuildSubagentsPrompt; built-in registry)
internal/
  cli/
    cmd.go        (alias rewrite at invocation entry)
    prompts.go    (prompts list/show/validate are alias-aware)
  config/
    config.go     (BuildAliasPrompt field)
```

### Component diagram (ASCII)

```
+--------------------------------+
| Invocation name: build         |
| (ralph | ralph build | run)    |
+---------------+----------------+
                |
                v
+---------------+----------------+
| Alias rewrite (CLI entry)      |
| build -> BuildAliasPrompt      |
| (default: build-classic)       |
+---------------+----------------+
                |
                v
+---------------+----------------+     +----------------------------+
| Prompt resolver                |---->| built-in build-subagents   |
| (precedence unchanged)         |     | built-in build-classic     |
+--------------------------------+     | built-in plan              |
                                       +----------------------------+
```

### Data flow summary

1. The CLI determines the invocation prompt name (`build` for bare `ralph`, `ralph build [scope]`, and `ralph run build [scope]`).
2. If the name is `build` and the alias target is not `build`, the name is rewritten to `BuildAliasPrompt` (default `build-classic`).
3. Prompt resolution proceeds unchanged with the rewritten name: inline > stdin > explicit prompt file > `PromptsDir/<name>.md` > built-ins (`build-classic`, `build-subagents`, `plan`).
4. The generated `build-subagents` prompt instructs batch execution through subagents; loop execution and completion detection are unchanged.

## Data model

### Core Entities

- `BuildAliasPrompt` (config field, string)
  - Flag `--build-alias-prompt`, env `RALPH_BUILD_ALIAS_PROMPT`, TOML key `build-alias-prompt`.
  - Optional; default `build-classic`; an empty value resolves to the default.
- Built-in prompt registry
  - `build-classic` — legacy single-task build prompt; content unchanged from the former `build` prompt.
  - `build-subagents` — batch build prompt (subagent per task).
  - `plan` — unchanged.
- Alias rule
  - Applies only to the literal invocation name `build`, in every entry point.
  - Exactly one rewrite; a target equal to `build` disables the rewrite (escape hatch).
  - Downstream layers (prompt file lookup, built-in selection, `[prompt-overrides.<name>]` keys, output banners) observe only the rewritten name.

### Relationships

- `BuildAliasPrompt` resolves through the standard configuration precedence: flags > env vars > config file > defaults.
- Front matter of the target prompt and `[prompt-overrides.<target>]` apply to the resolved prompt, keyed by the rewritten name.

### Persistence Notes

- None. Alias resolution is runtime-only; built-in prompt content is generated.

## Workflows

### Default invocation (`ralph` or `ralph build [scope]`)

1. Invocation name `build` is rewritten to the alias target (default `build-classic`).
2. No inline/stdin/explicit-file source applies.
3. The bundled generator emits the target prompt: `build-classic` by default, or `build-subagents` when configured (for example by `ralph init`); the banner names the resolved built-in prompt.

### Legacy behavior (`ralph build-classic`)

1. Name `build-classic` is not the alias name; no rewrite occurs.
2. The bundled generator emits the legacy single-task build prompt, identical to the former `build` prompt (including the `<COMPLETION_SIGNAL>` placeholder).

### Opt-in invocation (`build-alias-prompt = "build-subagents"`)

1. `ralph build` rewrites to `build-subagents`.
2. The bundled generator emits the batch prompt.

### Init opt-in (`ralph init`)

1. `ralph init` always writes `build-alias-prompt = "build-subagents"` into the generated TOML; the key is not part of the questionnaire and is never omitted.
2. Re-initializing an existing config overwrites any previous `build-alias-prompt` value with `build-subagents`.
3. Repos initialized this way run `ralph build` against `build-subagents`; all other setups keep the `build-classic` default.

### Self-alias escape hatch (`build-alias-prompt = "build"`)

1. The rewrite is skipped; `build` resolves through the normal chain.
2. `PromptsDir/build.md` (searching upward) is used when present, restoring pre-alias file-override behavior.
3. Without such a file, resolution fails: `build` is not a built-in prompt name.

### Unknown alias target

1. `build-alias-prompt = "nope"` matches no prompt file or built-in.
2. Resolution fails fast with the existing prompt-not-found error, before the loop starts.

### Batch execution (`build-subagents` run)

1. The agent studies the specs directory and the implementation plan (same study instructions as `build-classic`).
2. The agent selects up to 10 pending tasks in plan order; fewer when fewer are pending; when none are pending, it verifies the plan and emits the completion signal.
3. For each selected task, the agent dispatches at least one subagent using the agent CLI's native subagent mechanism, with a self-contained brief: the task, relevant spec paths, and validation commands.
4. The main agent orchestrates only (dispatch, integrate, validate); each task is implemented by its subagent(s), not directly in the main context.
5. After each task completes: validate the change, update the plan, and commit code and plan update together. A failed task is recorded in the plan as not complete and does not block the remaining selected tasks.
6. After the batch: emit `<COMPLETION_SIGNAL>` only when ALL plan tasks are complete and passing; otherwise stop without the signal so the next loop iteration picks up the next batch.

## APIs

- `internal/prompt.BuildSubagentsPrompt(cfg *config.Config) string` — generates the batch prompt.
- `internal/prompt.BuildPrompt(cfg *config.Config) string` — unchanged; backs `build-classic`.
- `internal/prompt.BuiltInPrompts() []PromptInfo` — returns `build-classic`, `build-subagents`, `plan`.

## Client SDK Design

- Not applicable.

## Configuration

### Field reference

`BuildAliasPrompt` — flag `--build-alias-prompt`, env `RALPH_BUILD_ALIAS_PROMPT`, TOML key `build-alias-prompt`, default `build-classic` — is owned by [configuration.md](configuration.md), which holds the canonical source tables and precedence.

- The target is validated at prompt resolution time, not at config load: an unknown name produces the prompt-not-found error.
- The value is a prompt name (built-in or `PromptsDir` file name), not a file path.

### Interactions

- `[prompt-overrides.<name>]` keys must use the resolved name (for example `build-subagents`), not `build`; `build` is no longer a prompt name.
- `MaxIterations` still bounds loop iterations; with the default 25 and batches of up to 10 tasks, a single run can cover up to 250 planned tasks.

## Permissions

- Unchanged from prompt resolution and agent execution.

## Security Considerations

- Prompt content remains sensitive (see [prompts.md](prompts.md)); subagent briefs may quote plan text, so treat agent output and logs as sensitive.
- No new external surface is introduced.

## Dependencies

- Standard library only; no new dependencies.

## Open Questions / Risks

- Batch size 10 is hardcoded; add `build-subagents-max-tasks` only if requested.
- Agents without a native subagent mechanism cannot honor the delegation requirement; Ralph cannot force subagent usage. The prompt mandates subagents and does not define a fallback; such agents are expected to degrade to sequential main-context execution.
- Migration: `[prompt-overrides.build]` and `PromptsDir/build.md` overrides stop applying to `ralph build`; retarget the key/file to the alias target or set `build-alias-prompt = "build"`.

## Verifications

- Under default configuration, `ralph` and `ralph build` produce prompt text equal to `ralph build-classic` (backwards compatible).
- `ralph build-classic` output equals the former `build` prompt text.
- `ralph --build-alias-prompt build-classic build` produces the legacy text; `RALPH_BUILD_ALIAS_PROMPT=build-classic ralph build` is equivalent; the flag wins over the env var.
- `build-alias-prompt = "nope"` with `ralph build` fails with prompt-not-found before agent execution; `ralph plan` is unaffected.
- `build-alias-prompt = "build"` with an existing `PromptsDir/build.md` makes `ralph build` use that file (pre-alias behavior).
- `ralph prompts list` shows built-ins `build-classic`, `build-subagents`, `plan`, plus an alias line for `build`.
- `ralph prompts show build` prints the target prompt content.
- The generated `build-subagents` text contains: the 10-task cap, the at-least-one-subagent-per-task rule, per-task validation/commit/plan update, the completion-signal-only-when-all-complete stop condition, and the `<COMPLETION_SIGNAL>` placeholder (`ralph prompts validate build-subagents` reports all checks ok).
- `ralph init` generates a `ralph.toml` containing `build-alias-prompt = "build-subagents"` without asking; afterwards `ralph build` produces the `build-subagents` prompt text.

## Appendices

### Generated prompt outline (`build-subagents`)

```markdown
# Agent Instructions (Build Mode with Subagents)

- Study `<SpecsDir>/*` (including `<SpecsDir>/<SpecsIndexFile>` and related specs).
- Study `<ImplementationPlanName>` and select the next pending tasks, up to a maximum of 10.
- If no tasks are pending, verify the plan is complete and reply with `<COMPLETION_SIGNAL>`.

## Task Execution

- Dispatch at least one subagent per selected task using the agent CLI's native
  subagent mechanism.
- Give each subagent a self-contained brief: the task, relevant spec paths, and
  validation commands.
- Orchestrate only: tasks are implemented by subagents, not directly in the main context.
- After each task: validate, update `<ImplementationPlanName>`, and commit code and
  plan update together.
- If a task fails, record it in the plan as not complete and continue with the
  remaining selected tasks.

## Stop Condition

- After the selected batch, stop. Do NOT pick up more tasks in the same run.
- If and only if ALL stories are complete and passing, reply with `<COMPLETION_SIGNAL>`.

## IMPORTANT

- Before changes, search the codebase. Do NOT assume functionality is missing.
- Use the verification log format: `YYYY-MM-DD: <command or URL> - <result>`.
- Keep a `Manual Deployment Tasks` section in the plan and use `None` when there are no tasks.
- You may add temporary logging as needed and remove if no longer needed.
```

`<SpecsDir>`, `<SpecsIndexFile>`, and `<ImplementationPlanName>` are substituted from config exactly as in `build-classic`; `<COMPLETION_SIGNAL>` is replaced by the loop at runtime.

### Migration notes

- `[prompt-overrides.build]` → rename the key to the alias target (`build-classic` by default, `build-subagents` in `ralph init`-generated configs).
- `PromptsDir/build.md` → rename the file to the target name, or set `build-alias-prompt = "build"` to keep file-based overrides working.

## Related Specifications

- [prompts.md](prompts.md) — resolution precedence, completion signal, built-in prompt summaries.
- [configuration.md](configuration.md) — `BuildAliasPrompt` field and precedence.
- [commands/prompts.md](commands/prompts.md) — listing/showing/validating prompts and the alias line.
- [commands/run.md](commands/run.md) — routing (`ralph` → `run build`).
- [config-by-prompt.md](config-by-prompt.md) — per-prompt overrides keyed by resolved prompt name.
- [commands/init.md](commands/init.md) — `ralph init` writes `build-alias-prompt = "build-subagents"` unconditionally.
