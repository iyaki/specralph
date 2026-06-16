# Contributing to Ralphex

Thanks for your interest in contributing.

This project follows a specs-first workflow with test-driven implementation. Please read this guide before opening a pull request.

## Prerequisites

- Go 1.25 (see `go.mod`)
- `make`
- Quality toolchain:
  - `golangci-lint`
  - `govulncheck`
  - `gosec`
  - `go-arch-lint`
  - `gremlins` (mutation testing, final-stage checks)

## Quickstart

```bash
make deps
make build
make test
```

Use a dev container (recommended for a reproducible setup):

- Open this repository in VS Code and run `Dev Containers: Reopen in Container`. Or use any other [compatible editor](https://containers.dev/supporting) with dev container support.
- Reference docs: https://code.visualstudio.com/docs/devcontainers/containers
- Project config: `.devcontainer/devcontainer.json`

Run from source:

```bash
make run ARGS='--help'
```

Show all available targets:

```bash
make help
```

## Specs-First Workflow

Ralphex is built around intentional specs and deterministic behavior.

1. Start with `specs/README.md` and read the relevant specs for your change.
2. Treat specs as the source of intent, then verify actual behavior in code and tests.
3. Keep implementation aligned with spec patterns and data shapes.
4. For programming tasks, use TDD: write a failing test first, then implement the smallest passing change.
5. Update specs only when behavior is intentionally changing.

Practical tips:

- In your PR, list the spec file(s) that informed your change.
- If implementation and spec diverge, align intentionally (with a spec update when needed).
- If your contribution is spec-only, update the spec first and stop there.

## Build and Run

Build with Make:

```bash
make build
```

Build directly with Go:

```bash
go build -o bin/ralph ./cmd/ralph
```

Run from source without building a binary:

```bash
make run ARGS='--help'
```

Cross-compile examples:

```bash
# Linux
GOOS=linux GOARCH=amd64 go build -o bin/ralph-linux ./cmd/ralph

# macOS
GOOS=darwin GOARCH=amd64 go build -o bin/ralph-darwin ./cmd/ralph

# Windows
GOOS=windows GOARCH=amd64 go build -o bin/ralph.exe ./cmd/ralph
```

## Testing and Quality Gates

Focused checks while iterating:

```bash
make lint
make test
make test-e2e
make test-race
make coverage
make security
make arch
```

One-command gate:

```bash
make quality
```

Notes:

- Coverage minimum is 95%.
- `make test` runs the full Go suite, including `test/e2e`; use `make test-e2e` to run only the end-to-end package.
- `make quality` includes mutation testing (`make mutation`), which can be slow. Use it in final validation stages.

## Local Hooks (Recommended)

This repository includes `lefthook` configuration in `lefthook.yml`.

Pre-commit currently runs formatting plus Go quality checks on `*.go` files.

If you use `lefthook`, install it and run once:

```bash
lefthook install
```

## Project Structure

```text
.
├── cmd/
│   └── ralph/main.go        # CLI entry point
├── internal/
│   ├── agent/               # Agent implementations
│   ├── cli/                 # Cobra commands
│   ├── config/              # Config loading/writing
│   ├── executor/            # Command execution helpers
│   ├── logger/              # Logging
│   └── prompt/              # Prompt rendering
├── specs/                   # Technical specifications
├── test/e2e/                # End-to-end tests
├── go.mod
├── go.sum
└── Makefile
```

### Agent Files

Each agent implementation is in its own file:

- `internal/agent/agent.go`: agent interface and factory
- `internal/agent/opencode.go`: OpenCode CLI integration
- `internal/agent/claude.go`: Claude CLI integration
- `internal/agent/cursor.go`: Cursor CLI integration

## Adding Support for a New Agent

Use the agent workflow skills:

- `.agents/skills/agent-spec-creation/SKILL.md`
- `.agents/skills/agent-implementation/SKILL.md`

Recommended flow:

1. Create or update specs first:
   - Use `agent-spec-creation`.
   - Rely on `spec-creator` for spec structure/quality.

2. Study the target CLI first:

   ```bash
   <agent-cli> --help
   <agent-cli> <subcommand> --help
   ```

3. Implement integration:
   - Use `agent-implementation` after specs are ready.
   - Add `internal/agent/<new-agent>.go`.
   - Implement the `Agent` interface in `internal/agent/agent.go`.
   - Wire the new agent in the factory.
   - Keep behavior deterministic and avoid logging sensitive prompt/match text.
   - Reuse shared helpers instead of duplicating execution logic.

4. Add tests in the same change.

At minimum, cover:

- Factory returns the new implementation when selected.
- Invalid/unsupported names return expected errors.
- Command/argument composition in normal execution.
- Failure paths (missing binary, non-zero exit, malformed output when relevant).

Useful targeted command while iterating:

```bash
go test -v ./internal/agent/...
```

Final validation sequence:

```bash
make test
make test-e2e
make quality
```

## Dependency Management

Add dependencies and clean up modules:

```bash
go get <module>
go mod tidy
```

## Pull Requests

Please keep PRs focused and easy to review.

- Include problem statement, approach, and test evidence.
- Reference the relevant spec files.
- Update docs/specs when behavior changes.
- Prefer small, single-purpose PRs.
