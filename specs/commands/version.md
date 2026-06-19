# Version Command

## Overview

The `version` subcommand displays the Ralph CLI version number and build metadata.

## Purpose

- Provide users with visibility into which version of Ralph is installed
- Support debugging and support scenarios with build metadata (commit hash, build date)
- Enable build-time version injection via linker flags without hardcoding values

## Command Interface

```bash
ralph version
```

### Flags

None. The command accepts no flags and outputs only the version string.

### Examples

**Default output:**
```bash
$ ralph version
ralph v0.0.0
```

**Release build:**
```bash
$ make build VERSION=1.2.3
$ ralph version
ralph v1.2.3
```

## Implementation Details

### Version Package (`internal/version`)

The `version` package provides build-time version information management:

```go
package version

// Build-time variables (injected via -ldflags)
var (
    Version = "dev"    // Semantic version (e.g., "1.2.3")
    Commit  = "unknown" // Git commit hash
    Date    = "unknown" // Build date in RFC3339 format
)

// Get returns complete version information
func Get() Info

// String returns formatted version string: "ralph v<version>"
func String() string

```

### Command Registration

The version command is registered in `internal/cli/cmd.go`:

```go
cmd.AddCommand(NewVersionCommand())
```

### Build-Time Injection

The Makefile injects version information using Go linker flags:

```makefile
VERSION ?= 0.0.0
COMMIT ?= $(shell git rev-parse --short HEAD 2>/dev/null || echo unknown)
DATE ?= $(shell date -u '+%Y-%m-%dT%H:%M:%SZ')

LDFLAGS := -X github.com/iyaki/specralph/internal/version.Version=$(VERSION) \
           -X github.com/iyaki/specralph/internal/version.Commit=$(COMMIT) \
           -X github.com/iyaki/specralph/internal/version.Date=$(DATE)

build:
	go build -ldflags "$(LDFLAGS)" -o bin/ralph ./cmd/ralph
```
**Release builds (GitHub Actions):**
- VERSION extracted from git tag (e.g., `v1.2.3` -> `1.2.3`)
- COMMIT from `git rev-parse --short HEAD`
- DATE in RFC3339 format

## Default Values

When version variables are not injected at build time (e.g., local development builds):

- `Version`: defaults to `"dev"`
- `Commit`: defaults to `"unknown"`
- `Date`: defaults to `"unknown"`

## Output Format

```
ralph v<version>
```

Where `<version>` is a semantic version string (e.g., `1.2.3`, `0.0.0`, `dev`).

The command always outputs a single line with the version number.

## Exit Codes

- `0`: Success (always exits with 0 when command executes)

## Testing

Tests are located in:
- `internal/version/version_test.go` - Version package tests
- `internal/cli/version_test.go` - Command structure and behavior tests

Run tests with:
```bash
make test
```

## Security Considerations

- Version information is public and non-sensitive
- Build metadata (commit hash) is already visible in git history
- No secrets or credentials are exposed through this command

## Dependencies

| Dependency | Purpose |
|------------|---------|
| `github.com/spf13/cobra` | Command structure and flag parsing |
| Makefile | Build-time version injection via ldflags |
| Go standard library | Basic I/O and string formatting |

## Related Specifications

- [release-workflow.md](release-workflow.md) - Release automation and versioning strategy
- [core-architecture.md](core-architecture.md) - Overall CLI architecture