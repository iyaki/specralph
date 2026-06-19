# Run Command

Status: Implemented

## Overview

### Purpose

- Define an explicit `run` command for loop execution while preserving existing prompt-driven usage.
- Remove ambiguity between subcommand names and prompt names.

### Goals

- Make `run` the canonical loop entrypoint.
- Keep `ralph` (no args) behavior equivalent to `ralph run build`.
- Preserve compatibility for `ralph <prompt> [scope]` when the prompt name is not a subcommand.
- Make collision behavior deterministic for current and future subcommands.

### Non-Goals

- Changing prompt source precedence or prompt content generation.
- Changing configuration precedence rules.
- Redesigning `init` workflow behavior.

### Scope

- In scope: command routing, default dispatch behavior, and command/prompt collision policy.
- Out of scope: agent internals, prompt templates, and config schema additions.

## Architecture

### Module/package layout (tree format)

```
cmd/
  ralph/
    main.go
internal/
  cli/
    cmd.go
    run.go
    init.go
specs/
- commands/run.md
```

### Component diagram (ASCII)

```
+------------------------------+
| User Input: ralph <args>     |
+---------------+--------------+
                |
                v
+---------------+--------------+
| Root Command Router          |
| - subcommands win            |
| - fallback to run/default    |
+-------+----------------------+
        |                      |
        v                      v
+-------+---------+   +--------+--------+
| Explicit Command |   | Run Loop Entry  |
| (e.g., init)     |   | (run / fallback)|
+------------------+   +-----------------+
```

### Data flow summary

1. User invokes `ralph` with args.
2. Router checks for explicit registered subcommand match first.
3. If subcommand matches, dispatch to that subcommand.
4. Otherwise dispatch to run behavior (`run [prompt] [scope]`).
5. If no args are provided, dispatch as `run build`.

## Data model

### Core Entities

- CommandDispatchRequest
  - Fields: `Args`, `KnownSubcommands`, `GlobalFlags`.
  - Represents one CLI invocation before dispatch.

- RunInvocation
  - Fields: `PromptName`, `Scope`, `IsAlias`.
  - Represents normalized input to loop execution.

### Relationships

- `CommandDispatchRequest` resolves to either a subcommand execution path or `RunInvocation`.
- `RunInvocation` feeds existing prompt resolution and loop execution without changing prompt precedence.

### Persistence Notes

- None. Dispatch is runtime-only.

## Workflows

### Default invocation (`ralph`)

1. User runs `ralph` with no positional args.
2. Router dispatches to run behavior with defaults.
3. Effective behavior is equivalent to `ralph run build`.

### Explicit subcommand (`ralph init`)

1. User runs `ralph init`.
2. Router matches `init` as a registered subcommand.
3. CLI executes init flow, not prompt execution.

### Explicit run (`ralph run init`)

1. User runs `ralph run init`.
2. Router dispatches to run command.
3. Run command executes loop using prompt name `init`.

### Prompt alias (`ralph <prompt> [scope]`)

1. User runs `ralph <prompt> [scope]` where `<prompt>` is not a registered subcommand.
2. Router normalizes invocation to `ralph run <prompt> [scope]`.
3. Loop behavior is identical to explicit run invocation.

### Name collision policy

1. If first token matches a registered subcommand, subcommand dispatch always wins.
2. Prompt names that collide with subcommand names must be invoked through `ralph run <prompt>`.
3. New subcommands are reserved words for command dispatch by default.

## APIs

- None. This is local CLI routing behavior.

## Client SDK Design

- Not applicable.

## Configuration

- Configuration precedence remains unchanged: flags > env vars > config file > defaults.
- Prompt resolution precedence remains unchanged and is defined in [specs/prompts.md](prompts.md).

### Command interface

- `ralph`
- `ralph run [prompt] [scope]`
- `ralph <subcommand> ...`
- `ralph <prompt> [scope]` (alias; only when `<prompt>` is not a subcommand)

### Defaults

| Invocation   | Normalized behavior |
| ------------ | ------------------- |
| `ralph`      | `ralph run build`   |
| `ralph run`  | `ralph run build`   |
| `ralph init` | `ralph init`        |

## Permissions

- Same as existing CLI execution and prompt loading paths.

## Security Considerations

- No new secrets surface is introduced.
- Deterministic command routing reduces accidental execution of unintended prompt names.

## Dependencies

| Dependency               | Purpose                    |
| ------------------------ | -------------------------- |
| `github.com/spf13/cobra` | Command parsing/dispatch   |
| Existing `internal/*`    | Prompt resolution and loop |

## Verifications

- `ralph` behaves equivalently to `ralph run build`.
- `ralph init` executes the init subcommand.
- `ralph run init` executes the loop with prompt name `init`.
- `ralph build` remains valid as an alias to `ralph run build`.
- Collision rule holds for all registered subcommands.

## Appendices

- None.
