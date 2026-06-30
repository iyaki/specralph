# GitHub Copilot CLI Agent

Status: Proposed

## Overview

### Purpose

- Define how Ralph executes the `copilot` CLI agent from GitHub.
- Specify command invocation, flags, and runtime behavior for non-interactive execution.
- Document custom instructions and agent profile integration for project context loading.

### Goals

- Document `copilot` invocation shape with all relevant flags for scripting and automation.
- Describe model selection, sandbox policies, and permission modes.
- Specify error handling for CLI missing, authentication failures, and permission timeouts.
- Define custom instructions (`.github/copilot-instructions.md`, `AGENTS.md`) loading behavior.
- Document built-in custom agents and how to invoke them.

### Non-Goals

- Describing Copilot internal model behavior or tool selection.
- Implementing interactive TUI mode (Ralph uses non-interactive execution only).
- Covering Copilot Cloud remote session workflows.

### Scope

- In scope: `copilot` CLI invocation with `-p/--prompt` flag, ephemeral sessions, custom instructions loading.
- Out of scope: prompt resolution (see prompts spec), config precedence (see configuration spec), interactive session management.

## Architecture

### Module/package layout (tree format)

```
internal/
  agent/
    copilot.go
specs/
  agents/
    copilot.md
```

### Component diagram (ASCII)

```
+------------------+
| CopilotAgent     |
| internal/agent   |
+---------+--------+
          |
          | builds args
          v
+---------+--------+
| copilot CLI      |
| --prompt         |
| --model          |
| --sandbox        |
| --allow-*        |
| --no-interactive |
+---------+--------+
          |
          | executes
          v
+---------+--------+
| GitHub Copilot   |
| API              |
+------------------+
```

### Data flow summary

1. Ralph selects `copilot` when `AgentName` is `copilot`.
2. The agent builds CLI args: `-p/--prompt`, optional `--model`, `--sandbox`, `--allow-all`, `[--yolo](file:///workspaces/ralph/internal/suggestions/suggestions.go#L46-L49)`, `--no-interactive`, and custom agent selection.
3. Copilot CLI discovers and loads custom instructions (global → repository → path-specific).
4. The agent executes `copilot -p <prompt>` and captures output.
5. Output is streamed to stdout; completion detected via `<promise>COMPLETE</promise>` signal.

## Data model

### Core Entities

- CopilotAgent
  - Fields: `Model`, `Sandbox`, `AllowAll`, `Agent`, `ResumeSession`.
  - Implements `Agent` interface: `Execute(prompt string, output io.Writer) (string, error)`, `Name() string`, `IsAvailable() bool`.

### Relationships

- Selected by `GetAgent` based on `AgentName == "copilot"`.
- Uses `Model` configuration field for model selection (e.g., `gpt-4o`, `o1`, `o3`, `o4-mini`, or `auto`).
- Uses `Sandbox` configuration field for sandbox policy (`enable`, `disable`).
- Uses `AllowAll` configuration field to enable all permissions (`--allow-all` or `[--yolo](file:///workspaces/ralph/internal/suggestions/suggestions.go#L46-L49)`).
- Uses `Agent` configuration field for custom agent selection (e.g., `explore`, `task`, `research`).
- `ResumeSession` controls session resumption (default: `false` for ephemeral execution).

### Persistence Notes

- By default, Ralph uses non-interactive execution without session persistence.
- When `ResumeSession` is set, Copilot can resume a previous session with `--continue` or `--resume <session-id>`.
- Sessions are stored under `~/.copilot/sessions/` (or `COPILOT_HOME` directory).
- Ralph defaults to ephemeral execution unless explicitly configured to resume sessions.

## Workflows

### Execute copilot (happy path)

