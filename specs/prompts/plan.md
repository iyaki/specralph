# Plan Prompt

Status: Implemented

## Overview

### Purpose

- Define the instruction-level content contract of the built-in `plan` prompt, which generates or updates the implementation plan.

### Goals

- Make generated prompt text reviewable and testable without reading generator code.
- Specify config substitutions and scope interpolation.

### Non-Goals

- Prompt resolution order and completion signal rules (see [../prompts.md](../prompts.md)).
- Plan execution — the prompt plans only; implementation is driven by the build prompts.

### Scope

- In scope: generated instruction content, config substitutions, scope interpolation.
- Out of scope: loop execution, agent adapters, alias resolution.

## Architecture

### Module/package layout (tree format)

```
internal/
  prompt/
    prompts.go    (PlanPrompt + plan* line helpers)
```

### Data flow summary

1. Prompt name `plan` falls through resolution to the bundled generator.
2. The `scope` argument is interpolated verbatim into the `Scope:` line and the plan-title requirement (rendered even when empty).
3. Config values are substituted into fixed instruction lines; `<COMPLETION_SIGNAL>` is replaced by the loop at runtime.

## Data model

### Substitutions

| Placeholder | Source | Notes |
| ----------- | ------ | ----- |
| `<SpecsDir>` | `SpecsDir` | |
| `<ImplementationPlanName>` | `ImplementationPlanName` | |
| `<Scope>` | `scope` CLI argument | verbatim; may be empty |

## Workflows

### Generate `plan`

1. Prompt name `plan` falls through resolution to the bundled generator.
2. The generator emits the instruction contract in the Appendix with substitutions applied.

## APIs

- `internal/prompt.PlanPrompt(cfg *config.Config, scope string) string`

## Client SDK Design

- Not applicable.

## Configuration

- Uses `SpecsDir` and `ImplementationPlanName`; definitions in [../configuration.md](../configuration.md).

## Permissions

- Read access to specs, source, and implementation plan files is required by the instructed workflow, not by generation itself.

## Security Considerations

- Scope text is user input embedded into the agent prompt; treat it as untrusted content passed to the agent CLI.

## Dependencies

- Standard library only.

## Open Questions / Risks

- An empty `scope` still renders the `Scope:` line (possibly empty).
- Plan size and phase count are uncapped.

## Verifications

- `ralph prompts show plan` starts with `# Agent Instructions (Planning Mode)` and contains: "Plan only. Do NOT implement anything.", the `[x]`-only-when-verified rule, the phased plan requirements (goal, status, paths, checklist, Definition of Done, risks), the verification log format `YYYY-MM-DD: <command or URL> - <result>`, the `Known Existing Work` section, the `Manual Deployment Tasks` section with the exact `None` fallback, and `<COMPLETION_SIGNAL>`.
- `ralph run plan my-feature` renders `Scope: my-feature` and instructs a plan titled `Implementation Plan (my-feature)`.

## Appendices

### Instruction contract

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

- [../prompts.md](../prompts.md) — resolution precedence and completion signal rules.
- [build-classic.md](build-classic.md) and [build-subagents.md](build-subagents.md) — the build prompts that execute the plan.
- [../commands/prompts.md](../commands/prompts.md) — listing and viewing built-in prompts.
- [../configuration.md](../configuration.md) — substituted config fields.
