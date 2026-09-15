# Specralph Specifications

Design docs and technical specifications

## Vision and Goals

- [vision-goals.md](vision-goals.md)

## Core Architecture

- [core-architecture.md](core-architecture.md)

## Technical details

- [development-testing.md](development-testing.md)
- [e2e-testing.md](e2e-testing.md)
- [release-workflow.md](release-workflow.md)

## Behaviours

- [agent-env-overrides.md](agent-env-overrides.md)
- [agents.md](agents.md)
- [build-subagents.md](build-subagents.md) — Build prompt with subagents (up to 10 tasks, at least one subagent each) and the configurable `build` alias
- [built-in-prompts.md](built-in-prompts.md) — Content contracts for the built-in `build-classic` and `plan` prompts
- [config-by-prompt.md](config-by-prompt.md)
- [config-local-overlay.md](config-local-overlay.md)
- [configuration.md](configuration.md)
- [logging.md](logging.md)
- [prompts.md](prompts.md)

### CLI Commands

- [commands/help.md](commands/help.md) — Get help about commands
- [commands/init.md](commands/init.md) — Initialize Ralph configuration
- [commands/prompts.md](commands/prompts.md) — List and view prompts
- [commands/prompts-authoring.md](commands/prompts-authoring.md) — Guide and validation for authoring prompts
- [commands/run.md](commands/run.md) — Run prompt loop
- [commands/skill.md](commands/skill.md) — Install the bundled prompt-authoring skill
- [commands/version.md](commands/version.md) — Show version info

### Agents Implementations

- [agents/opencode.md](agents/opencode.md)
- [agents/claude.md](agents/claude.md)
- [agents/cursor.md](agents/cursor.md)
- [agents/oh-my-pi.md](agents/oh-my-pi.md)
- [agents/codex.md](agents/codex.md) — OpenAI Codex CLI integration
- [agents/copilot.md](agents/copilot.md) — GitHub Copilot CLI integration
- [agents/antigravity.md](agents/antigravity.md) — Google Antigravity CLI integration