1. Verify `copilot` binary availability via `LookPath`.
2. Build args:
   - Base: `-p <prompt>` or `--prompt <prompt>` (required, positional or flagged)
   - Model: `--model <Model>` if specified (or `auto` for auto-selection)
   - Sandbox: `--sandbox enable` (default) or `--sandbox disable`
   - Permissions: `--allow-all` or `[--yolo](file:///workspaces/ralph/internal/suggestions/suggestions.go#L46-L49)` if `AllowAll` is true
   - Custom agent: `--agent <Agent>` if specified (e.g., `explore`, `task`, `research`, `general-purpose`)
   - Session: `--continue` to resume last session, or `--resume <session-id>` for specific session
   - Quiet mode: `-q` to reduce output verbosity (optional)
   - Prompt: final positional argument (if not using `-p` flag)
3. Execute `copilot` CLI with combined args.
4. Stream stdout/stderr to output writer.
5. Parse output for completion signal (Copilot CLI does not use `<promise>` tags; completion is implicit when output ends).
6. Return combined output and any error.

### Execute copilot (CLI missing)

1. `IsAvailable()` returns false (binary not in PATH).
2. Warning is printed to stderr.
3. Execution is still attempted (may fail with OS-level error).
4. Error is returned along with any partial output.

### Execute copilot (authentication failure)

1. Copilot CLI exits with error (missing or expired authentication).
2. Error message indicates authentication issue.
3. Ralph returns error with message: "GitHub Copilot authentication required. Run 'copilot login' first."
4. User must run `copilot login` to authenticate via GitHub OAuth device flow or personal access token.

### Execute copilot (permission timeout)

1. If Copilot requests permission for a tool/path/URL and running in non-interactive mode.
2. Copilot may hang waiting for approval input.
3. Ralph should enforce a timeout on execution (configurable, default 300s).
4. On timeout, process is terminated; error returned: "Copilot permission timeout exceeded".
5. Mitigation: Use `--allow-all` or `[--yolo](file:///workspaces/ralph/internal/suggestions/suggestions.go#L46-L49)` flag in trusted environments.

### Execute copilot (sandbox violation)

1. Copilot attempts command outside allowed sandbox scope.
2. With `--sandbox enable`: file modifications and network access restricted.
3. With `--sandbox disable`: shell commands allowed without restrictions.
4. Copilot returns error or fails silently; Ralph propagates with context.

### Custom instructions loading (Copilot automatic)

1. **Before execution**, Copilot automatically discovers and loads custom instructions:
   - User-level: `~/.copilot/instructions.md` or `~/.copilot/AGENTS.md`.
   - Organization-level: `.github-private/instructions.md` in organization/enterprise repo.
   - Repository-level: `.github/copilot-instructions.md` at project root.
   - Path-specific: `.github/instructions/**/*.instructions.md` (glob pattern).
   - Agent files: `AGENTS.md`, `.agents.md`, or directory-level agent profiles in `.github/agents/`.
2. Files loaded in order (user → organization → repository → path-specific), with later files augmenting earlier ones.
3. Combined size limited by configuration (default 64 KiB).
4. Empty or malformed files skipped.
5. **Ralph does not manage custom instructions** — Copilot handles discovery automatically.
6. To customize: user creates/edits instruction files directly; Copilot reloads on each `copilot -p` run.

### Custom agents

1. Ralph can invoke built-in custom agents via `--agent <name>`:
   - `explore`: Quick codebase analysis without adding to main context.
   - `task`: Execute commands (tests, builds) with brief summaries on success, full output on failure.
   - `general-purpose`: Complex, multi-step tasks in separate context.
   - `code-review`: Review changes with focus on genuine issues only.
   - `research`: Deep research across codebase, repositories, and web with citations.
   - `rubber-duck`: Constructive critic (auto-consulted by Copilot).
2. Custom agents specified in config or via selection logic.
3. Ralph passes `--agent <name>` flag to invoke a specific agent.

## APIs

- None. This is a local CLI integration with GitHub Copilot API access managed by Copilot.

## Client SDK Design

- Not applicable.

## Configuration

### Ralph Configuration Fields

