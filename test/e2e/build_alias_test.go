package e2e_test

import "testing"

// TestE2EBuildAliasDefaultResolvesBuildClassic pins the backwards-compat
// contract: a bare `build` invocation resolves to the build-classic built-in.
func TestE2EBuildAliasDefaultResolvesBuildClassic(t *testing.T) {
	t.Run("BareBuildUsesClassicBuiltin", func(t *testing.T) {
		runTestCase(t, TestCase{
			Name: "build alias defaults to build-classic",
			Args: []string{"build"},
			Env: map[string]string{
				"DEBUG": "1",
			},
			ExpectedExitCode: 0,
			ExpectedStdoutContains: []string{
				"USING DEFAULT 'BUILD-CLASSIC' PROMPT",
				"# Agent Instructions (Build Mode)",
				"- Study `specs/*` (including `specs/README.md` and related specs).",
				"[build-classic]",
			},
			ForbiddenOutput: []string{
				"# Agent Instructions (Build Mode with Subagents)",
				"BUILD-SUBAGENTS",
			},
		})
	})

	t.Run("ExplicitBuildClassicMatchesAliasOutput", func(t *testing.T) {
		runTestCase(t, TestCase{
			Name: "explicit build-classic resolves to the same built-in",
			Args: []string{"build-classic"},
			Env: map[string]string{
				"DEBUG": "1",
			},
			ExpectedExitCode: 0,
			ExpectedStdoutContains: []string{
				"USING DEFAULT 'BUILD-CLASSIC' PROMPT",
				"# Agent Instructions (Build Mode)",
				"[build-classic]",
			},
			ForbiddenOutput: []string{
				"# Agent Instructions (Build Mode with Subagents)",
			},
		})
	})
}

// TestE2EBuildAliasConfigTargetEmitsSubagentsPrompt verifies the opt-in path:
// `build-alias-prompt = "build-subagents"` in TOML routes `build` to the batch
// prompt with the spec-mandated markers.
func TestE2EBuildAliasConfigTargetEmitsSubagentsPrompt(t *testing.T) {
	runTestCase(t, TestCase{
		Name: "build-alias-prompt build-subagents emits batch prompt",
		Args: []string{"build"},
		Env: map[string]string{
			"DEBUG": "1",
		},
		Files: map[string]string{
			"ralph.toml": `build-alias-prompt = "build-subagents"`,
		},
		ExpectedExitCode: 0,
		ExpectedStdoutContains: []string{
			"USING DEFAULT 'BUILD-SUBAGENTS' PROMPT",
			"# Agent Instructions (Build Mode with Subagents)",
			"up to a maximum of 10",
			"at least one subagent per selected task",
			"[build-subagents]",
		},
		ForbiddenOutput: []string{
			"BUILD-CLASSIC",
			"# Agent Instructions (Build Mode)",
		},
	})
}

// TestE2EBuildAliasUnknownTargetFailsBeforeAgent verifies fail-fast behavior
// for an alias target that matches no built-in or prompt file.
func TestE2EBuildAliasUnknownTargetFailsBeforeAgent(t *testing.T) {
	runTestCase(t, TestCase{
		Name:             "unknown build-alias-prompt fails before agent execution",
		Args:             []string{"build"},
		ExpectedExitCode: 1,
		Files: map[string]string{
			"ralph.toml": `build-alias-prompt = "nope"`,
		},
		ExpectedStderrContains: []string{
			"failed to get prompt: prompt file not found for 'nope'",
		},
		ForbiddenOutput: []string{
			"[ralph-test-agent] Starting",
		},
	})
}

// TestE2EBuildAliasPlanUnaffectedByAliasTarget verifies the rewrite applies
// only to the literal `build` invocation name.
func TestE2EBuildAliasPlanUnaffectedByAliasTarget(t *testing.T) {
	runTestCase(t, TestCase{
		Name: "plan ignores build-alias-prompt",
		Args: []string{"plan"},
		Env: map[string]string{
			"DEBUG": "1",
		},
		Files: map[string]string{
			"ralph.toml": `build-alias-prompt = "nope"`,
		},
		ExpectedExitCode: 0,
		ExpectedStdoutContains: []string{
			"USING DEFAULT 'PLAN' PROMPT",
			"[plan]",
		},
		ForbiddenOutput: []string{
			"prompt file not found for 'nope'",
		},
	})
}

// TestE2EBuildAliasEscapeHatchUsesPromptFile pins the pre-alias escape hatch:
// `build-alias-prompt = "build"` skips the rewrite so PromptsDir/build.md is
// used when present.
func TestE2EBuildAliasEscapeHatchUsesPromptFile(t *testing.T) {
	runTestCase(t, TestCase{
		Name: "escape hatch build-alias-prompt build uses prompt file",
		Args: []string{"build"},
		Env: map[string]string{
			"DEBUG": "1",
		},
		Files: map[string]string{
			"ralph.toml": `build-alias-prompt = "build"
prompts-dir = ".ralph/prompts"`,
			".ralph/prompts/build.md": `# Legacy Build Override

Implement one task.
When done, output: <COMPLETION_SIGNAL>`,
		},
		ExpectedExitCode: 0,
		ExpectedStdoutContains: []string{
			"USING PROMPT FILE",
			"# Legacy Build Override",
			"[build]",
		},
		ForbiddenOutput: []string{
			"BUILD-CLASSIC",
			"BUILD-SUBAGENTS",
		},
	})
}
