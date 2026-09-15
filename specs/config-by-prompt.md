# Config by Prompt

Status: Implemented

## Overview

### Purpose

- Define how a prompt markdown file can override agent runtime settings using YAML front matter.
- Define how a config file can set per-prompt `model` and `agent-mode` overrides in a dedicated section.
- Ensure front matter metadata is stripped before prompt text is sent to the external agent CLI.

### Goals

- Support per-prompt overrides for `model` and `agent-mode`.
- Support per-prompt overrides from both front matter and config file.
- Keep existing global configuration behavior intact.
- Define deterministic precedence across flags, env vars, front matter, config file, and defaults.
- Specify clear behavior for invalid front matter and unknown keys.

### Non-Goals

- Introducing new prompt source types.
- Adding support for overriding `agent` in front matter.
- Defining implementation-specific parser internals.

### Scope

- In scope: markdown prompt files loaded via `--prompt-file` and `PromptsDir/<prompt>.md`.
- Out of scope: inline prompts (`--prompt`) and stdin prompt text (`-`).

## Architecture

### Module/package layout (tree format)

```
internal/
  cli/
    cmd.go
  config/
    config.go
  prompt/
    prompts.go
```

### Component diagram (ASCII)

```
+-------------------------+
| Prompt Source Resolver  |
+------------+------------+
             |
             v
+------------+------------+
| Front Matter Extractor  |
| (file-based markdown)   |
+------+------------------+
       | metadata
       v
+------+------------------+      +-------------------------+
| Effective Config Merge  |<-----| flags/env/file defaults |
+------+------------------+      +-------------------------+
       |
       v
+------+------------------+
| Sanitized Prompt Body   |
+------+------------------+
       |
       v
+------+------------------+
| Agent Execute(prompt)   |
+-------------------------+
```

### Data flow summary

1. Prompt resolver reads prompt text from a file-based markdown source.
2. Resolver checks whether the file starts with YAML front matter.
3. Supported keys (`model`, `agent-mode`) are extracted as prompt-level overrides.
4. Front matter block is removed from prompt text.
5. Effective runtime settings are resolved with precedence rules.
6. Only sanitized prompt body is passed to `Agent.Execute(...)`.

## Data model

### Core Entities

- PromptFrontMatterSettings
  - `Model` (optional string)
  - `AgentMode` (optional string)

- PromptConfigOverride
  - `Model` (optional string)
  - `AgentMode` (optional string)

- PromptConfigOverridesMap
  - Key: prompt name (for example `build-subagents`, `plan`, `my-prompt`)
  - Value: `PromptConfigOverride`

- EffectiveAgentSettings
  - Final `Model` and `AgentMode` values used by agent factory/execution.

### Relationships

- Prompt front matter participates in configuration resolution only for file-based markdown prompts.
- Prompt config overrides participate in configuration resolution when prompt name is known.
- Effective values are computed once per run before each agent invocation cycle.

### Persistence Notes

- Front matter is persisted only inside prompt markdown files.
- No new standalone persistence layer is introduced.

## Workflows

### File prompt with valid front matter (happy path)

1. User executes with prompt file or named prompt from `PromptsDir`.
2. Front matter is parsed from leading YAML block.
3. `model` and/or `agent-mode` are merged into effective runtime settings.
4. YAML block is removed.
5. Sanitized markdown body is sent to agent.

### File prompt without front matter

1. Resolver reads markdown file.
2. No front matter block is detected.
3. Effective runtime settings come from existing sources only.
4. Prompt body is sent unchanged.

### Per-prompt override from config file

1. User runs `ralph <promptName>`.
2. Resolver looks for a matching entry under config prompt override section.
3. Matching `model` and `agent-mode` become prompt-level candidates.
4. Front matter, if present, is still evaluated and can override this source.

### Invalid front matter YAML

1. Resolver detects front matter delimiters but YAML is malformed.
2. Execution fails fast with a clear parse error.
3. Agent execution does not start.