| Field           | Type   | Default           | Description                                      |
| --------------- | ------ | ----------------- | ------------------------------------------------ |
| `AgentName`     | string | `"copilot"`       | Must be `"copilot"` for this agent.              |
| `Model`         | string | `"auto"`          | Model override (e.g., `gpt-4o`, `o1`, `o3`, `o4-mini`, or `auto`). |
| `Sandbox`       | string | `"enable"`        | Sandbox policy (`enable`, `disable`).            |
| `AllowAll`      | bool   | `false`           | Enable all permissions (tools, paths, URLs).     |
| `Agent`         | string | `""`              | Custom agent name (optional, e.g., `explore`, `task`). |
| `ResumeSession` | bool   | `false`           | Resume last session instead of ephemeral run.    |

### Copilot CLI Flag Mapping

| Ralph Config    | Copilot Flag                        | Notes                                               |
| --------------- | ----------------------------------- | --------------------------------------------------- |
| `Model`         | `--model <MODEL>`                   | Overrides config; `auto` for model auto-selection. |
| `Sandbox`       | `--sandbox <enable|disable>`        | Enable or disable shell command sandboxing.        |
| `AllowAll`      | `--allow-all` or `[--yolo](file:///workspaces/ralph/internal/suggestions/suggestions.go#L46-L49)` | Enable all permissions for the session.            |
| `Agent`         | `--agent <NAME>`                    | Invoke a specific custom agent.                    |
| `ResumeSession` | `--continue` or `--resume <ID>`     | Resume last or specific session.                   |

### Custom Instructions Configuration (Copilot-managed)

| Config Field                  | Default   | Description                                      |
| ----------------------------- | --------- | ------------------------------------------------ |
| `project_doc_max_bytes`       | 65536     | Max combined size of custom instruction files.   |
| `custom_instructions_files`   | `[]`      | Additional filenames (e.g., `["TEAM_GUIDE.md"]`). |

**Note**: These are configured in `~/.copilot/config.json` or `COPILOT_HOME`, not Ralph config. Ralph does not manage custom instructions content.

### Example Configuration

```toml
# Ralph config.toml
agent_name = "copilot"
model = "gpt-4o"
sandbox = "enable"
allow_all = false
agent = "explore"
resume_session = false

# Optional: Copilot-specific config in ~/.copilot/config.json
# {
#   "model": {
#     "name": "gpt-4o"
#   },
#   "sandbox": {
#     "enabled": true
#   }
# }
```

## Permissions

- Requires OS permission to execute `copilot` binary.
- Requires file system access to read custom instructions files (handled by Copilot).
- Requires network access for GitHub Copilot API calls (handled by Copilot).
- Sandbox policies restrict shell command execution:
  - `enable`: Restricts filesystem writes and network access.
  - `disable`: Unrestricted shell command execution.
- `--allow-all` or `[--yolo](file:///workspaces/ralph/internal/suggestions/suggestions.go#L46-L49)`: Disables permission prompts; suitable only for trusted environments.

## Security Considerations

### Sandbox Bypass Risks

- **`--sandbox disable`**: Allows arbitrary shell commands, network access, and filesystem operations. Only use in:
  - Isolated CI runners or containers.
  - Environments with no sensitive data or credentials.
  - Scenarios where the cost of compromise is acceptable.
- **`--sandbox enable`**: Safer default; restricts potentially dangerous operations.
- **`--allow-all` or `[--yolo](file:///workspaces/ralph/internal/suggestions/suggestions.go#L46-L49)`**: Disables all permission prompts. Use only when:
  - Running in trusted, isolated environments.
  - Automating well-tested workflows with predictable commands.

### Code Execution

- Copilot can execute arbitrary shell commands via `bash`, `sh`, or other interpreters.
- Commands are subject to sandbox and permission policies.
- **Risk**: Malicious prompts could attempt to exploit Copilot tool access.
- **Mitigation**:
  - Use `--sandbox enable` for untrusted prompts.
  - Avoid `--allow-all` in shared or production environments.
  - Audit custom instructions files for injection risks (loaded verbatim).

