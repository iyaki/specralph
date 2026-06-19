package e2e_test

import (
	"testing"
)

func TestE2ECompletionFlow(t *testing.T) {
	tc := TestCase{
		Name: "Happy Path: Completion Detected",
		Args: []string{"--prompt-file", "prompt.txt"},
		Env: map[string]string{
			"RALPH_TEST_AGENT_MODE": "complete_once",
		},
		Files: map[string]string{
			"prompt.txt": "Just a simple prompt",
		},
		ExpectedExitCode: 0,
		ExpectedStdoutContains: []string{
			"<promise>COMPLETE</promise>",
		},
	}

	runTestCase(t, tc)
}

func TestE2EMaxIterations(t *testing.T) {
	tc := TestCase{
		Name: "Failure Path: Max Iterations Reached",
		Args: []string{"--prompt-file", "prompt.txt", "--max-iterations", "2"},
		Env: map[string]string{
			"RALPH_TEST_AGENT_MODE": "never_complete",
		},
		Files: map[string]string{
			"prompt.txt": "Just a simple prompt",
		},
		ExpectedExitCode: 1,
		ExpectedStderrContains: []string{
			"max iterations reached",
		},
	}

	runTestCase(t, tc)
}

func TestE2EReturnErrorPath(t *testing.T) {
	tc := TestCase{
		Name: "Failure Path: Agent Return Error",
		Args: []string{"--prompt", "Trigger return error path", "--max-iterations", "1"},
		Env: map[string]string{
			"RALPH_TEST_AGENT_MODE": "return_error",
		},
		ExpectedExitCode: 1,
		ExpectedStdoutContains: []string{
			"Simulated agent failure",
			"Command execution warning:",
		},
		ExpectedStderrContains: []string{
			"max iterations reached",
		},
		ForbiddenOutput: []string{
			"<promise>COMPLETE</promise>",
		},
	}

	runTestCase(t, tc)
}

func TestE2ESlowCompletePath(t *testing.T) {
	tc := TestCase{
		Name: "Happy Path: Slow Complete With Deterministic Delay",
		Args: []string{"--prompt", "Trigger slow complete path"},
		Env: map[string]string{
			"RALPH_TEST_AGENT_MODE": "slow_complete",
		},
		ExpectedExitCode: 0,
		ExpectedStdoutContains: []string{
			"Thinking...",
			"<promise>COMPLETE</promise>",
		},
		MinimumDurationMs: 80,
	}

	runTestCase(t, tc)
}

func TestE2EMissingPromptFile(t *testing.T) {
	tc := TestCase{
		Name: "Failure Path: Missing Prompt File",
		Args: []string{"--prompt-file", "non-existent-prompt.txt"},
		Env: map[string]string{
			"RALPH_TEST_AGENT_MODE": "never_complete",
		},
		ExpectedExitCode: 1,
		ExpectedStderrContains: []string{
			"failed to read prompt file",
		},
	}

	runTestCase(t, tc)
}

func TestE2ELogging(t *testing.T) {
	tc := TestCase{
		Name: "Logging: Enabled via env",
		Args: []string{"--log-file", "ralph.log", "--prompt-file", "prompt.txt"},
		Env: map[string]string{
			"RALPH_TEST_AGENT_MODE": "complete_once",
			"RALPH_LOG_ENABLED":     "1",
		},
		Files: map[string]string{
			"prompt.txt": "Just a simple prompt",
		},
		ExpectedExitCode: 0,
		ExpectedFiles: []string{
			"ralph.log",
		},
		ExpectedFileContent: map[string][]string{
			"ralph.log": {
				"===== Specralph run started at",
			},
		},
	}

	runTestCase(t, tc)
}
