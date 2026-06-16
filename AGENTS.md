# Agent Guidelines

## Spec-First Workflow

- Read `specs/README.md` before any feature work.
- Assume specs describe intent, not implementation.
- Verify reality in the codebase before claiming something exists.
- Implement to spec patterns and data shapes; update specs only when asked.
- When Writting specs, **NEVER** follow Test Driven Development practices. Write the spec first and stop.
- For programming tasks, always load Test Driven Development skill.

## Testing and Quality Gates

- Follow Test Driven Development practices: write failing tests before implementation.
- Local suite: `make quality`.
- Targeted runs:
  - `make lint|test|test-e2e|test-race|coverage|mutation|security|arch`.
- Coverage gate: min 95%.
- Run the full Go test suite with `make test` (includes `test/e2e`).
- Run only the end-to-end tests with `make test-e2e`.
- Execute mutation testing with `make mutation` ONLY in final stages of the task development. **NEVER** execute mutation testing during the Test Driven Development process.

## Build and Run

- Build the CLI binary: `make build`.
- Run from source (no build): `make run ARGS='<command> [flags]'`.
- Get help for available targets: `make help`.

## Tooling Expectations

- Go version: 1.25 (see `go.mod`).
- Mutation testing tool: `gremlins`.
- Lint and security via `golangci-lint`, `govulncheck`, `go-sec`, `go-arch-lint`, `go-fmt`.

## Implementation Guidance

- Keep scans deterministic and reproducible.
- Skip binary/oversized files per spec; record skipped file stats.
- Treat match text as sensitive; avoid logging it in console.
- When multiple code paths do similar work with small variations, consolidate into shared services with request structs.
