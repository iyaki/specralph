# Development and Testing

## Overview

### Purpose

- Define the standard development workflow for building, running, and testing Ralph.
- Provide a testable reference for build and test commands and required tooling.

### Goals

- Document build, test, and coverage gate commands.
- Specify local run workflow and debug mode behavior.
- Capture required tooling versions and code style expectations.
- Reference dedicated test strategy specs, including end-to-end coverage.

### Non-Goals

- Release or deployment procedures.
- CI/CD pipeline design.

### Scope

- In scope: local development commands, test workflows, coverage gate, and debug mode.
- In scope: pointers to specialized testing specs that define deeper coverage requirements.
- Out of scope: external agent CLI installation and system packaging.

### Related specs

- End-to-end testing requirements and scenario matrix are defined in [e2e-testing.md](e2e-testing.md).

## Architecture

### Module/package layout (tree format)

```
cmd/
  ralph/
internal/
  cli/
```

### Component diagram (ASCII)

```
+------------------+
| Makefile Targets |
+---------+--------+
          |
          v
+---------+--------+
| go build/test    |
+---------+--------+
          |
          v
+---------+--------+
| ralph binary     |
+------------------+
```

### Data flow summary

1. Build uses `go build` to produce the `ralph` binary.
2. Tests run with `go test` across all packages.
3. Coverage gate enforces >= 95% using `go tool cover`.

## Data model

### Core Entities

- BuildTarget
  - Binary name: `ralph`.
  - Source entry point: `cmd/ralph/main.go`.

- TestSuite
  - Scope: `./...` (all packages, including `test/e2e`).
 
- E2ESuite
  - Scope: `./test/e2e` only.
  - Coverage threshold: 95% minimum.

### Relationships

- Build output is required for local CLI runs.
- Debug mode short-circuits the agent loop for faster test runs.

### Persistence Notes

- Build artifacts are local and not committed.

## Workflows

### Build (Makefile)

1. Run `make build`.
2. `go build -o bin/ralph ./cmd/ralph` produces the binary.

### Build (direct)

1. Run `go build -o bin/ralph ./cmd/ralph`.

### Test (full suite)

1. Run `make test`.
2. Executes `go test -v ./...` (including `./test/e2e`).

### Test (end-to-end)

1. Run `make test-e2e`.
2. Executes `go test -v ./test/e2e`.

### Test (coverage gate)

1. Run `make test-coverage`.
2. Uses a temporary coverprofile with `-covermode=atomic`.
3. Fails if total coverage is below 95%.

### Local run

1. Build the binary.
2. Execute `./bin/ralph` with desired arguments.

### Debug mode

1. Set `DEBUG=1` in the environment.
2. Run any CLI command.
3. The loop prints the prompt and exits after the first iteration.

## Configuration

- `DEBUG=1` enables debug mode to short-circuit the agent loop.
- Other configuration options are specified in the configuration spec.

## Permissions

- Requires local file system permissions to build and run binaries.

## Security Considerations

- Debug mode can output full prompts; avoid running with secrets in shared logs.

## Dependencies

| Dependency | Purpose                        |
| ---------- | ------------------------------ |
| Go 1.25    | Build and test toolchain       |
| Make       | Build/test convenience targets |

## Verifications

- `make build` produces `bin/ralph`.
- `make test` completes successfully and covers the full Go suite.
- `make test-e2e` executes only the e2e suite successfully.
- `make coverage` (or `make test-coverage`) fails when coverage is below 95%.
- `DEBUG=1 ./bin/ralph plan` exits after one iteration.

## Appendices

- Code style: `gofmt`.
