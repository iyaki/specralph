---
name: agent-implementation
description: Implement or update support for an underlying agent CLI in Ralph after specs are defined, including factory wiring and tests.
---

# Agent Implementation Workflow

## Purpose

Implement or update agent CLI integrations in Ralph in a consistent, deterministic, test-first way.

## When to Use

Use this skill when the task involves any of the following:

- Adding a new agent integration under `internal/agent/`.
- Updating command construction, flags, env vars, or output parsing for an existing agent.
- Updating agent factory selection in `internal/agent/agent.go`.
- Adding or fixing tests around agent selection/execution behavior.

## Required Workflow

1. Read intent and constraints first:
   - `specs/README.md`
   - `specs/agents.md`
   - relevant files in `specs/agents/*.md`
2. Validate target agent CLI behavior before coding:
   - `<agent-cli> --help`
   - relevant subcommand help (for example `<agent-cli> run --help`)
3. Follow TDD strictly:
   - write or update failing tests before implementation
   - verify tests fail for the expected reason
   - implement minimal code to pass
4. Keep implementation aligned with existing patterns:
   - `internal/agent/agent.go`
   - `internal/agent/opencode.go`
   - `internal/agent/claude.go`
   - `internal/agent/cursor.go`
5. Keep behavior deterministic and avoid logging sensitive prompt/match text.

## Constraints

- Do not run iterative Ralph loop automation.
- Keep args/env handling explicit and testable.
- Reuse shared helpers when behavior is equivalent.

## Implementation Checklist

- Add `internal/agent/<new-agent>.go` with an `Agent` interface implementation.
- Wire the new agent into the factory/switch in `internal/agent/agent.go`.
- Register the canonical name in `SupportedAgents()` and any alias in the factory (`internal/agent/agent.go`); `ralph init` options/prompt and error messages derive from it automatically — no separate init list exists.
- Reuse shared execution/helpers where possible instead of duplicating logic.

## Test Checklist

At minimum, add coverage for:

- Factory returns the correct implementation for the agent name.
- `SupportedAgents`/`IsSupportedAgent` cover the new name and aliases.
- Unsupported/invalid names return expected errors.
- Command and argument composition for normal execution.
- Failure paths (missing binary, non-zero exit, malformed output when relevant).

## Validation Commands

Run in this order unless the user requests otherwise:

```bash
go test -v ./internal/agent/...
make test        # full Go suite, includes test/e2e
make test-e2e    # end-to-end suite only
make quality
```

Run mutation testing only in final validation stages:

```bash
make mutation
```

## Completion Criteria

- Tests cover factory mapping plus success and error execution paths.
- Local validation passes (or failures are clearly reported).
- Specs/docs are updated when behavior changes, including the README "Supported Agents" table and `specs/agents.md` in the same change.
