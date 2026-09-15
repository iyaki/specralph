# Prompts

Status: Implemented

## Overview

### Purpose

- Define how Ralph resolves prompts (including the `build` alias) and generates default build/plan prompts.
- Provide a testable description of prompt precedence and prompt content inputs.

### Goals

- Specify prompt resolution order and failure behavior.
- Document the built-in build (`build-classic`, `build-subagents`) and plan prompts at a behavioral level.
- Describe how prompt files are discovered on disk.

### Non-Goals

- Defining new prompt types or templates.
- Editing prompt content beyond what is implemented in code.
- Prompt discovery or listing CLI behavior (see [commands/prompts.md](commands/prompts.md)).

### Scope

- In scope: prompt sources, resolution order, default prompt generation.
- Out of scope: agent execution, config precedence (see configuration spec), CLI command UI.

## Architecture

### Module/package layout (tree format)

```
internal/
  prompt/
    prompts.go
```

### Component diagram (ASCII)

```
+------------------+
| Prompt Resolver  |
| internal/prompt  |
+---------+--------+
          |
          v
+---------+--------+        +------------------+
| Prompt Content   |<------>| Inline/StdIn     |
+---------+--------+        +------------------+
          |                 +------------------+
          +---------------->| Prompt File(s)   |
          |                 +------------------+
          |                 +------------------+
          +---------------->| Built-in Prompts |
                            +------------------+
```

### Data flow summary

1. Prompt resolver checks for inline prompt text.
2. If not inline, it checks stdin usage.
3. If not stdin, it checks explicit prompt file path.
4. If not explicit, it searches for a prompt file in the prompts directory (walking upward).
5. If not found, it falls back to built-in prompts for `build-classic`, `build-subagents`, and `plan` (the `build` alias rewrites the name first; see [prompts/build-subagents.md](prompts/build-subagents.md)).
6. If no source is valid, it returns an error.

Note: prompt resolution behavior is independent of command routing. Routing and collision rules are defined in [commands/run.md](commands/run.md).

## Data model

### Core Entities

- PromptSource
  - Enum-like: `Inline`, `Stdin`, `File`, `BuiltIn`.
  - Determines how prompt text is obtained.

- PromptRequest
  - Inputs: `promptName`, `scope`, and config fields `CustomPrompt`, `PromptFile`, `PromptsDir`.
  - Output: prompt text string.

### Relationships

- `PromptRequest` is derived from CLI args and config (see configuration spec for fields).
- `PromptSource` is selected by precedence order.

### Persistence Notes

- Prompt files are plain-text Markdown files on disk.

## Workflows

### Resolve prompt (happy path)

1. If `promptName` is `build`, rewrite it to the alias target (`BuildAliasPrompt`, default `build-classic`); if the target is `build`, keep it.
2. If `CustomPrompt` is set, return it.
3. If `PromptFile` is `-` or `promptName` is `-`, read prompt text from stdin.
4. If `PromptFile` is set, read that file.
5. If a prompt file exists at `PromptsDir/<promptName>.md` (searching upward), read it.
6. If `promptName` is `build-classic`, `build-subagents`, or `plan`, generate the built-in prompt.
7. Otherwise, return an error.

### Resolve prompt through explicit run command

1. User invokes `ralph run [prompt] [scope]`.
2. Run command determines `promptName` and optional `scope`.
3. Prompt resolver applies the same precedence chain as the generic happy path.
4. Resulting prompt text is passed to loop execution.

### Resolve prompt through alias

1. User invokes `ralph <prompt> [scope]`.
2. Command router treats this as an alias to `ralph run <prompt> [scope]` when `<prompt>` is not a registered subcommand.
3. Prompt resolver behavior is identical to explicit run invocation.

### Resolve prompt when command name collides with prompt name

1. User invokes `ralph <name>` where `<name>` is both a subcommand and a prompt file name.
2. Command router dispatches to the subcommand.
3. To execute the prompt, user invokes `ralph run <name>`.

### Resolve prompt (missing file)

1. Prompt file path is provided but cannot be read.
2. Resolver returns an error including the path.

### Resolve prompt (unknown name)

1. `promptName` does not match a prompt file or built-in prompt.
2. Resolver returns an error: prompt not found.

## APIs

