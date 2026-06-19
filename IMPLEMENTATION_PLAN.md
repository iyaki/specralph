# Implementation Plan (commands/help)

**Status:** Help Command Complete (Cobra auto-provided), Prompts Command Partially Complete (1/2)
**Last Updated:** 2026-06-19
**Primary Spec:** [specs/commands/help.md](specs/commands/help.md), [specs/commands/prompts.md](specs/commands/prompts.md)

---

## Quick Reference

| System/Module          | Spec                                    | Module/Package      | Status     |
|------------------------|-----------------------------------------|---------------------|------------|
| Help Command           | [specs/commands/help.md](specs/commands/help.md) | `internal/cli/cmd.go` | ✅ Complete |
| Prompts Command        | [specs/commands/prompts.md](specs/commands/prompts.md) | `internal/cli/prompts.go` | ⚠️ Partial  |
| Run Command (default)  | [specs/commands/run.md](specs/commands/run.md)     | `internal/cli/run.go` | ✅ Complete |
| Init Command           | [specs/commands/init.md](specs/commands/init.md)   | `internal/cli/init.go` | ✅ Complete |
| Version Command        | [specs/commands/version.md](specs/commands/version.md) | `internal/cli/version.go` | ✅ Complete |
| Completion Command     | Auto (Cobra)                            | N/A                 | ✅ Complete |

**Registered Commands** (from `internal/cli/cmd.go`):
- `init` ✅
- `run` ✅
- `version` ✅
- `help` ✅ (Cobra auto)
- `completion` ✅ (Cobra auto)
- `prompts` ⚠️ **PARTIAL** (list only, show missing)

---

## Phased Plan

### Phase 1: Help Command Verification (Complete)

**Goal:** Verify Cobra-provided help command meets spec requirements

**Status:** ✅ Complete

**Paths:**
- `internal/cli/cmd.go`
- `internal/cli/run.go`

**Checklist:**
- [x] `ralph help` outputs root help
- [x] `ralph help <command>` outputs command-specific help
- [x] `ralph --help` equivalent to `ralph help`
- [x] `ralph <command> --help` equivalent to `ralph help <command>`
- [x] Unknown commands return error: `unknown command "<name>" for "ralph"`
- [x] Help output includes Usage, Available Commands, Flags sections
- [x] Examples shown in root help (build, plan, custom prompt)
- [x] `prompts` command appears in Available Commands list ✅ (verified 2026-06-19)

**Definition of Done:**
- All Cobra default help behaviors verified working
- No custom implementation needed per spec

**Risks/Dependencies:** None - fully functional

---

### Phase 2: Prompts Command - List Subcommand (Complete)

**Goal:** Implement `prompts list` subcommand to discover and display available prompts

**Status:** ✅ Complete

**Paths:**
- `internal/cli/prompts.go` (exists)
- `internal/prompt/prompts.go` (supports built-in prompts)

**Checklist:**
- [x] Create `internal/cli/prompts.go` with `NewPromptsCommand()`
- [x] Implement `prompts list` subcommand
- [x] Discover built-in prompts (build, plan)
- [x] Output formatted list with Name, Description
- [x] Show usage hint: "Use 'ralph run <prompt-name>' to execute"
- [x] Command registered in `internal/cli/cmd.go`
- [x] Tests pass

**Verified:**
```bash
$ ralph prompts list
Built-in Prompts:
  build      Implement a single task from IMPLEMENTATION_PLAN.md after studying specs,
             then validate, commit, and update the plan.
  plan       Generate or update IMPLEMENTATION_PLAN.md with a phase-based plan after
             studying specs, existing code, and identifying gaps.

Use 'ralph run <prompt-name>' to execute a prompt.
```

**Definition of Done:**
- `prompts list` shows built-in prompts ✅
- `prompts` command appears in help output ✅

**Risks/Dependencies:** None - functional

---

### Phase 3: Prompts Command - Show Subcommand (Missing)

**Goal:** Implement `prompts show <name>` subcommand to display full prompt content

**Status:** ❌ Not Started

**Paths:**
- `internal/cli/prompts.go` (add `show` subcommand)
- `internal/prompt/prompts.go` (may need content retrieval function)
- `internal/cli/prompts_test.go` (new test file)

**Checklist:**
- [ ] Add `show` subcommand to `internal/cli/prompts.go`
- [ ] Implement `prompts show build` - outputs full built-in build prompt
- [ ] Implement `prompts show plan` - outputs full built-in plan prompt
- [ ] Implement `prompts show <custom>` - outputs custom prompt file content (frontmatter stripped)
- [ ] Implement error handling for nonexistent prompts
- [ ] Write tests in `prompts_test.go`
- [ ] Verify exit codes: 0 on success, non-zero on error

