package cli_test

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/iyaki/specralph/internal/cli"
)

func runPrecedenceTest(
	t *testing.T,
	configContent string,
	promptContent string,
	envVars map[string]string,
	cliArgs []string,
	expectedOutput string,
) {
	t.Helper()

	tmp := t.TempDir()
	writeExecutable(t, tmp, "opencode", "#!/bin/sh\necho \"ARGS: $*\"\necho \"<promise>COMPLETE</promise>\"\n")
	t.Setenv("PATH", tmp)

	configFile := filepath.Join(tmp, "ralph.toml")
	if err := os.WriteFile(configFile, []byte(configContent), 0644); err != nil {
		t.Fatalf("failed to write config file: %v", err)
	}

	promptFile := filepath.Join(tmp, "prompt.md")
	if err := os.WriteFile(promptFile, []byte(promptContent), 0644); err != nil {
		t.Fatalf("failed to write prompt file: %v", err)
	}

	for k, v := range envVars {
		t.Setenv(k, v)
	}

	cmd := cli.NewRalphCommand()
	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(&out)

	baseArgs := []string{
		"--config", configFile,
		"--prompt-file", promptFile,
		"--max-iterations", "1",
	}
	cmd.SetArgs(append(baseArgs, cliArgs...))

	if err := cmd.Execute(); err != nil {
		t.Fatalf("execute failed: %v", err)
	}

	output := out.String()
	if !strings.Contains(output, expectedOutput) {
		t.Errorf("expected output to contain args %q, got:\n%s", expectedOutput, output)
	}
}

func TestConfigPrecedence_FrontMatterOverridesGlobal(t *testing.T) {
	configContent := `
model = "global-model"
agent-mode = "global-mode"
`
	promptContent := `---
model: frontmatter-model
agent-mode: frontmatter-mode
---
# Task
`
	runPrecedenceTest(
		t,
		configContent,
		promptContent,
		nil,
		[]string{"build"},
		"--model frontmatter-model --agent frontmatter-mode",
	)
}

func TestConfigPrecedence_CLIOverridesAll(t *testing.T) {
	configContent := `
model = "global-model"
agent-mode = "global-mode"
[prompt-overrides.build]
model = "override-model"
agent-mode = "override-mode"
`
	promptContent := `---
model: frontmatter-model
agent-mode: frontmatter-mode
---
# Task
`
	envVars := map[string]string{
		"RALPH_MODEL":      "env-model",
		"RALPH_AGENT_MODE": "env-mode",
	}

	runPrecedenceTest(
		t,
		configContent,
		promptContent,
		envVars,
		[]string{
			"--model", "cli-model",
			"--agent-mode", "cli-mode",
			"build",
		},
		"--model cli-model --agent cli-mode",
	)
}

func TestConfigPrecedence_EnvOverridesFrontMatter(t *testing.T) {
	promptContent := `---
model: frontmatter-model
agent-mode: frontmatter-mode
---
# Task
`
	envVars := map[string]string{
		"RALPH_MODEL":      "env-model",
		"RALPH_AGENT_MODE": "env-mode",
	}

	runPrecedenceTest(
		t,
		"",
		promptContent,
		envVars,
		[]string{"build"},
		"--model env-model --agent env-mode",
	)
}

func TestConfigPrecedence_ConfigOverrideOverridesGlobal(t *testing.T) {
	configContent := `
model = "global-model"
agent-mode = "global-mode"
[prompt-overrides.build]
model = "override-model"
agent-mode = "override-mode"
`
	runPrecedenceTest(
		t,
		configContent,
		"# Task",
		nil,
		[]string{"build"},
		"--model override-model --agent override-mode",
	)
}

func TestConfigPrecedence_AgentEnvOverridesDoNotAffectModelOrAgentModePrecedence(t *testing.T) {
	configContent := `
model = "global-model"
agent-mode = "global-mode"

[prompt-overrides.build]
model = "config-override-model"
agent-mode = "config-override-mode"

[env]
RALPH_MODEL = "config-env-model"
RALPH_AGENT_MODE = "config-env-agent-mode"
`
	promptContent := `---
model: frontmatter-model
agent-mode: frontmatter-mode
---
# Task
`

	runPrecedenceTest(
		t,
		configContent,
		promptContent,
		nil,
		[]string{
			"--env", "RALPH_MODEL=cli-env-model",
			"--env", "RALPH_AGENT_MODE=cli-env-agent-mode",
			"build",
		},
		"--model frontmatter-model --agent frontmatter-mode",
	)
}
