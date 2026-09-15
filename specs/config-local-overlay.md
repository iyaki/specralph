# Config Local Overlay

Status: Implemented

## Overview

### Purpose

- Define support for an optional `ralph-local.toml` file that overlays a selected base config file.
- Enable team-shared baseline configuration in version control with user-specific local overrides outside version control.

### Goals

- Preserve deterministic config resolution.
- Define exact lookup rules for `ralph-local.toml` in all base-config selection modes.
- Define merge semantics for scalar values, tables, arrays, and `prompt-overrides`.
- Keep existing high-priority sources (`flags`, `env`) unchanged.

### Non-Goals

- Adding new config keys.
- Changing prompt front matter behavior.
- Implementing automatic `.gitignore` edits.

### Scope

- In scope: base-config selection, local overlay discovery, merge semantics, precedence, and failure modes.
- Out of scope: config discovery in parent directories.

## Architecture

### Component diagram (ASCII)

```
+------------------+
|  CLI flags       |
+---------+--------+
          |
          v
+---------+--------+        +------------------+
| Config Resolver  |<------>| Env vars         |
+----+--------+----+        +------------------+
     |        |
     |        +------------------------------+
     |                                       |
     v                                       v
+----+-----------------+         +-----------+-----------+
| Base config file     |         | Local overlay file    |
| (selected once)      |         | ralph-local.toml      |
+----+-----------------+         +-----------+-----------+
     |                                       |
     +------------------- merge -------------+
                         (local over base)
                                |
                                v
                      +---------+---------+
                      | Effective file cfg |
                      +---------+---------+
                                |
                                v
                      +---------+---------+
                      | Final resolved cfg |
                      +--------------------+
```

## Data model

### Core entities

- Base config: TOML file selected by existing rules (`--config`, `RALPH_CONFIG`, default filenames).
- Local overlay: optional `ralph-local.toml` file in the same directory as the selected base config.
- Merged file config: deterministic merge result where local values override base values.

### Merge semantics

- Scalar keys (`string`, `int`, `bool`): local overrides base when key exists in local.
- Tables/maps: deep merge by key; keys absent in local remain from base.
- Arrays/lists: full replacement when key exists in local.
- `prompt-overrides`: deep merge by prompt name, then by prompt fields.

Example:

```toml
# base: ralph.toml
model = "gpt-5"

[prompt-overrides.build-subagents]
model = "gpt-5.3-codex"

# local: ralph-local.toml
[prompt-overrides.build-subagents]
agent-mode = "planner"
```

Expected merged result for `prompt-overrides.build-subagents`:

- `model = "gpt-5.3-codex"` (from base)
- `agent-mode = "planner"` (from local)

## Workflows

### Base config selection

1. If `--config` is set, use that file as base config.
2. Else if `RALPH_CONFIG` is set, use that file as base config.
3. Else discover `ralph.toml` in the current directory.

### Local overlay selection

1. Compute `dir = dirname(<selected base config>)`.
2. Look for `dir/ralph-local.toml`.
3. If local file exists, parse and merge it over base config.
4. If local file does not exist, continue with base config only.

### Error handling

1. Base config parse/read error: fail fast with file path.
2. Local overlay parse/read error: fail fast with file path.
3. Missing local overlay: no error.

## Configuration

### Effective precedence

- General precedence: `flags > env vars > merged-config(local over base) > defaults`.

### `model` and `agent-mode` precedence

1. CLI flags (`--model`, `--agent-mode`)
2. Environment variables (`RALPH_MODEL`, `RALPH_AGENT_MODE`)
3. Prompt front matter
4. Prompt-specific merged config override (`[prompt-overrides.<prompt>]` from merged config)
5. Global merged config keys (`model`, `agent-mode`)
6. Defaults

### Team workflow guidance

- Commit baseline config (for example `ralph.toml`) to the repository.
- Keep `ralph-local.toml` untracked in consumer repositories.
- Consumer repos that adopt Ralph should add `ralph-local.toml` to their `.gitignore`.

## Security Considerations

- Treat both base and local config as potentially sensitive.
- Error messages should identify file paths but should not print secret values.
- Merge behavior must be deterministic and reproducible.

## Open Questions / Risks

- Whether to support additional local overlay names in the future.
- Whether invalid local overlay should be optionally downgraded to warning in a non-strict mode.

## Verifications

- `--config ./team/ralph.toml` reads local overlay from `./team/ralph-local.toml`.
- `RALPH_CONFIG=./team/ralph.toml` reads local overlay from `./team/ralph-local.toml`, not from current working directory.
- Default-discovered `ralph.toml` uses sibling `ralph-local.toml` when present.
- Missing `ralph-local.toml` keeps existing behavior unchanged.
- Invalid `ralph-local.toml` fails before agent execution starts.
- Conflicts resolve as `flags > env > merged file config > defaults`.
- `prompt-overrides` keys merge deeply between base and local config.
- Front matter retains precedence over merged `prompt-overrides`.
