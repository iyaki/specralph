# E2E Coverage Matrix

Last Updated: 2026-04-20
Primary Spec: `specs/e2e-testing.md`

This matrix maps the supported CLI/config/output behavior surface to concrete e2e tests.
Test identifiers use Go `TestName[/Subtest]` notation.

## Commands and Routing

| Surface | Expected behavior | E2E tests |
| --- | --- | --- |
| `ralph` | No-arg invocation routes to `run build` | `TestE2ERunCommandRouting/RunDefaultsToBuildPrompt` |
| `ralph <prompt> [scope]` alias | Non-subcommand prompt names execute via run path | `TestE2EConfigPrecedence_PromptFileFromConfigFile`, `TestE2EConfigPrecedence_NoSpecsIndexFromConfigFile`, `TestE2EConfigByPromptOverrideFromConfigApplies` |
| `ralph init` | Registered subcommand wins routing collisions | `TestE2ERunCommandRouting/InitSubcommandWinsCollision`, `TestE2EInitCommand/InitWithoutTTYFailsFast` |
| `ralph run init` | `init` is treated as prompt name when invoked through `run` | `TestE2ERunCommandRouting/RunInitTreatsInitAsPromptName`, `TestE2EInitCommand/InitPromptNameRunsViaRunSubcommand` |

## Prompt Resolution and Prompt-Level Settings

| Surface | Expected behavior | E2E tests |
| --- | --- | --- |
| `--prompt` | Inline prompt text is used directly | `TestE2EInlinePrompt` |
| stdin prompt (`ralph -`) | Explicit stdin prompt source is supported | `TestE2EStdinPrompt` |
| stdin prompt (implicit) | Piped stdin is consumed when no prompt args are provided | `TestE2EStdinPrompt_Implicit` |
| `--prompt-file <path>` | Prompt file is loaded from explicit path | `TestE2ECompletionFlow`, `TestE2ELogging` |
| missing prompt file | Missing prompt file fails with non-zero exit and clear stderr | `TestE2EMissingPromptFile` |
| front matter overrides | `model` and `agent-mode` front matter values apply and are stripped from prompt body | `TestE2EConfigByPromptFrontMatterAppliesAndIsStripped` |
| invalid front matter | Malformed front matter fails before agent execution | `TestE2EConfigByPromptInvalidFrontMatterFailsBeforeAgentRun` |
| prompt override precedence | Env values override front matter values | `TestE2EConfigByPromptEnvOverridesFrontMatter` |
| config prompt overrides | `[prompt-overrides.<prompt>]` applies per prompt | `TestE2EConfigByPromptOverrideFromConfigApplies`, `TestE2EConfigLocalOverlay_PromptOverridesDeepMerge` |

## CLI Flag and Config-Source Coverage