### Authentication

- Copilot requires authentication via:
  - `copilot login` (GitHub OAuth device flow).
  - `COPILOT_GITHUB_TOKEN`, `GH_TOKEN`, or `GITHUB_TOKEN` environment variables.
  - Fine-grained personal access tokens with "Copilot Requests" permission.
- **Risk**: Tokens exposed in process lists, logs, or environment.
- **Mitigation**:
  - Use system credential store (default behavior).
  - Never set tokens as job-level environment variables in CI when running untrusted code.
  - Use inline export: `COPILOT_GITHUB_TOKEN=<token> copilot -p ...` for single invocation.
  - Store config files with restrictive permissions (`chmod 600`).
  - Rotate tokens regularly.

### Prompt Injection via Custom Instructions

- Custom instructions files are loaded into the system prompt.
- **Risk**: Malicious content could override instructions or exfiltrate data.
- **Mitigation**:
  - Version-control custom instructions files; review changes via pull requests.
  - Limit file size to prevent excessive token consumption.
  - Audit global and repository-level instructions periodically.

### Process List Exposure

- Prompts passed as CLI arguments may appear in `ps` output.
- **Mitigation**: For highly sensitive prompts, pipe via stdin (Copilot supports stdin input).

### Binary Trust

- `copilot` binary resolved from PATH.
- **Risk**: Malicious binary substitution could intercept prompts or credentials.
- **Mitigation**:
  - Install Copilot via official channels (npm, Homebrew, WinGet, official script).
  - Verify binary checksums from GitHub releases.
  - Use absolute path in Ralph config if PATH is untrusted.

## Dependencies

| Dependency       | Purpose                           | Required | Version Constraints |
| ---------------- | --------------------------------- | -------- | ------------------- |
| `copilot` CLI    | GitHub Copilot CLI agent          | Yes      | Latest stable       |
| Go `os/exec`     | Execute external CLI              | Yes      | Standard library    |
| Go `bytes`, `io`| Output buffering and streaming    | Yes      | Standard library    |

### Installation

```bash
# Install Copilot CLI via npm (Node.js 22+ required)
npm install -g @github/copilot

# Prerelease version
npm install -g @github/copilot@prerelease

# macOS/Linux via Homebrew
brew install --cask copilot-cli

# macOS/Linux via install script
curl -fsSL https://gh.io/copilot-install | bash

# Windows via WinGet
winget install GitHub.Copilot

# Authenticate
copilot login
```

### Prerequisites

- Active GitHub Copilot subscription (all plans).
- For organization/enterprise users: Copilot CLI policy must be enabled by admin.
- Node.js 22+ (for npm installation).
- PowerShell v6+ (Windows only).

## Testing Strategy

### Unit Tests

- Test `CopilotAgent.Execute()` with various flag combinations.
- Verify flag construction from config fields.
- Test `IsAvailable()` with/without binary in PATH.
- Mock CLI execution and verify output capture.

### Integration Tests

- Run `copilot -p "echo hello"` (or similar benign prompt) and verify output.
- Test with `--sandbox enable/disable` flags.
- Test custom agent invocation: `--agent explore`.
- Test model selection: `--model gpt-4o`.
- Test `--allow-all` mode in trusted environment.

### End-to-End Tests

- E2E test: `ralph --agent copilot build` with a simple prompt.
- Verify authentication error handling (unauthenticated state).
- Verify timeout handling for permission prompts.
- Test session resumption with `--continue`.

## Open Questions / Risks

- How does Copilot CLI signal task completion in non-interactive mode? (No `<promise>` tags; relies on output end.)
- Should Ralph enforce a timeout on all Copilot executions by default?
- Should `--allow-all` be opt-in only via explicit config, never default?
- How to handle Copilot's request for its own update prompts?
- Should Ralph parse Copilot's JSON output mode (if available) for structured progress?

## Verifications

