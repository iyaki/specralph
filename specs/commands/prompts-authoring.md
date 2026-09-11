# Prompts Authoring

Status: Proposed

## Overview

### Purpose

Make it practical for AI agents (and humans) to author custom prompt files for specralph using only the installed binary:

- Expose the prompt authoring contract (file conventions, completion signal, prompt structure) via the CLI.
- Provide a validation feedback loop for newly written prompt files.

### Goals

- Specify the `prompts guide` subcommand: embedded, offline authoring documentation.
- Specify the `prompts validate` subcommand: static checks on a prompt file or named prompt.
- Specify a run-time warning when an executed prompt file lacks a completion signal.

### Non-Goals

- Template variable substitution in custom prompts (e.g. plan-file name placeholders).
- A `prompts new` scaffolding command.
- JSON output for `prompts list` or `validate`.
- Changes to prompt resolution precedence (see [prompts.md](../prompts.md)).

### Scope

- In scope: `prompts guide`, `prompts validate`, missing-signal warning in the run flow.
- Out of scope: prompt resolution order, agent execution, loop completion detection (see [run.md](run.md)).

## Architecture

### Module/package layout (tree format)

```
internal/
  cli/
    prompts.go        (guide + validate subcommands)
    run.go            (missing-signal warning after GetPrompt)
  prompt/
    prompts.go        (shared check helpers: HasCompletionSignal moved/exported from cli, description extraction)
```

### Data flow summary

1. `prompts guide` prints a static authoring contract to stdout. No config, no disk access.
2. `prompts validate <target>` resolves the target (file path or prompt name), reads the content, and applies static checks.
3. `ralph run` / `ralph <name>`: after prompt resolution, if the source was a prompt file and the text contains neither the placeholder nor the literal signal, print a one-line warning to stderr.

## Data model

### Core Entities

- `ValidationResult`
  - `Path string` — file the result refers to (resolved path for named prompts).
  - `Checks []Check` — ordered check outcomes.
- `Check`
  - `Name string` — `resolve`, `frontmatter`, `description`, `completion-signal`.
  - `Status string` — `ok`, `warn`, `fail`.
  - `Message string` — human-readable detail (empty when ok).

## Workflows

### Show authoring guide

1. User invokes `ralph prompts guide`.
2. Command prints the authoring contract to stdout:
   - Prompt file location: `PromptsDir` (default `$HOME/.ralph`), `<name>.md`, invoked as `ralph run <name>` or `ralph <name>`.
   - Frontmatter `description` convention (optional; used by `prompts list`).
   - Completion signal contract: include `<COMPLETION_SIGNAL>` (replaced at run time with `<promise>COMPLETE</promise>`) or the literal tag.
   - Recommended structure, mirroring the built-ins: objective → study inputs → single-task steps → validation → stop condition with completion signal.
   - Mention of `scope`: second argument of `ralph run <prompt> <scope>`; custom prompts receive it only if they reference it themselves (today: not substituted — state this explicitly).
   - Distribution: one line pointing at `ralph skill install [dir]` (default `.agents/skills`) so agents working in a project discover these conventions automatically.
3. Exit code 0. Output is static text; pipe-friendly, no banners.

### Validate a prompt

1. User invokes `ralph prompts validate <target>`.
2. Target resolution:
   - If `<target>` ends in `.md` or contains a path separator: treat as file path (must exist).
   - Otherwise: resolve as prompt name via the same chain as `prompts show` (built-in first, then custom files in `PromptsDir`).
3. Checks, in order; first `fail` does not prevent later checks from running (report all):
   - `resolve` — content obtainable; `fail` on missing file / unknown name.
   - `frontmatter` — if content starts with `---`, the closing `---` delimiter must exist; `warn` on unbalanced delimiters. If frontmatter contains `description:`, it must be non-empty; `warn` otherwise.
   - `description` — a description is extractable (frontmatter `description` or first non-empty non-heading body line); `warn` if none.
   - `completion-signal` — content contains `<COMPLETION_SIGNAL>` or `<promise>COMPLETE</promise>`; `fail` if neither (a loop executed with this prompt can never complete).
4. Output to stdout, one line per check, then a summary:

   ```
   prompts/review.md
   ok       resolve
   ok       frontmatter
   warn     description: no description line found
   fail     completion-signal: neither <COMPLETION_SIGNAL> nor <promise>COMPLETE</promise> found
   1 failed, 1 warning
   ```

5. Exit code: 0 if no `fail`, 1 otherwise. Warnings do not affect the exit code.
6. Unknown target: error `prompt "<target>" not found`, exit 1.

### Warn at run time on missing signal

1. `run` resolves the prompt via `GetPrompt` (existing flow).
2. If the resolved source is a prompt file (`PromptFile` or `PromptsDir/<name>.md`) and the text contains neither signal form, print to stderr:
   `warning: prompt file has no completion signal (<COMPLETION_SIGNAL>); the loop will only stop at max iterations`
3. Inline and stdin prompts are exempt (deliberate one-offs; the root command help already documents the signal for them).
4. Built-ins always contain the signal; the check is a no-op for them but is not special-cased.

## APIs

- `internal/prompt.HasCompletionSignal(text string) bool` — true if either signal form is present. Migrate the existing unexported `cli.HasCompletionSignal` (literal-only) or add the placeholder-aware variant beside it; `RunLoop` keeps its exact literal detection unchanged.
- `internal/prompt.ValidatePromptText(text string) []Check` — runs frontmatter/description/signal checks on text; resolution stays in the CLI layer.
- No new config fields.

## Configuration

- Uses existing `PromptsDir` only. See configuration spec for precedence.

## Permissions

- Read access to the target file and `PromptsDir`.

## Security Considerations

- `validate` prints check messages, never prompt content; match text stays out of output.

## Dependencies

- Standard library only.

## Open Questions / Risks

- Should `completion-signal` be a `fail` or a `warn` when the prompt is clearly a multi-iteration plan prompt intended to run to max iterations? (Current call: `fail`; revisit with user feedback.)
- Should `guide` content be a testable fixture or inline string? (Fixture file under `internal/cli`, embedded via `go:embed`, is the default.)

## Verifications

- `ralph prompts guide` prints the authoring contract; exit 0.
- `ralph prompts validate build` reports all checks ok; exit 0.
- A file containing `<COMPLETION_SIGNAL>` with a description line: all checks ok; exit 0.
- A file with no signal: `completion-signal` fails, exit 1.
- A file with unbalanced frontmatter: `frontmatter` warns, exit still driven by other checks.
- `ralph prompts validate nonexistent` errors; exit 1.
- `ralph run review` with a signal-less `prompts/review.md` prints the stderr warning once, then proceeds.

## Related Specifications

- [prompts.md](prompts.md) — `prompts list` / `prompts show`.
- [../prompts.md](../prompts.md) — prompt resolution precedence and completion signal.
- [run.md](run.md) — run flow where the missing-signal warning is emitted.
- [skill.md](skill.md) — Installing the embedded prompt-authoring skill that teaches this workflow.
