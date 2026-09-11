# Skill Command

Status: Proposed

## Overview

### Purpose

Distribute the bundled prompt-authoring skill (`specralph-prompts`) into projects, so AI agents working there can discover specralph prompt conventions (file location, structure, completion signal) without external documentation.

### Goals

- Specify the `ralph skill install [dir]` subcommand.
- Specify the skill source: embedded in the binary at build time from repository source files, so an installation always matches the installing binary's version.
- Specify overwrite behavior and output.

### Non-Goals

- Network downloads (GitHub fetches, version pinning, update checks) — the binary is the single source.
- Skill management beyond install: no `list`, `uninstall`, or per-agent format conversion.
- Lock-file integration with external skill managers (`.skill-lock.json` files are not modified).
- Changes to prompt resolution or authoring behavior (see [prompts-authoring.md](prompts-authoring.md)).

### Scope

- In scope: `skill` command group with a single `install` subcommand.
- Out of scope: prompt authoring contract content (owned by the skill file itself).

## Architecture

### Module/package layout (tree format)

```
.agents/
  skills/
    specralph-prompts/
      SKILL.md            (canonical skill source, committed)
internal/
  cli/
    skill.go              (command; embeds the skill at build time)
```

### Data flow summary

1. Build embeds `.agents/skills/specralph-prompts/SKILL.md` into the binary (build-time step; the exact embedding mechanism is an implementation detail — Go's `embed` cannot read dot-directories, so the build may copy or relocate the file into the embedding package).
2. `ralph skill install [dir]` writes the embedded skill to `<dir>/specralph-prompts/SKILL.md`.

## Data model

### Core Entities

- None beyond the skill file itself: one embedded asset, one output path.

## Workflows

### Install the skill (happy path)

1. User invokes `ralph skill install [dir]`.
2. `<dir>` defaults to `.agents/skills` when omitted.
3. Target path is `<dir>/specralph-prompts/SKILL.md`.
4. If the target exists and `--force` is not set, return an error: `skill already exists at <target> (use --force to overwrite)`.
5. Create intermediate directories as needed (`0755`).
6. Write the embedded skill content (`0644`).
7. Print to stdout: `Installed specralph-prompts skill to <target>`. Exit code 0.

### Overwrite an existing installation

1. User invokes `ralph skill install [dir] --force`.
2. Existing target file is overwritten with the embedded content. No backup is kept (the binary is the source of truth; local edits are intentionally discarded).

### Install failure

1. Target directory cannot be created or written (permissions, read-only path).
2. Command returns an error wrapping the cause; exit code non-zero. No partial file is left behind.

## APIs

- None. Local file operation only.

## Configuration

- None. The command is config-independent: it does not load `ralph.toml`, and the prompts-dir setting is unrelated to the install target.

## Command Collision

`skill` is a registered subcommand, so it shadows a prompt file named `skill.md`: invoking `ralph skill` runs the subcommand, and the prompt requires `ralph run skill`. This follows the collision rules defined in [run.md](run.md).

## Permissions

- Write access to the target directory.

## Security Considerations

- The installed content is the static, embedded skill text; no user input is reflected into the file path beyond `<dir>`.
- Path arguments are used as given; no upward-traversal sanitization beyond what `filepath.Join` semantics provide (the user already has filesystem write access to wherever they point the command).

## Dependencies

- Standard library plus the build-time embedding of the skill asset. No new external dependencies.

## Open Questions / Risks

- Should a successful install print a hint about which agents read `.agents/skills` (e.g. "supported by skills-aware agents")? Deferred until an agent-support matrix exists.
- Repos using an external skill manager may prefer registering the skill manually; the lock files of those tools are intentionally untouched.

## Verifications

- `ralph skill install` creates `.agents/skills/specralph-prompts/SKILL.md` in the working directory and prints the install path.
- `ralph skill install /tmp/x` creates `/tmp/x/specralph-prompts/SKILL.md`.
- Installing twice without `--force` fails with `skill already exists at <target> (use --force to overwrite)`, leaving the original file untouched.
- Installing twice with `--force` succeeds and replaces the content with the embedded version.
- The installed file passes `skill-creator` validation (`quick_validate.py`): valid frontmatter, kebab-case name `specralph-prompts`, description without angle brackets and under 1024 characters.
- `ralph skill --help` lists the `install` subcommand.

## Related Specifications

- [prompts-authoring.md](prompts-authoring.md) — the authoring workflow the skill teaches (`prompts guide`, `prompts validate`).
- [run.md](run.md) — command/prompt name collision rules.
- [help.md](help.md) — command discoverability.