- `ralph --agent copilot build` invokes `copilot -p`.
- `ralph --agent copilot --model gpt-4o build` includes `--model gpt-4o`.
- `ralph --agent copilot --agent explore build` includes `--agent explore`.
- `ralph --agent copilot build` with `sandbox: disable` passes `--sandbox disable`.
- Authentication failure prints actionable error.
- Permission timeout terminates process and returns error.

## Appendices

### Invocation Examples

#### Simple prompt (ephemeral, sandboxed)

```bash
copilot -p "Fix the bug in @src/app.js" --sandbox enable
```

#### With model selection and custom agent

```bash
copilot -p "Explain the authentication flow" --model gpt-4o --agent explore
```

#### With all permissions (trusted environment)

```bash
copilot -p "Refactor the API endpoints" --allow-all --sandbox disable
```

#### Resume last session

```bash
copilot --continue -p "Continue the previous refactoring"
```

#### Quiet mode for automation

```bash
copilot -q -p "Run tests and report failures" --agent task
```

### Built-in Custom Agents

| Agent            | Slug               | Purpose                                           |
| ---------------- | ------------------ | ------------------------------------------------- |
| Explore          | `explore`          | Quick codebase analysis, isolated from main context. |
| Task             | `task`             | Run commands with concise summaries.              |
| General Purpose  | `general-purpose`  | Complex multi-step tasks in separate context.     |
| Code Review      | `code-review`      | Review changes, focus on genuine issues.          |
| Research         | `research`         | Deep research across codebase and web.            |
| Rubber Duck      | `rubber-duck`      | Auto-consulted for second opinions.               |

### Authentication Methods

1. **OAuth Device Flow** (Recommended)
   ```bash
   copilot login
   ```
   - Opens browser for GitHub authentication.
   - Token stored in system credential store.

2. **Personal Access Token** (For automation)
   - Create fine-grained PAT with "Copilot Requests" permission.
   - Export token: `COPILOT_GITHUB_TOKEN=github_pat_...`
   - Or use `GH_TOKEN` / `GITHUB_TOKEN`.

3. **GitHub CLI Token Reuse**
   - Uses token from `gh auth` if available.

### Output Parsing

- Copilot CLI outputs plain text to stdout.
- No structured completion signal (unlike `<promise>COMPLETE</promise>`).
- Completion inferred when CLI exits.
- For progress tracking, parse tool invocation markers (e.g., `[Tool: write_file]`).

### Environment Variables

| Variable                 | Purpose                                                      |
| ------------------------ | ------------------------------------------------------------ |
| `COPILOT_GITHUB_TOKEN`   | GitHub authentication token (highest precedence).            |
| `GH_TOKEN`               | GitHub token (from `gh` CLI).                                |
| `GITHUB_TOKEN`           | GitHub token (CI environments).                              |
| `COPILOT_HOME`           | Override `~/.copilot` config/session directory.              |
| `COPILOT_SANDBOX`        | Default sandbox policy (`enable`/`disable`).                 |
| `COPILOT_MODEL`          | Default model selection.                                     |
| `NO_COLOR`               | Disable colored output.                                      |
| `FORCE_COLOR`            | Force colored output.                                        |

### Keyboard Shortcuts (Interactive Mode - Reference Only)

Ralph uses non-interactive execution, but these shortcuts are relevant for manual Copilot CLI use:

| Shortcut          | Purpose                                     |
| ----------------- | ------------------------------------------- |
| `@ FILENAME`      | Include file in context.                    |
| `! COMMAND`       | Run shell command directly.                 |
| <kbd>Shift+Tab</kbd> | Cycle between standard/plan/autopilot modes. |
| <kbd>Ctrl+C</kbd>    | Cancel operation.                           |
| `/agent`          | Browse custom agents.                       |
| `/model`          | Switch AI model.                            |
| `/sandbox enable` | Enable sandbox for session.                 |
| `/allow-all`      | Enable all permissions.                     |