**Current Gap:**
```bash
$ ralph prompts show build
Error: accepts at most 2 arg(s), received 3  # ❌ Subcommand not implemented
```

**Definition of Done:**
- `prompts show build` outputs full build prompt content
- `prompts show plan` outputs full plan prompt content
- `prompts show nonexistent` returns error with non-zero exit code
- Tests pass with coverage ≥95%

**Risks/Dependencies:**
- Need to ensure prompt content generation functions are accessible from CLI layer

---

## Verification Log

### 2026-06-19: Help Command Verification

**Commands Run:**
- `./bin/ralph --help` ✅ Root help displays correctly
- `./bin/ralph help` ✅ Equivalent to --help
- `./bin/ralph help prompts` ✅ Shows prompts command help
- `./bin/ralph prompts --help` ✅ Shows prompts subcommands

**Results:**
- ✅ Help command fully functional via Cobra
- ✅ `prompts` command registered and visible in help
- ✅ Prompts command structure created with `list` subcommand

**Files Analyzed:**
- `internal/cli/cmd.go` - Command registration (line 47-49 adds init, run, version)
- `internal/cli/prompts.go` - New prompts command (implemented)
- `internal/prompt/prompts.go` - Prompt generation logic

### 2026-06-19: Prompts List Verification

**Commands Run:**
- `./bin/ralph prompts list` ✅

**Results:**
- ✅ Built-in prompts (build, plan) displayed with descriptions
- ✅ Usage hint included ("Use 'ralph run <prompt-name>' to execute")
- ✅ Command appears in help output under Available Commands

**Bugs Discovered:** None

### 2026-06-19: Prompts Show Gap Identified

**Commands Run:**
- `./bin/ralph prompts show build` ❌ Error - subcommand not implemented
- `./bin/ralph prompts show nonexistent` ❌ Error - subcommand not implemented

**Results:**
- ❌ `show` subcommand NOT implemented
- ✅ Infrastructure in place (prompts.go exists, prompt functions available)

**Files to Modify:**
- `internal/cli/prompts.go` - Add show subcommand

### 2026-06-19: Test Suite Verification

**Commands Run:**
- `make quality` ✅ All tests pass

**Results:**
- ✅ All existing tests pass
- ❌ No tests yet for `prompts show` (command doesn't exist)
- ✅ Existing code coverage maintains ≥95% (spec requirement)

**Files Tested:**
- `internal/cli/cmd_test.go` - Command structure tests
- `internal/prompt/prompts_test.go` - Prompt resolution tests
- `internal/prompt/prompts_internal_test.go` - Internal helper tests

---

## Summary

| Phase | Description                                | Status      | Completion |
|-------|--------------------------------------------|-------------|------------|
| 1     | Help Command Verification                  | ✅ Complete  | 100%       |
| 2     | Prompts Command - List Subcommand          | ✅ Complete  | 100%       |
| 3     | Prompts Command - Show Subcommand          | ❌ Pending   | 0%         |

**Remaining Effort:**
- Phase 3: Prompts Show Subcommand - ~2-3 hours
  - Add `show` subcommand to prompts.go: 1h
  - Hook up to existing prompt content functions: 0.5h
  - Write tests: 1h
  - Manual verification: 0.5h

---

## Known Existing Work

**Help Command:**
- Fully provided by Cobra framework (no custom code needed)
- Registered automatically when commands are added via `cmd.AddCommand()`
- Output format controlled by Cobra's help template system

**Prompts Command:**
- `internal/cli/prompts.go` exists with `prompts list` implementation
- `prompts list` discovers built-in prompts from `internal/prompt` package
- Command registered in `internal/cli/cmd.go` (verified via `ralph --help`)

**Registered Commands:**
- `init` - Interactive config setup (implemented)
- `run` - Prompt loop execution (implemented)
- `version` - Version info output (implemented)
- `completion` - Shell completion scripts (Cobra auto-generated)
- `prompts` - Prompt discovery (list ✅, show ❌)

**Prompt Infrastructure:**
- `internal/prompt/prompts.go` has `BuildPrompt()` and `PlanPrompt()` functions
- Built-in prompts (build, plan) already generated by these functions
- `GetPrompt()` function handles prompt resolution from multiple sources

---

## Manual Deployment Tasks

None.

---

## Next Actions

1. **Implement `prompts show <name>` subcommand** (Phase 3)
   - Add subcommand handler in `internal/cli/prompts.go`
   - Resolve prompt name (built-in vs custom)
   - Output full content (frontmatter stripped for custom)
   - Add error handling

2. **Write comprehensive tests**
   - Test built-in prompts (show build, show plan)
   - Test custom prompts (show with frontmatter, without)
   - Test error cases (nonexistent prompt, permission denied)

3. **Verify end-to-end**
   - Manual testing of all subcommands
   - Verify help output includes examples
   - Run `make quality` for full test suite + coverage