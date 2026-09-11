---
name: specralph-prompts
description: Author, install, and validate custom prompt files for specralph (the ralph CLI agentic loop runner). Use this skill whenever creating or editing ralph prompt files, preparing tasks to run with ralph run, writing prompts that must stop the loop via the completion signal, or debugging a ralph loop that never stops or stops too early.
---

# Specralph Prompt Authoring

Ralph sends a prompt to an AI agent and loops until the agent's output contains the completion signal. A prompt file is plain Markdown living in the configured prompts directory; its filename (without `.md`) is how it is invoked. Getting the file right means the loop does exactly one task and stops cleanly.

## The one rule that matters

Every prompt must tell the agent to output the completion signal when the work is done. Without it, the loop only stops at max iterations — wasted runs and half-finished work. Either form works:

- `<COMPLETION_SIGNAL>` — ralph replaces it at runtime with the real tag
- `<promise>COMPLETE</promise>` — the literal tag, case-sensitive

Put it in the prompt's stop condition, for example:

```markdown
## Stop Condition

- After completing the selected task, stop. Do NOT start another task.
- If everything is complete, reply with: `<COMPLETION_SIGNAL>`
```

## Workflow

1. Study the conventions from the binary itself: `ralph prompts list` shows available prompts; `ralph prompts show build` (or `plan`) shows the built-in structure worth imitating. Newer ralph versions also provide `ralph prompts guide` (the full authoring contract) and `ralph prompts validate <name>` (static checks on a prompt file) — prefer them when available.
2. Write the prompt as `<name>.md` in the prompts directory (config field `prompts-dir`, default `$HOME/.ralph`). An optional YAML frontmatter `description:` field gives `ralph prompts list` a useful summary.
3. Structure it like the built-ins: objective, study inputs (specs, implementation plan), single-task steps, validation, stop condition with the signal. One task per run — ralph loops, the agent picks one task, validates, commits.
4. Run it: `ralph run <name>`, or the alias `ralph <name>` when `<name>` is not a registered subcommand. Ralph prints the resolved prompt, then enters the loop.
5. Verify the loop stops: the agent's output must contain the literal signal. A run that never stops means the prompt is missing the signal — fix the prompt file, not the config.

## Notes

- The `scope` argument (`ralph plan my-feature`) only affects built-in prompts today; custom prompts must inline whatever context they need.
- Distribute this skill into a project with `ralph skill install [dir]` (default `.agents/skills`) so agents working there discover these conventions; the installed skill always matches the ralph binary version that installs it.
