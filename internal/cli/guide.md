# Ralph Prompt Authoring Guide

How to write a custom prompt file so a ralph loop runs exactly one task and stops cleanly.

## File Location

- Prompt files live in the prompts directory (`prompts-dir` config field, default `$HOME/.ralph`) as `<name>.md`.
- The filename (without `.md`) is the prompt name: run it with `ralph run <name>`, or the alias `ralph <name>` when `<name>` is not a registered subcommand.

## Frontmatter (Optional)

Start the file with YAML frontmatter to give `ralph prompts list` a useful summary:

    ---
    description: One-line summary shown by ralph prompts list
    ---

## Completion Signal (Required)

The loop stops only when the agent's output contains the completion signal. Without it, a run only stops at max iterations. Put one of these forms in the prompt's stop condition:

- `<COMPLETION_SIGNAL>` — ralph replaces it at runtime with the literal tag
- `<promise>COMPLETE</promise>` — the literal tag, case-sensitive

Example stop condition:

    ## Stop Condition

    - After completing the selected task, stop.
    - When everything is done, reply with: `<COMPLETION_SIGNAL>`

## Recommended Structure

Mirror the built-ins (`ralph prompts show build`, `ralph prompts show plan`):

1. Objective — the single most important task for this run.
2. Study Inputs — specs and the implementation plan to read first.
3. Steps — one task per loop iteration: implement, validate, commit, stop.
4. Validation — how to prove the work (tests, commands).
5. Stop Condition — when to reply with the completion signal.

## Scope Argument

`ralph run <prompt> <scope>` passes `<scope>` to built-in prompts only. It is NOT substituted into custom prompts — inline whatever context you need in the file itself.

## Validate Before Running

`ralph prompts validate <name-or-file>` checks frontmatter, description, and the completion signal, reporting every problem in one pass.

## Distribute to a Project

`ralph skill install [dir]` (default `.agents/skills`) installs the prompt-authoring skill so agents working in a project discover these conventions automatically; the installed skill always matches the ralph binary that installs it.
