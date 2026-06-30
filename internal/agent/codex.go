package agent

import (
	"io"
	"os"
	"strings"
)

// CodexAgent implements the Agent interface for the OpenAI Codex CLI.
type CodexAgent struct {
	Model     string
	AgentMode string
	Env       []string
}

// Execute runs codex with the given prompt.
// Codex CLI uses: codex exec [--model <model>] [--sandbox <policy>]
// [--ask-for-approval <mode>] [--ephemeral] [--json]
// [--output-last-message <path>] [--profile <name>] [-c key=value] <prompt>
// Configuration via environment variables:
//   - CODEX_SANDBOX: sandbox policy (default: read-only)
//   - CODEX_APPROVAL_MODE: approval mode (default: never)
//   - CODEX_EPHEMERAL: set to "true" to enable ephemeral mode (default: true)
//   - CODEX_OUTPUT_PATH: path to write final message
//   - CODEX_PROFILE: config profile name
func (a *CodexAgent) Execute(prompt string, output io.Writer) (string, error) {
	args := []string{"exec"}

	// Model selection
	if a.Model != "" {
		args = append(args, "--model", a.Model)
	}

	// Sandbox policy (default: read-only, configurable via CODEX_SANDBOX)
	sandbox := getEnvOrDefault("CODEX_SANDBOX", "read-only")
	args = append(args, "--sandbox", sandbox)

	// Approval mode (default: never for automation, configurable via CODEX_APPROVAL_MODE)
	approvalMode := getEnvOrDefault("CODEX_APPROVAL_MODE", "never")
	args = append(args, "--ask-for-approval", approvalMode)

	// Ephemeral mode (default: true for stateless execution, configurable via CODEX_EPHEMERAL)
	ephemeral := getEnvOrDefault("CODEX_EPHEMERAL", "true")
	if strings.ToLower(ephemeral) == "true" {
		args = append(args, "--ephemeral")
	}

	// Output path for final message (configurable via CODEX_OUTPUT_PATH)
	if outputPath := os.Getenv("CODEX_OUTPUT_PATH"); outputPath != "" {
		args = append(args, "--output-last-message", outputPath)
	}

	// Profile selection (configurable via CODEX_PROFILE)
	if profile := os.Getenv("CODEX_PROFILE"); profile != "" {
		args = append(args, "--profile", profile)
	}

	// Agent mode (if specified)
	if a.AgentMode != "" {
		args = append(args, "--agent", a.AgentMode)
	}

	// Add the prompt as the final positional argument
	args = append(args, prompt)

	return executeAgentCommand("codex", args, a.Env, output, "codex")
}

// Name returns the name of the agent.
func (a *CodexAgent) Name() string {
	return "codex"
}

// IsAvailable checks if codex is available in PATH.
func (a *CodexAgent) IsAvailable() bool {
	return isAgentAvailable("codex")
}

// getEnvOrDefault returns environment variable value or default if not set.
func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}

	return defaultValue
}
