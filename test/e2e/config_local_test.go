package e2e_test

import "testing"

func TestE2EConfigLocalOverlay_ConfigFlagUsesSiblingOverlay(t *testing.T) {
	runTestCase(t, TestCase{
		Name: "config local overlay with --config",
		Args: []string{"--config", "team/ralph.toml"},
		Files: map[string]string{
			"team/ralph.toml":       `max-iterations = 5`,
			"team/ralph-local.toml": `max-iterations = 9`,
			"ralph-local.toml":      `max-iterations = 99`,
		},
		ExpectedExitCode: 0,
		ExpectedStdoutContains: []string{
			"Max iterations: 9",
		},
		ForbiddenOutput: []string{
			"Max iterations: 99",
		},
	})
}

func TestE2EConfigLocalOverlay_RalphConfigEnvUsesSiblingOverlay(t *testing.T) {
	runTestCase(t, TestCase{
		Name: "config local overlay with RALPH_CONFIG",
		Env: map[string]string{
			"RALPH_CONFIG": "team/ralph.toml",
		},
		Files: map[string]string{
			"team/ralph.toml":       `max-iterations = 6`,
			"team/ralph-local.toml": `max-iterations = 10`,
			"ralph-local.toml":      `max-iterations = 99`,
		},
		ExpectedExitCode: 0,
		ExpectedStdoutContains: []string{
			"Max iterations: 10",
		},
		ForbiddenOutput: []string{
			"Max iterations: 99",
		},
	})
}

func TestE2EConfigLocalOverlay_DefaultDiscoveryUsesSiblingOverlay(t *testing.T) {
	runTestCase(t, TestCase{
		Name: "config local overlay with default discovery",
		Files: map[string]string{
			"ralph.toml":       `max-iterations = 4`,
			"ralph-local.toml": `max-iterations = 11`,
		},
		ExpectedExitCode: 0,
		ExpectedStdoutContains: []string{
			"Max iterations: 11",
		},
	})
}

func TestE2EConfigLocalOverlay_InvalidOverlayFailsBeforeAgentExecution(t *testing.T) {
	runTestCase(t, TestCase{
		Name: "invalid ralph-local.toml fails before agent run",
		Files: map[string]string{
			"ralph.toml":       `max-iterations = 4`,
			"ralph-local.toml": `max-iterations = "broken`,
		},
		ExpectedExitCode: 1,
		ExpectedStderrContains: []string{
			"failed to load overlay config file",
		},
		ForbiddenOutput: []string{
			"[ralph-test-agent] Starting",
		},
	})
}

func TestE2EConfigLocalOverlay_PromptOverridesDeepMerge(t *testing.T) {
	runTestCase(t, TestCase{
		Name: "prompt-overrides deep merge across base and local",
		Args: []string{"build"},
		Env: map[string]string{
			"RALPH_MODEL":      "",
			"RALPH_AGENT_MODE": "",
		},
		Files: map[string]string{
			"ralph.toml": `[prompt-overrides.build-classic]
model = "base-model"`,
			"ralph-local.toml": `[prompt-overrides.build-classic]
agent-mode = "overlay-mode"`,
		},
		ExpectedExitCode: 0,
		ExpectedStdoutContains: []string{
			"--model",
			"base-model",
			"--agent",
			"overlay-mode",
		},
	})
}
