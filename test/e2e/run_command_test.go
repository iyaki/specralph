package e2e_test

import "testing"

func TestE2ERunCommandRouting(t *testing.T) {
	t.Run("RunDefaultsToBuildPrompt", func(t *testing.T) {
		runTestCase(t, TestCase{
			Name: "Run command defaults to build",
			Args: []string{"run"},
			Env: map[string]string{
				"RALPH_TEST_AGENT_MODE": "complete_once",
			},
			ExpectedExitCode: 0,
			ExpectedStdoutContains: []string{
				"[build-classic]",
				"<promise>COMPLETE</promise>",
			},
		})
	})

	t.Run("InitSubcommandWinsCollision", func(t *testing.T) {
		runTestCase(t, TestCase{
			Name:             "init subcommand has priority over prompt alias",
			Args:             []string{"init"},
			ExpectedExitCode: 1,
			ExpectedStderrContains: []string{
				"ralph init requires an interactive terminal",
			},
		})
	})

	t.Run("RunInitTreatsInitAsPromptName", func(t *testing.T) {
		runTestCase(t, TestCase{
			Name:             "run init resolves prompt named init",
			Args:             []string{"run", "init"},
			ExpectedExitCode: 1,
			ExpectedStderrContains: []string{
				"prompt file not found for 'init'",
			},
			ForbiddenOutput: []string{
				"[ralph-test-agent] Starting",
			},
		})
	})
}
