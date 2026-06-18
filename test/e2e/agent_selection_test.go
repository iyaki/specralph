package e2e_test

import "testing"

var agentSelectionTestCases = []TestCase{
	{
		Name: "select claude agent",
		Args: []string{
			"--agent", "claude",
			"--model", "claude-sonnet-4",
			"--max-iterations", "1",
			"--prompt", "hello",
		},
		ExpectedExitCode: 0,
		ExpectedStdoutContains: []string{
			"Using agent: claude",
			"[ralph-test-agent] Args:",
			"--dangerously-skip-permissions",
			"--model",
			"claude-sonnet-4",
		},
	},
	{
		Name: "select cursor agent",
		Args: []string{
			"--agent", "cursor",
			"--max-iterations", "1",
			"--prompt", "hello",
		},
		ExpectedExitCode: 0,
		ExpectedStdoutContains: []string{
			"Using agent: cursor",
			"[ralph-test-agent] Args:",
		},
	},
	{
		Name: "select omp agent",
		Args: []string{
			"--agent", "omp",
			"--model", "m4",
			"--max-iterations", "1",
			"--prompt", "hello",
		},
		ExpectedExitCode: 0,
		ExpectedStdoutContains: []string{
			"Using agent: omp",
			"[ralph-test-agent] Args:",
			"--print",
			"--no-title",
			"--no-session",
			"--model",
			"m4",
		},
	},
	{
		Name: "unknown agent returns error",
		Args: []string{
			"--agent", "unknown-agent",
			"--max-iterations", "1",
			"--prompt", "hello",
		},
		ExpectedExitCode: 1,
		ExpectedStderrContains: []string{
			"unknown agent \"unknown-agent\"",
		},
	},
}

func TestE2EAgentSelection(t *testing.T) {
	for _, tc := range agentSelectionTestCases {
		t.Run(tc.Name, func(t *testing.T) {
			runTestCase(t, tc)
		})
	}
}
