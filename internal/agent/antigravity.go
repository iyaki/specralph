package agent

import (
	"io"
	"os"
)

// AntigravityAgent implements the Agent interface for the Google Antigravity CLI.
type AntigravityAgent struct {
	Model     string
	AgentMode string
	Env       []string
}

// Execute runs agy with the given prompt.
// Antigravity CLI uses: agy -p <prompt> [--model <model>]
// Configuration via environment variables:
//   - ANTIGRAVITY_MODEL: model override (e.g., "Gemini 2.5 Pro", "Gemini 2.5 Flash")
//
// Note: Sandbox mode and tool permissions are configured via
// ~/.gemini/antigravity-cli/settings.json (managed by agy CLI), not Ralph config.
func (a *AntigravityAgent) Execute(prompt string, output io.Writer) (string, error) {
	args := []string{"-p", prompt}

	// Model selection (if specified)
	model := os.Getenv("ANTIGRAVITY_MODEL")
	if model == "" {
		model = a.Model
	}
	if model != "" {
		args = append(args, "--model", model)
	}

	// Note: Sandbox mode and tool permissions are configured via
	// ~/.gemini/antigravity-cli/settings.json, not CLI flags.
	// Users must configure these via interactive agy session (/config command).

	return executeAgentCommand("agy", args, a.Env, output, "agy")
}

// Name returns the name of the agent.
func (a *AntigravityAgent) Name() string {
	return "antigravity"
}

// IsAvailable checks if agy is available in PATH.
func (a *AntigravityAgent) IsAvailable() bool {
	return isAgentAvailable("agy")
}