- None. Prompts are resolved locally.

## Client SDK Design

- Not applicable.

## Configuration

- Prompt resolution uses config fields: `CustomPrompt`, `PromptFile`, `PromptsDir`, `SpecsDir`, `SpecsIndexFile`, `ImplementationPlanName`.
- See configuration spec for full definitions and precedence.

## Permissions

- Requires read access to prompt files and stdin.

## Security Considerations

- Prompt text may include sensitive data; avoid committing secrets to prompt files.
- Prompt content is logged to stdout and optionally to a log file; treat logs as sensitive.

## Dependencies

- Standard library only (`os`, `io`, `path/filepath`).

## Open Questions / Risks

- Should prompt discovery search the current directory before `PromptsDir/<name>.md`?
- Should prompt file lookup be strict to prevent parent directory traversal?

## Verifications

- `ralph --prompt "hello" build` uses inline prompt.
- `echo "hi" | ralph -` reads from stdin.
- `ralph --prompt-file ./prompts/build.md build` uses that file.
- `ralph plan` uses built-in plan prompt when no file exists.
- `ralph run plan` uses built-in plan prompt when no file exists.
- `ralph init` executes init subcommand, while `ralph run init` resolves prompt `init`.
- `ralph build` resolves through the `build` alias to `build-classic` by default; `ralph init`-generated configs target `build-subagents` (see [prompts/build-subagents.md](prompts/build-subagents.md)).

## Completion Signal

Ralph detects task completion by searching for the following XML-like tag in agent output:

```xml
<promise>COMPLETE</promise>
```

The built-in `build-classic`, `build-subagents`, and `plan` prompts use the placeholder `<COMPLETION_SIGNAL>`, which Ralph automatically replaces with `<promise>COMPLETE</promise>` at runtime before sending the prompt to the agent.

When creating custom prompts (inline via `--prompt` or in prompt files), you can either:

1. Use the placeholder `<COMPLETION_SIGNAL>` — Ralph will replace it for you:
   ```markdown
   Implement feature X.
   Write tests for Y.
   When everything is done, output: <COMPLETION_SIGNAL>
   ```

2. Or directly include the actual signal:
   ```markdown
   Implement feature X.
   Write tests for Y.
   When everything is done, output: <promise>COMPLETE</promise>
   ```

The tag is case-sensitive and must appear exactly as shown. Ralph will stop after detecting the first occurrence.

## Appendices

### Built-in prompt behavior (summary)

- `build-classic` (legacy single-task build prompt) — full content contract in [prompts/build-classic.md](prompts/build-classic.md):
  - Instructs to study specs and the implementation plan.
  - Requires implementing a single task, validating, updating plan, and committing.
  - Injects the completion signal `<promise>COMPLETE</promise>` automatically. See [Completion Signal](#completion-signal) for details on custom prompts.

- `build-subagents` (opt-in via `build-alias-prompt`, written by `ralph init`):
  - Uses the same study instructions as `build-classic`.
  - Selects up to 10 pending tasks and dispatches at least one subagent per task; validates, commits, and updates the plan per task.
  - Emits the completion signal only when all plan tasks are complete. See [prompts/build-subagents.md](prompts/build-subagents.md).

- Plan prompt — full content contract in [prompts/plan.md](prompts/plan.md):
  - Instructs to generate/update the implementation plan in a structured format.
  - Requires study/gap analysis against specs and code.
  - Injects the completion signal `<promise>COMPLETE</promise>` automatically. See [Completion Signal](#completion-signal) for details on custom prompts.

Full instruction contracts for the built-in prompts live in [prompts/build-classic.md](prompts/build-classic.md), [prompts/build-subagents.md](prompts/build-subagents.md), and [prompts/plan.md](prompts/plan.md).

## Related Specifications

- [prompts/build-subagents.md](prompts/build-subagents.md) — Build prompt with subagents, the `build` alias, and the `build-alias-prompt` config field.
- [commands/prompts.md](commands/prompts.md) — CLI `prompts` command for listing and viewing prompts. Use `ralph prompts list` to discover available prompts and `ralph prompts show <name>` to inspect full prompt content.
- [commands/run.md](commands/run.md) — Using prompts with the `run` command.
- [configuration.md](configuration.md) — Configuration fields used in prompt resolution.