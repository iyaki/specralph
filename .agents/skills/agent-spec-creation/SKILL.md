---
name: agent-spec-creation
description: Create or update agent integration specs before implementation. Use this whenever a request is about defining how a new or existing agent should behave in Ralph.
---

# Agent Spec Creation Workflow

## Purpose

Create implementation-ready agent integration specifications without changing production code.

## When to Use

Use this skill when the task is to define or revise agent behavior, inputs, outputs, error handling, or acceptance criteria in specs.

## Required Workflow

1. Read intent and existing spec context first:
   - `specs/README.md`
   - `specs/agents.md`
   - relevant files in `specs/agents/*.md`
2. Rely on the `spec-creator` skill for document structure and quality requirements.
3. Study the target CLI behavior before writing spec details:
   - `<agent-cli> --help`
   - relevant subcommand help (for example `<agent-cli> run --help`)
4. Define explicit, testable requirements:
   - canonical config name and accepted aliases: document one canonical name, list aliases explicitly (for example `antigravity` with alias `agy`)
   - command shape and required flags
   - environment variables
   - output parsing/streaming expectations
   - non-interactive execution behavior
   - failure modes and expected surfaced errors
5. Update spec files only.

## Constraints

- Do not implement code in `internal/agent/*` while using this skill.
- Do not run iterative Ralph loop automation.
- Keep the workflow linear: read, specify, and stop after spec updates.

## Completion Criteria

- Spec changes are clear, deterministic, and testable.
- Scope and non-goals are explicit.
- Related index files (for example `specs/README.md`) are updated when new spec files are added.
