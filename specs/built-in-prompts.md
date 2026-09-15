# Built-in Prompts

Status: Implemented

## Overview

### Purpose

- Define the content contracts of the built-in `build-classic` and `plan` prompts at the instruction level.
- Specify the config substitutions each prompt performs and where scope text is interpolated.

### Goals

- Make generated prompt text reviewable and testable without reading generator code.
- Keep one source of truth per fact: resolution and precedence in [prompts.md](prompts.md), `build-subagents` content in [build-subagents.md](build-subagents.md), config field tables in [configuration.md](configuration.md).

### Non-Goals

- Re-defining prompt resolution, the `build` alias, or completion signal rules (see [prompts.md](prompts.md)).
- Changing prompt content; this spec describes implemented contracts.
- Custom prompt authoring and validation (see [commands/prompts-authoring.md](commands/prompts-authoring.md)).

### Scope

- In scope: generated instruction content for `build-classic` and `plan`, config substitutions, scope handling for `plan`.
- Out of scope: loop execution, agent adapters, `build-subagents` (owned by [build-subagents.md](build-subagents.md)).

## Architecture

### Module/package layout (tree format)

```
internal/
  prompt/
    prompts.go    (BuildPrompt, PlanPrompt, plan* line helpers)
```

### Component diagram (ASCII)

```
+---------------------------+
| Prompt resolver           |
| (prompts.md)              |
+-------------+-------------+
              |  bundled fallback
              v
+-------------+-------------+
| Built-in generators       |
| BuildPrompt / PlanPrompt  |
+-------------+-------------+
              |  config substitutions
              v
+-------------+-------------+
| Prompt text +             |
| <COMPLETION_SIGNAL>       |
| (replaced by the loop)    |
+---------------------------+
```

### Data flow summary

1. Prompt resolution falls through to the bundled generator; a banner naming the resolved built-in is written to output (exact banner wording is an implementation detail).
2. Config values are substituted into fixed instruction lines.
3. The `<COMPLETION_SIGNAL>` placeholder is replaced by the loop at runtime; gating rules live in [prompts.md](prompts.md).

## Data model

### Core Entities

- Built-in prompt names covered here: `build-classic`, `plan`. (`build-subagents` is specified in [build-subagents.md](build-subagents.md).)
- Substitutions
  - `<SpecsDir>` — `SpecsDir`
  - `<SpecsIndexFile>` — `SpecsDir/SpecsIndexFile` reference, present only when `SpecsIndexFile` is set and `NoSpecsIndex` is false
  - `<ImplementationPlanName>` — `ImplementationPlanName`
  - `<Scope>` — the `scope` CLI argument (`plan` only; interpolated verbatim, even when empty)

### Relationships

- Generators consume the resolved `Config`; no persistence, no additional sources.

### Persistence Notes

- None. Prompt text is generated per invocation.

## Workflows

### Generate `build-classic`

1. Prompt name `build-classic` reaches the bundled generator (directly, or via the `build` alias when configured).
2. The generator emits the instruction contract in Appendix A with config substitutions applied.

### Generate `plan`

1. Prompt name `plan` reaches the bundled generator.
2. The `scope` argument is interpolated verbatim into the `Scope:` line and into the plan-title requirement.
3. The generator emits the instruction contract in Appendix B.

## APIs

- `internal/prompt.BuildPrompt(cfg *config.Config) string`
- `internal/prompt.PlanPrompt(cfg *config.Config, scope string) string`

## Client SDK Design

- Not applicable.

## Configuration

- Uses `SpecsDir`, `SpecsIndexFile`, `NoSpecsIndex`, and `ImplementationPlanName`; definitions and precedence in [configuration.md](configuration.md).

## Permissions

- Read access to specs and implementation plan files is required by the instructed workflow, not by generation itself.

## Security Considerations

- Generated text embeds configuration paths and scope text; scope is user input that ends up in the agent prompt — treat it as untrusted content passed to the agent CLI.

## Dependencies

- Standard library only.

## Open Questions / Risks

- Should an empty `scope` render a `Scope:` line at all (currently always rendered, possibly empty)?
- Should the plan prompt cap plan size or phase count? Currently uncapped.

## Verifications

