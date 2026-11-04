// Package agent provides integrations for supported AI agent CLIs.
package agent

import (
	"fmt"
	"io"
	"strings"
)

// Agent represents an AI agent CLI that can execute prompts.
type Agent interface {
	// Execute runs the agent with the given prompt and returns the output.
	Execute(prompt string, output io.Writer) (string, error)

	// Name returns the name of the agent.
	Name() string

	// IsAvailable checks if the agent CLI is available on the system.
	IsAvailable() bool
}

// supportedAgents lists canonical agent names in the order they are offered
// by `ralph init` and referenced in error messages.
var supportedAgents = []string{"oh-my-pi", "opencode", "claude", "cursor", "codex", "copilot", "antigravity"}

// agentAliases maps accepted aliases to their canonical agent name.
var agentAliases = map[string]string{
	"agy": "antigravity",
	"omp": "oh-my-pi",
}

// SupportedAgents returns the canonical supported agent names.
func SupportedAgents() []string {
	return append([]string(nil), supportedAgents...)
}

// IsSupportedAgent reports whether name is a canonical agent name or an accepted alias.
func IsSupportedAgent(name string) bool {
	for _, supported := range supportedAgents {
		if name == supported {
			return true
		}
	}
	_, ok := agentAliases[name]

	return ok
}

// GetAgent returns the appropriate agent based on configuration.
func GetAgent(agentName, model, agentMode string, env []string) (Agent, error) {
	effectiveEnv := cloneStringSlice(env)

	switch agentName {
	case "oh-my-pi", "omp":
		return &OmpAgent{Model: model, AgentMode: agentMode, Env: effectiveEnv}, nil
	case "claude":
		return &ClaudeAgent{Model: model, AgentMode: agentMode, Env: effectiveEnv}, nil
	case "cursor":
		return &CursorAgent{Model: model, AgentMode: agentMode, Env: effectiveEnv}, nil
	case "opencode":
		return &OpencodeAgent{Model: model, AgentMode: agentMode, Env: effectiveEnv}, nil
	case "codex":
		return &CodexAgent{Model: model, AgentMode: agentMode, Env: effectiveEnv}, nil
	case "copilot":
		return &CopilotAgent{Model: model, AgentMode: agentMode, Env: effectiveEnv}, nil
	case "antigravity", "agy":
		return &AntigravityAgent{Model: model, AgentMode: agentMode, Env: effectiveEnv}, nil
	default:
		return nil, fmt.Errorf("unknown agent %q (supported: %s)", agentName, strings.Join(SupportedAgents(), ", "))
	}
}