### Unknown front matter keys

1. Resolver parses front matter successfully.
2. Unknown keys are ignored for runtime-setting merge.
3. Supported keys still apply.

## APIs

- None. Behavior is internal to prompt/config resolution.

## Client SDK Design

- Not applicable.

## Configuration

### Config file section for per-prompt overrides

Config files may define a dedicated map-like section named `prompt-overrides`.

Example (`ralph.toml`):

```toml
[prompt-overrides.build-subagents]
model = "gpt-5.3-codex"
agent-mode = "planner"

[prompt-overrides.plan]
model = "claude-sonnet-4"
agent-mode = "reviewer"
```

Notes:

- Each subsection key is the prompt name used at invocation time.
- Values in this section apply only when that prompt is selected.
- This section does not replace existing global `model` and `agent-mode` keys; it complements them.

- Keys refer to the resolved prompt name: the `build` alias resolves to its target before overrides are looked up (see [prompts/build-subagents.md](prompts/build-subagents.md)).

### Supported front matter keys

| Key          | Type   | Meaning                               |
| ------------ | ------ | ------------------------------------- |
| `model`      | string | Per-prompt model override             |
| `agent-mode` | string | Per-prompt agent mode/sub-agent value |

### Source precedence for `model` and `agent-mode`

1. CLI flags (`--model`, `--agent-mode`)
2. Environment variables (`RALPH_MODEL`, `RALPH_AGENT_MODE`)
3. Prompt file front matter
4. Prompt-specific config override (`[prompt-overrides.<promptName>]`)
5. Global config keys (`model`, `agent-mode`)
6. Defaults

## Permissions

- Requires read permission to prompt markdown files.

## Security Considerations

- Front matter may include sensitive metadata; avoid logging raw front matter content.
- Prompt metadata must not be forwarded to the external agent as prompt text.
- Parsing must be deterministic and non-executable.

## Dependencies

- A YAML parser dependency may be required for robust front matter parsing.
- Dependency selection must preserve deterministic behavior and minimal attack surface.

## Open Questions / Risks

- How to disambiguate markdown files that intentionally start with `---` but are not front matter.
- Whether unknown keys should also emit a non-fatal warning.
- Whether to add optional strict mode for key validation in a future iteration.

## Verifications

- `ralph --prompt-file ./prompts/build.md build` applies `model` and `agent-mode` from front matter when present.
- `ralph --model my-cli-model --prompt-file ./prompts/build.md build` keeps CLI flag value over front matter.
- `ralph --config ./ralph.toml build` applies `[prompt-overrides.build-subagents]` values when no higher-precedence source is set.
- `ralph --config ./ralph.toml --prompt-file ./prompts/build.md build` uses front matter over `[prompt-overrides.build-subagents]`.
- A malformed front matter block returns an error before agent execution.
- Debug or trace output confirms prompt body sent to agent excludes front matter block.

## Appendices

### Example prompt file

```markdown
---
model: gpt-5.3-codex
agent-mode: planner
---

# Task

Implement the requested feature using the repository specs.
```

Expected behavior:

- Effective model: `gpt-5.3-codex` unless overridden by flag or env var.
- Effective agent mode: `planner` unless overridden by flag or env var.
- Prompt sent to agent starts at `# Task` (front matter removed).

### Example config file (per-prompt section)

```toml
model = "gpt-4"
agent-mode = "default"

[prompt-overrides.build-subagents]
model = "gpt-5.3-codex"
agent-mode = "planner"
```

Expected behavior:

- Running `ralph build-subagents` — or `ralph build` when the alias targets it (for example in `ralph init`-generated configs) — uses `gpt-5.3-codex` and `planner` unless overridden by flag/env/front matter.
- Running `ralph plan` falls back to global `model = "gpt-4"` and `agent-mode = "default"` unless it has its own prompt override.
