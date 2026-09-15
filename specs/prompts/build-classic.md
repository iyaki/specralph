# Build Classic Prompt

Status: Implemented

## Overview

### Purpose

- Define the instruction-level content contract of the built-in `build-classic` prompt, the legacy single-task build prompt.
- Serve as the backwards-compatibility reference: `build` resolves here by default (alias rules in [build-subagents.md](build-subagents.md)).

### Goals

- Make generated prompt text reviewable and testable without reading generator code.
- Specify the config substitutions the generator performs.

### Non-Goals

- Prompt resolution order and completion signal rules (see [../prompts.md](../prompts.md)).
- Batch/subagent execution (see [build-subagents.md](build-subagents.md)).
- Custom prompt authoring (see [../commands/prompts-authoring.md](../commands/prompts-authoring.md)).

### Scope

- In scope: generated instruction content and config substitutions.
- Out of scope: loop execution, agent adapters, alias resolution.

## Architecture

### Module/package layout (tree format)

```
internal/
  prompt/
    prompts.go    (BuildPrompt)
```

### Data flow summary

1. Prompt name `build-classic` reaches the bundled generator (directly, or via the `build` alias when configured); a banner naming the resolved built-in is written to output.
2. Config values are substituted into fixed instruction lines.
3. The `<COMPLETION_SIGNAL>` placeholder is replaced by the loop at runtime; gating rules live in [../prompts.md](../prompts.md).

## Data model

### Substitutions

| Placeholder | Source | Notes |
| ----------- | ------ | ----- |
| `<SpecsDir>` | `SpecsDir` | |
| `<INDEX_REFERENCE>` | `SpecsDir` + `SpecsIndexFile` | ` (including `<SpecsDir>/<SpecsIndexFile>` and related specs)` when `SpecsIndexFile` is set and `NoSpecsIndex` is false; empty otherwise |
| `<ImplementationPlanName>` | `ImplementationPlanName` | |

## Workflows

### Generate `build-classic`

1. Prompt name `build-classic` falls through resolution to the bundled generator.
2. The generator emits the instruction contract in the Appendix with substitutions applied.

## APIs

- `internal/prompt.BuildPrompt(cfg *config.Config) string`

## Client SDK Design

- Not applicable.

## Configuration

- Uses `SpecsDir`, `SpecsIndexFile`, `NoSpecsIndex`, and `ImplementationPlanName`; definitions in [../configuration.md](../configuration.md).

## Permissions

- Read access to specs and implementation plan files is required by the instructed workflow, not by generation itself.

## Security Considerations

- Generated text embeds configuration paths; treat prompt text and logs as sensitive (see [../prompts.md](../prompts.md)).

## Dependencies

- Standard library only.

## Open Questions / Risks

- None.

## Verifications

- `ralph prompts show build-classic` starts with `# Agent Instructions (Build Mode)` and contains: "pick the single most important task", the do-not-start-another-task stop condition, the verification log format `YYYY-MM-DD: <command or URL> - <result>`, the `Manual Deployment Tasks` requirement, and `<COMPLETION_SIGNAL>`.
- With `--no-specs-index`, the specs index parenthetical is omitted.
- `ralph build-classic` output equals the former `build` prompt text.

## Appendices

### Instruction contract

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

(The `Manual Deployment Tasks` line's awkward wrapping is preserved verbatim from the generator.)

## Related Specifications

- [../prompts.md](../prompts.md) — resolution precedence and completion signal rules.
- [build-subagents.md](build-subagents.md) — the batch prompt and the `build` alias.
- [../commands/prompts.md](../commands/prompts.md) — listing and viewing built-in prompts.
- [../configuration.md](../configuration.md) — substituted config fields.