| Surface | Expected behavior | E2E tests |
| --- | --- | --- |
| `--config`, `RALPH_CONFIG`, default `ralph.toml` | Base config source precedence is deterministic | `TestE2EConfigPrecedence_ConfigFlagOverride`, `TestE2EConfigPrecedence_RalphConfigEnvOverride`, `TestE2EConfigPrecedence_ConfigFileWins` |
| invalid base config parse | Malformed `ralph.toml` fails before agent execution | `TestE2EConfigPrecedence_InvalidBaseConfigFailsBeforeAgentExecution` |
| `ralph-local.toml` sibling overlay | Local overlay is discovered relative to selected base config and deep merged | `TestE2EConfigLocalOverlay_ConfigFlagUsesSiblingOverlay`, `TestE2EConfigLocalOverlay_RalphConfigEnvUsesSiblingOverlay`, `TestE2EConfigLocalOverlay_DefaultDiscoveryUsesSiblingOverlay`, `TestE2EConfigLocalOverlay_PromptOverridesDeepMerge` |
| invalid overlay parse | Malformed `ralph-local.toml` fails before agent execution | `TestE2EConfigLocalOverlay_InvalidOverlayFailsBeforeAgentExecution` |
| unsupported `config-file` key | `config-file` key in base or overlay config fails fast | `TestE2EConfigPrecedence_ConfigFileKeyInBaseConfigFails`, `TestE2EConfigPrecedence_ConfigFileKeyInOverlayFails` |
| `--max-iterations`, `RALPH_MAX_ITERATIONS`, config `max-iterations` | Precedence resolves as flags > env > file > defaults | `TestE2EConfigPrecedence_FlagWins`, `TestE2EConfigPrecedence_EnvWins`, `TestE2EConfigPrecedence_ConfigFileWins`, `TestE2EMaxIterations` |
| config `prompt-file` | File-sourced `prompt-file` applies when flag is not set | `TestE2EConfigPrecedence_PromptFileFromConfigFile` |
| `--specs-dir`, `--specs-index` | Custom specs directory/index are applied to generated prompt content | `TestE2ESpecsFlags/Custom_Specs_Dir_and_Index` |
| `--no-specs-index`, config `no-specs-index` | Specs index inclusion can be disabled by flag or config | `TestE2ESpecsFlags/No_Specs_Index`, `TestE2EConfigPrecedence_NoSpecsIndexFromConfigFile` |
| `--implementation-plan-name` | Custom implementation plan name is applied to prompt content | `TestE2EPlanFlags/Custom_Implementation_Plan_Name` |
| `--agent` | Supported agents select correct adapter; unknown agent fails fast | `TestE2EAgentSelection/SelectClaudeAgent`, `TestE2EAgentSelection/SelectCursorAgent`, `TestE2EAgentSelection/UnknownAgentReturnsError` |
| `--model` | Model override is forwarded to agent CLI args | `TestE2EModelFlags/ModelOverride`, `TestE2EConfigByPromptFrontMatterAppliesAndIsStripped` |
| `--agent-mode` | Agent mode override is forwarded to agent CLI args | `TestE2EModelFlags/AgentModeOverride`, `TestE2EConfigByPromptFrontMatterAppliesAndIsStripped` |
| `--env`, config `[env]` | Child process env override precedence and validation are deterministic | `TestE2EEnvOverrides/FlagOnlyOverride`, `TestE2EEnvOverrides/ConfigOnlyOverride`, `TestE2EEnvOverrides/FlagOverridesConfig`, `TestE2EEnvOverrides/RepeatedFlagLastWins`, `TestE2EEnvOverrides/InvalidEntryFailsBeforeExecution` |

## Logging and Observable Output Coverage

| Surface | Expected behavior | E2E tests |
| --- | --- | --- |
| stdout stream | Run metadata, iteration output, and completion signal are emitted | `TestE2ECompletionFlow`, `TestE2ERunCommandRouting/RunDefaultsToBuildPrompt`, `TestE2ELoggingStdoutParity` |
| stderr stream | Fatal startup/validation errors are emitted with non-zero exit | `TestE2EMissingPromptFile`, `TestE2EConfigByPromptInvalidFrontMatterFailsBeforeAgentRun`, `TestE2EConfigPrecedence_ConfigFileKeyInBaseConfigFails`, `TestE2EConfigLocalOverlay_InvalidOverlayFailsBeforeAgentExecution` |
| exit codes | Success and failure exit semantics are deterministic | `TestE2ECompletionFlow`, `TestE2EMaxIterations`, `TestE2EReturnErrorPath`, `TestE2EAgentSelection/UnknownAgentReturnsError` |
| deterministic slow-complete path | `RALPH_TEST_AGENT_MODE=slow_complete` delays deterministically before emitting completion | `TestE2ESlowCompletePath` |
  | log creation and disablement | Logging defaults to disabled and can be enabled by path flag/config | `TestE2ELoggingFlags/DefaultNoLog`, `TestE2ELoggingFlags/EnabledViaFlag`, `TestE2ELoggingFlags/EnabledViaConfig`, `TestE2ELoggingFlags/FlagOverridesConfig`, `TestE2ELoggingFlags/LogTruncate`, `TestE2ELogging` |
| log truncation and content parity | Truncation mode and stdout parity are preserved | `TestE2ELoggingFlags/LogTruncate`, `TestE2ELoggingStdoutParity` |
| log file security | Log file permissions are restrictive (`0600`) | `TestE2ELoggingPermissions` |

## Pending Gaps (Tracked in `IMPLEMENTATION_PLAN.md` Phase 7.2)

| Required behavior | Status |
| --- | --- |
| `RALPH_TEST_AGENT_MODE=return_error` scenario | Complete (`TestE2EReturnErrorPath`) |
| `RALPH_TEST_AGENT_MODE=slow_complete` deterministic delay scenario | Complete (`TestE2ESlowCompletePath`) |
