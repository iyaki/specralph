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

Verbatim output of `internal/prompt.PlanPrompt(cfg, scope)` with placeholder config values substituted; the only variable input is `<Scope>`, interpolated verbatim (possibly empty) into the `Scope:` line and the plan-title requirement.

```markdown
# Agent Instructions (Planning Mode)

Scope: <Scope>

## Objective

Generate or update `<ImplementationPlanName>` in a structured, phase-based format with:

- Clear status metadata
- Quick reference tables
- Phase sections with paths and checklists
- Verification log entries
- Summary tables and remaining effort

Plan only. Do NOT implement anything.

## Study and Gap Analysis

- Study `<SpecsDir>/*` to learn application requirements.
- Study `<ImplementationPlanName>` (if present; it may be incorrect).
- Study relevant source code to compare against specs.
- Use `git` to study recent changes on the specs related to the specified current scope.

Rules:

- Do NOT assume missing; confirm via code search first.
- Identify where work already exists, partial implementations, TODOs, placeholders,
  skipped/flaky tests, or inconsistent patterns.
- Keep the plan concise but complete; prefer lists and tables over paragraphs.
- Use `[x]` only when verified in code. Use `[ ]` if missing or unverified.
- Regenerate the plan if it becomes stale, contradictory,
  or significantly out of sync with code.
- If the specified scope has relationships with other domain areas,
  implementation may be needed in those areas as well
  (always study the related specs and code). Include this in the plan.

## Output Format Requirements

Write `<ImplementationPlanName>` using this structure and level of detail:

Header

- Title: `Implementation Plan (<Scope>)`
- Status line: `**Status:** <summary (e.g., "UI Components Complete (39/39)")>`
- Last Updated date: `YYYY-MM-DD`
- Reference to primary spec(s)

Quick Reference

- A table mapping systems/subsystems to:
  - Specs
  - Modules/packages
  - Web packages
  - Migrations or other artifacts
- Use `✅` to mark items already implemented.

Phased Plan

- Use numbered phases (e.g., Phase 9, Phase 10) aligned to the spec's domain.
- Each phase includes:
  - Goal
  - Status (if applicable)
  - Paths (directories or file patterns)
  - Checklist with `[x]` for verified complete and `[ ]` for missing
  - Definition of Done (tests run, commands/URLs, files touched)
  - Risks/Dependencies (brief)
- Break phases into subsections (e.g., 9.1, 9.2)
  with scope-specific paths and item lists.
- Include "Reference pattern" links
  when there's a canonical directory or file to follow.

Verification Log

- A chronological log of verification steps with dates.
- Each entry includes:
  - What was verified (endpoints, commands, builds, tests, UI routes)
  - Exact commands or URLs used
  - Tests run and results
  - Bug fixes discovered (if any)
  - Files touched (if known from code search)
  - Use format: `YYYY-MM-DD: <command or URL> - <result>`

Summary

- Table of phases with completion status
  - "Remaining effort" line summarizing unfinished sections

Known Existing Work

- Brief section listing confirmed existing implementations to prevent duplicate work

Manual Deployment Tasks

- Required section to document manual steps needed before or during
  production deployment (manual configuration, third-party service setup,
  API key acquisition, etc).
- If not applicable, write exactly: `None`.

## Stop Condition

**IMPORTANT**: After writing/updating if `<ImplementationPlanName>` already reflects the current gaps, reply with:
`<COMPLETION_SIGNAL>`
```

## Related Specifications

- [../prompts.md](../prompts.md) — resolution precedence and completion signal rules.
- [build-classic.md](build-classic.md) and [build-subagents.md](build-subagents.md) — the build prompts that execute the plan.
- [../commands/prompts.md](../commands/prompts.md) — listing and viewing built-in prompts.
- [../configuration.md](../configuration.md) — substituted config fields.
