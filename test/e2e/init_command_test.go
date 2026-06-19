package e2e_test

import "testing"

func TestE2EInitCommand(t *testing.T) {
	t.Run("InitWithoutTTYFailsFast", func(t *testing.T) {
		runTestCase(t, TestCase{
			Name:             "init requires interactive terminal",
			Args:             []string{"init"},
			ExpectedExitCode: 1,
			ExpectedStderrContains: []string{
				"ralph init requires an interactive terminal",
			},
		})
	})

	t.Run("InitPromptNameRunsViaRunSubcommand", func(t *testing.T) {
		runTestCase(t, TestCase{
			Name:             "run init executes prompt resolution path",
			Args:             []string{"run", "init"},
			ExpectedExitCode: 1,
			ExpectedStderrContains: []string{
				"prompt file not found for 'init'",
			},
			ForbiddenOutput: []string{
				"Initialized Specralph configuration",
				"ralph init requires an interactive terminal",
			},
		})
	})
}
