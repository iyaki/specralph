# Logging

Status: Implemented

## Overview

### Purpose

- Define how Ralph initializes logging and writes logs to disk.
- Provide a testable description of log enablement, file selection, and headers.

### Goals

- Specify when logging is enabled or disabled.
- Describe log file creation, append/truncate behavior, and headers.
- Document git metadata capture in log headers.

### Non-Goals

- Remote log shipping or structured logging.
- Log rotation or retention policies.

### Scope

- In scope: log enablement, file creation, headers, and lifecycle.
- Out of scope: agent output semantics and prompt content rules.

## Architecture

### Module/package layout (tree format)

```
internal/
  logger/
    logger.go
```

### Component diagram (ASCII)

```
+------------------+
| Logger           |
| internal/logger  |
+---------+--------+
          |
          v
+---------+--------+
| Log File (disk)  |
+------------------+
```

### Data flow summary

1. CLI initializes a logger after configuration is loaded.
2. Logging is enabled only when a log file path is specified (via `--log-file`, `log-file` config, or `RALPH_LOG_FILE` env).
3. If enabled, the logger opens/creates the log file.
4. A run header and git metadata are written at startup.
5. CLI writes output to stdout and the log file via a multi-writer.

## Data model

### Core Entities

- Logger
  - Fields: `enabled`, `file`.
  - Responsibilities: manage log file lifecycle and header writing.

- LogFile
  - Fields: `Path`, `Append` (bool).
  - Derived from config fields and environment variables.

### Relationships

- Logger behavior depends on configuration field `LogFile` and `LogTruncate`.
- Environment variable `RALPH_LOG_FILE` can set the log path; `RALPH_LOG_APPEND=0` forces truncation.
| Store | Format     | Location                   | Notes                                                      |
| ----- | ---------- | -------------------------- | ---------------------------------------------------------- |
| Logs  | Plain text | `./ralph.log` when enabled | Header includes timestamp and git metadata; unresolved git values are recorded as `N/A`. |

## Workflows

### Initialize logging (disabled — default)

1. Logging is disabled by default (empty `LogFile`).
2. Enable by setting `--log-file`, `log-file` in config, or `RALPH_LOG_FILE` env.
3. When disabled, logger returns without a file.

### Initialize logging (enabled)

1. Create log directory if it does not exist.
2. Open file in append or truncate mode.
3. Write header with timestamp and git branch/commit (or `N/A` when git metadata cannot be resolved).

### Close logging

1. On CLI exit, logger closes the file if present.

## APIs

- None. Logging is internal.

## Client SDK Design

- Not applicable.

## Configuration

- See configuration spec for option definitions and precedence.
- Relevant fields:

## Permissions

- Requires write access to log file path.
- Requires directory creation permissions for the log file folder.

## Security Considerations

- Logs may contain prompt text and agent outputs; treat log files as sensitive.
- Log file paths should avoid world-writable directories to reduce tampering risk.

## Dependencies

- Standard library only (`os`, `os/exec`, `path/filepath`, `time`).

## Open Questions / Risks

- Should log header include the config file path to aid debugging?
- Should logging be disabled by default in CI environments?

## Verifications

- By default, no log file is created (empty `LogFile`).
- With `log-file` set in config, `--log-file` flag, or `RALPH_LOG_FILE` env, a log file is created.
- With `RALPH_LOG_APPEND=0`, log file is truncated on start.
## Appendices

- None.
