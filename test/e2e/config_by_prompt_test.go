package e2e_test

import "testing"

func TestE2EConfigByPromptFrontMatterAppliesAndIsStripped(t *testing.T) {
	runTestCase(t, TestCase{
		Name: "front matter overrides model and agent-mode",
		Args: []string{"--prompt-file", "prompt.md"},
		Env: map[string]string{
			"RALPH_TEST_AGENT_MODE": "complete_once",
			"RALPH_MODEL":           "",
			"RALPH_AGENT_MODE":      "",
		},
		Files: map[string]string{
			"prompt.md": `---
model: frontmatter-model
agent-mode: frontmatter-mode
---
# Prompt Body
Implement feature`,
		},
		ExpectedExitCode: 0,
		ExpectedStdoutContains: []string{
			"[ralph-test-agent] Args:",
			"--model",
			"frontmatter-model",
			"--agent",
			"frontmatter-mode",
			"# Prompt Body",
		},
		ForbiddenOutput: []string{
			"model: frontmatter-model",
			"agent-mode: frontmatter-mode",
		},
	})
}

func TestE2EConfigByPromptOverrideFromConfigApplies(t *testing.T) {
	runTestCase(t, TestCase{
		Name: "prompt-overrides section applies for build prompt",
		Args: []string{"build"},
		Env: map[string]string{
			"RALPH_TEST_AGENT_MODE": "complete_once",
			"RALPH_MODEL":           "",
			"RALPH_AGENT_MODE":      "",
		},
		Files: map[string]string{
			"ralph.toml": `model = "global-model"
agent-mode = "global-mode"

[prompt-overrides.build]
model = "override-model"
agent-mode = "override-mode"`,
		},
		ExpectedExitCode: 0,
		ExpectedStdoutContains: []string{
			"[ralph-test-agent] Args:",
			"--model",
			"override-model",
			"--agent",
			"override-mode",
		},
		ForbiddenOutput: []string{
			"global-model",
			"global-mode",
		},
	})
}

func TestE2EConfigByPromptEnvOverridesFrontMatter(t *testing.T) {
	runTestCase(t, TestCase{
		Name: "env values override front matter",
		Args: []string{"--prompt-file", "prompt.md"},
		Env: map[string]string{
			"RALPH_TEST_AGENT_MODE": "complete_once",
			"RALPH_MODEL":           "env-model",
			"RALPH_AGENT_MODE":      "env-mode",
		},
		Files: map[string]string{
			"prompt.md": `---
model: frontmatter-model
agent-mode: frontmatter-mode
---
Prompt body`,
		},
		ExpectedExitCode: 0,
		ExpectedStdoutContains: []string{
			"--model",
			"env-model",
			"--agent",
			"env-mode",
		},
		ForbiddenOutput: []string{
			"frontmatter-model",
			"frontmatter-mode",
		},
	})
}

func TestE2EConfigByPromptInvalidFrontMatterFailsBeforeAgentRun(t *testing.T) {
	runTestCase(t, TestCase{
		Name: "invalid front matter returns clear error",
		Args: []string{"--prompt-file", "prompt.md"},
		Files: map[string]string{
			"prompt.md": "---\nmodel: [\n---\nPrompt body",
		},
		ExpectedExitCode: 1,
		ExpectedStderrContains: []string{
			"failed to parse front matter",
		},
		ForbiddenOutput: []string{
			"[ralph-test-agent] Starting",
			"Starting Specralph - Max iterations",
		},
	})
}