- `ralph prompts show build-classic` starts with `# Agent Instructions (Build Mode)` and contains: "pick the single most important task", the do-not-start-another-task stop condition, the verification log format `YYYY-MM-DD: <command or URL> - <result>`, the `Manual Deployment Tasks` requirement, and `<COMPLETION_SIGNAL>`.
- With `--no-specs-index`, the `build-classic` text omits the specs index parenthetical.
- `ralph prompts show plan` starts with `# Agent Instructions (Planning Mode)` and contains: "Plan only. Do NOT implement anything.", the `[x]`-only-when-verified rule, the phased plan requirements (goal, status, paths, checklist, Definition of Done, risks), the verification log format, the `Known Existing Work` section, the `Manual Deployment Tasks` section with the exact `None` fallback, and `<COMPLETION_SIGNAL>`.
- `ralph run plan my-feature` renders `Scope: my-feature` and instructs a plan titled `Implementation Plan (my-feature)`.

## Appendices

### Appendix A — `build-classic` instruction contract

```markdown
# Agent Instructions (Build Mode)

- Study `<SpecsDir>/*`<INDEX_REFERENCE>.
- Study `<ImplementationPlanName>` and pick the single most important task.
- Implement the task
- Validate the implementation
- Commit the changes
- Update the plan
- Commit the update plan
- Stop after the commit

## Stop Condition

- After completing the selected task, stop. Do NOT start another task in the same run.
- If ALL stories are complete and passing, reply with:
  `<COMPLETION_SIGNAL>`

## IMPORTANT

- Before changes, search the codebase. Do NOT assume functionality is missing.
- Implement ONLY one task. Stop after committing.
- Update `<ImplementationPlanName>` when the task is done.
- Use the verification log format: `YYYY-MM-DD: <command or URL> - <result>`.
- Keep a `Manual Deployment Tasks` section in implementation
  the plan and use `None` when there are no tasks.
- You may implement missing functionality if required, but study relevant `<SpecsDir>/*` first.
- You may add temporary logging as needed and remove if no longer needed.
```

`<INDEX_REFERENCE>` is ` (including `<SpecsDir>/<SpecsIndexFile>` and related specs)` when the index is enabled, otherwise empty.

### Appendix B — `plan` instruction contract

Structure:

1. `# Agent Instructions (Planning Mode)` heading and `Scope: <Scope>` line.
2. **Objective** — generate or update `<ImplementationPlanName>` in a structured, phase-based format with: clear status metadata, quick reference tables, phase sections with paths and checklists, verification log entries, summary tables and remaining effort. "Plan only. Do NOT implement anything."
3. **Study and Gap Analysis** — study `<SpecsDir>/*`; study `<ImplementationPlanName>` (if present; it may be incorrect); study relevant source code against specs; use `git` to study recent spec changes for the scope. Rules: do not assume missing — confirm via code search; identify existing work, partial implementations, TODOs, placeholders, skipped/flaky tests, inconsistent patterns; concise but complete, lists and tables over paragraphs; `[x]` only when verified in code, `[ ]` when missing or unverified; regenerate the plan when stale or contradictory; if the scope relates to other domain areas, include them (study their specs and code).
4. **Output Format Requirements** — the plan must contain:
   - *Header*: title `Implementation Plan (<Scope>)`; status line `**Status:** <summary>`; last-updated date `YYYY-MM-DD`; reference to primary spec(s).
   - *Quick Reference*: table mapping systems/subsystems to specs, modules/packages, web packages, migrations/artifacts; `✅` marks implemented items.
   - *Phased Plan*: numbered phases aligned to the spec's domain; each phase has goal, status (if applicable), paths, `[x]`/`[ ]` checklist, Definition of Done (tests run, commands/URLs, files touched), brief risks/dependencies; subsections (e.g., 9.1) with scope-specific paths; "Reference pattern" links when a canonical directory or file exists.
   - *Verification Log*: chronological entries with date, what was verified, exact commands or URLs, tests and results, bug fixes, files touched — format `YYYY-MM-DD: <command or URL> - <result>`.
   - *Summary*: phases table with completion status plus a "Remaining effort" line.
   - *Known Existing Work*: confirmed existing implementations to prevent duplicate work.
   - *Manual Deployment Tasks*: required section for manual production steps; exactly `None` when not applicable.
5. **Stop Condition** — after writing/updating, if `<ImplementationPlanName>` already reflects the current gaps, reply with `<COMPLETION_SIGNAL>`.

## Related Specifications

- [prompts.md](prompts.md) — resolution precedence and completion signal rules.
- [build-subagents.md](build-subagents.md) — the `build-subagents` prompt and the `build` alias.
- [commands/prompts.md](commands/prompts.md) — listing and viewing built-in prompts.
- [configuration.md](configuration.md) — substituted config fields.
