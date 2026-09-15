package cli_test

import (
	"bytes"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/iyaki/specralph/internal/cli"
)

func TestNewPromptsCommandBasicProperties(t *testing.T) {
	cmd := cli.NewPromptsCommand()
	if cmd.Use != "prompts" {
		t.Fatalf("unexpected use string: %q", cmd.Use)
	}
	if cmd.Short != "List and view available prompts" {
		t.Fatalf("unexpected short description: %q", cmd.Short)
	}

	cmd.SetArgs([]string{"--help"})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("help should execute successfully: %v", err)
	}
}

func TestNewPromptsListCommandBasicProperties(t *testing.T) {
	cmd := cli.NewPromptsListCommand()
	if cmd.Use != "list" {
		t.Fatalf("unexpected use string: %q", cmd.Use)
	}
	if cmd.Short != "List available prompts" {
		t.Fatalf("unexpected short description: %q", cmd.Short)
	}

	cmd.SetArgs([]string{"--help"})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("help should execute successfully: %v", err)
	}
}

func TestNewPromptsShowCommandBasicProperties(t *testing.T) {
	cmd := cli.NewPromptsShowCommand()
	if cmd.Use != "show <name>" {
		t.Fatalf("unexpected use string: %q", cmd.Use)
	}
	if cmd.Short != "Show full prompt content" {
		t.Fatalf("unexpected short description: %q", cmd.Short)
	}

	cmd.SetArgs([]string{"--help"})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("help should execute successfully: %v", err)
	}
}

func TestPromptsListShowsBuiltInPrompts(t *testing.T) {
	wd, err := os.Getwd()
	if err != nil {
		t.Fatalf("failed to get cwd: %v", err)
	}
	tmp := t.TempDir()
	if err := os.Chdir(tmp); err != nil {
		t.Fatalf("failed to chdir: %v", err)
	}
	t.Cleanup(func() {
		_ = os.Chdir(wd)
	})

	homeDir := t.TempDir()
	t.Setenv("HOME", homeDir)

	cmd := cli.NewPromptsListCommand()
	var out bytes.Buffer
	cmd.SetOut(&out)

	if err := cmd.Execute(); err != nil {
		t.Fatalf("expected execute success, got: %v", err)
	}

	output := out.String()
	if !strings.Contains(output, "Built-in Prompts:") {
		t.Errorf("expected output to contain 'Built-in Prompts:', got %q", output)
	}
	for _, want := range []string{
		"build-classic",
		"build-subagents",
		"plan",
		"Aliases:",
		"build -> build-classic (configurable via build-alias-prompt)",
	} {
		if !strings.Contains(output, want) {
			t.Errorf("expected output to contain %q, got %q", want, output)
		}
	}
	if !strings.Contains(output, "Use 'ralph run <prompt-name>' to execute a prompt.") {
		t.Errorf("expected usage hint, got %q", output)
	}
}

func TestPromptsListShowsCustomPrompts(t *testing.T) {
	wd, err := os.Getwd()
	if err != nil {
		t.Fatalf("failed to get cwd: %v", err)
	}
	tmp := t.TempDir()
	if err := os.Chdir(tmp); err != nil {
		t.Fatalf("failed to chdir: %v", err)
	}
	t.Cleanup(func() {
		_ = os.Chdir(wd)
	})

	homeDir := t.TempDir()
	t.Setenv("HOME", homeDir)

	// Create a custom prompts directory in HOME/.ralph
	promptsDir := filepath.Join(homeDir, ".ralph")
	if err := os.MkdirAll(promptsDir, 0755); err != nil {
		t.Fatalf("failed to create prompts dir: %v", err)
	}

	customPrompt := filepath.Join(promptsDir, "review.md")
	content := `# Code Review Prompt

This is a custom code review prompt.
It checks for security issues and performance.
`
	if err := os.WriteFile(customPrompt, []byte(content), 0644); err != nil {
		t.Fatalf("failed to write custom prompt: %v", err)
	}

	cmd := cli.NewPromptsListCommand()
	var out bytes.Buffer
	cmd.SetOut(&out)

	if err := cmd.Execute(); err != nil {
		t.Fatalf("expected execute success, got: %v", err)
	}

	output := out.String()
	if !strings.Contains(output, "Custom Prompts:") {
		t.Errorf("expected output to contain 'Custom Prompts:', got %q", output)
	}
	if !strings.Contains(output, "review") {
		t.Errorf("expected output to mention 'review' prompt, got %q", output)
	}
}

func TestPromptsListAliasFollowsConfig(t *testing.T) {
	wd, err := os.Getwd()
	if err != nil {
		t.Fatalf("failed to get cwd: %v", err)
	}
	tmp := t.TempDir()
	if err := os.Chdir(tmp); err != nil {
		t.Fatalf("failed to chdir: %v", err)
	}
	t.Cleanup(func() {
		_ = os.Chdir(wd)
	})

	homeDir := t.TempDir()
	t.Setenv("HOME", homeDir)
	t.Setenv("RALPH_BUILD_ALIAS_PROMPT", "build-subagents")

	cmd := cli.NewPromptsListCommand()
	var out bytes.Buffer
	cmd.SetOut(&out)

	if err := cmd.Execute(); err != nil {
		t.Fatalf("expected execute success, got: %v", err)
	}

	output := out.String()
	if !strings.Contains(output, "build -> build-subagents (configurable via build-alias-prompt)") {
		t.Errorf("expected alias line to follow config, got %q", output)
	}
}

func TestPromptsShowBuild(t *testing.T) {
	testPromptsShow(t, "build", "Agent Instructions (Build Mode)")
}

func TestPromptsShowPlan(t *testing.T) {
	testPromptsShow(t, "plan", "Agent Instructions (Planning Mode)")
}

func TestPromptsShowBuildSubagents(t *testing.T) {
	testPromptsShow(t, "build-subagents", "Agent Instructions (Build Mode with Subagents)")
}

func TestPromptsShowBuildFollowsAliasConfig(t *testing.T) {
	t.Setenv("RALPH_BUILD_ALIAS_PROMPT", "build-subagents")
	testPromptsShow(t, "build", "Agent Instructions (Build Mode with Subagents)")
}

func testPromptsShow(t *testing.T, promptName, expectedHeader string) {
	t.Helper()
	wd, err := os.Getwd()
	if err != nil {
		t.Fatalf("failed to get cwd: %v", err)
	}
	tmp := t.TempDir()
	if err := os.Chdir(tmp); err != nil {
		t.Fatalf("failed to chdir: %v", err)
	}
	t.Cleanup(func() {
		_ = os.Chdir(wd)
	})

	homeDir := t.TempDir()
	t.Setenv("HOME", homeDir)

	cmd := cli.NewPromptsShowCommand()
	cmd.SetArgs([]string{promptName})
	var out bytes.Buffer
	cmd.SetOut(&out)

	if err := cmd.Execute(); err != nil {
		t.Fatalf("expected execute success, got: %v", err)
	}

	output := out.String()
	if !strings.Contains(output, expectedHeader) {
		t.Errorf("expected output to contain %q, got %q", expectedHeader, output)
	}
	if !strings.Contains(output, "Study `specs/*`") {
		t.Errorf("expected output to mention studying specs, got %q", output)
	}
	if !strings.Contains(output, "IMPLEMENTATION_PLAN.md") {
		t.Errorf("expected output to mention implementation plan, got %q", output)
	}
}

func TestPromptsShowNonExistent(t *testing.T) {
	wd, err := os.Getwd()
	if err != nil {
		t.Fatalf("failed to get cwd: %v", err)
	}
	tmp := t.TempDir()
	if err := os.Chdir(tmp); err != nil {
		t.Fatalf("failed to chdir: %v", err)
	}
	t.Cleanup(func() {
		_ = os.Chdir(wd)
	})

	homeDir := t.TempDir()
	t.Setenv("HOME", homeDir)

	cmd := cli.NewPromptsShowCommand()
	cmd.SetArgs([]string{"nonexistent"})
	var out bytes.Buffer
	cmd.SetOut(&out)

	err = cmd.Execute()
	if err == nil {
		t.Fatal("expected error for nonexistent prompt, got nil")
	}
	if !strings.Contains(err.Error(), "not found") {
		t.Fatalf("expected 'not found' error, got: %v", err)
	}
}

func TestPromptsShowCustomPrompt(t *testing.T) {
	wd, err := os.Getwd()
	if err != nil {
		t.Fatalf("failed to get cwd: %v", err)
	}
	tmp := t.TempDir()
	if err := os.Chdir(tmp); err != nil {
		t.Fatalf("failed to chdir: %v", err)
	}
	t.Cleanup(func() {
		_ = os.Chdir(wd)
	})

	homeDir := t.TempDir()
	t.Setenv("HOME", homeDir)

	// Create a custom prompts directory in HOME/.ralph
	promptsDir := filepath.Join(homeDir, ".ralph")
	if err := os.MkdirAll(promptsDir, 0755); err != nil {
		t.Fatalf("failed to create prompts dir: %v", err)
	}

	customPrompt := filepath.Join(promptsDir, "custom.md")
	content := `# Custom Prompt

This is a custom prompt for testing.
`
	if err := os.WriteFile(customPrompt, []byte(content), 0644); err != nil {
		t.Fatalf("failed to write custom prompt: %v", err)
	}

	cmd := cli.NewPromptsShowCommand()
	cmd.SetArgs([]string{"custom"})
	var out bytes.Buffer
	cmd.SetOut(&out)

	if err := cmd.Execute(); err != nil {
		t.Fatalf("expected execute success, got: %v", err)
	}

	output := out.String()
	if !strings.Contains(output, "Custom Prompt") {
		t.Errorf("expected output to contain custom prompt title, got %q", output)
	}
	if !strings.Contains(output, "This is a custom prompt for testing.") {
		t.Errorf("expected output to contain custom prompt body, got %q", output)
	}
}

func TestPromptsShowCustomPromptWithFrontmatter(t *testing.T) {
	wd, err := os.Getwd()
	if err != nil {
		t.Fatalf("failed to get cwd: %v", err)
	}
	tmp := t.TempDir()
	if err := os.Chdir(tmp); err != nil {
		t.Fatalf("failed to chdir: %v", err)
	}
	t.Cleanup(func() {
		_ = os.Chdir(wd)
	})

	homeDir := t.TempDir()
	t.Setenv("HOME", homeDir)

	// Create a custom prompts directory in HOME/.ralph
	promptsDir := filepath.Join(homeDir, ".ralph")
	if err := os.MkdirAll(promptsDir, 0755); err != nil {
		t.Fatalf("failed to create prompts dir: %v", err)
	}

	customPrompt := filepath.Join(promptsDir, "with-fm.md")
	content := `---
model: default
agent-mode: task
---
# Prompt with Frontmatter

This content should be displayed without frontmatter.
`
	if err := os.WriteFile(customPrompt, []byte(content), 0644); err != nil {
		t.Fatalf("failed to write custom prompt: %v", err)
	}

	cmd := cli.NewPromptsShowCommand()
	cmd.SetArgs([]string{"with-fm"})
	var out bytes.Buffer
	cmd.SetOut(&out)

	if err := cmd.Execute(); err != nil {
		t.Fatalf("expected execute success, got: %v", err)
	}

	output := out.String()
	// Frontmatter should be stripped
	if strings.Contains(output, "model: default") {
		t.Errorf("expected frontmatter to be stripped, but found it in output: %q", output)
	}
	if strings.Contains(output, "agent-mode: task") {
		t.Errorf("expected frontmatter to be stripped, but found it in output: %q", output)
	}
	if !strings.Contains(output, "Prompt with Frontmatter") {
		t.Errorf("expected output to contain title, got %q", output)
	}
	if !strings.Contains(output, "This content should be displayed without frontmatter.") {
		t.Errorf("expected output to contain body, got %q", output)
	}
}

func TestPromptsValidateBuiltIn(t *testing.T) {
	for _, tc := range []struct{ input, resolved string }{
		{input: "build", resolved: "build-classic"},
		{input: "build-classic", resolved: "build-classic"},
		{input: "build-subagents", resolved: "build-subagents"},
		{input: "plan", resolved: "plan"},
	} {
		t.Run(tc.input, func(t *testing.T) {
			wd, err := os.Getwd()
			if err != nil {
				t.Fatalf("failed to get cwd: %v", err)
			}
			tmp := t.TempDir()
			if err := os.Chdir(tmp); err != nil {
				t.Fatalf("failed to chdir: %v", err)
			}
			t.Cleanup(func() {
				_ = os.Chdir(wd)
			})
			t.Setenv("HOME", t.TempDir())

			cmd := cli.NewPromptsValidateCommand()
			cmd.SetArgs([]string{tc.input})
			var out bytes.Buffer
			cmd.SetOut(&out)

			if err := cmd.Execute(); err != nil {
				t.Fatalf("expected built-in prompt to validate, got: %v", err)
			}

			output := out.String()
			for _, want := range []string{
				tc.resolved + "\n",
				"ok       resolve",
				"ok       frontmatter",
				"ok       description",
				"ok       completion-signal",
				"0 failed, 0 warnings",
			} {
				if !strings.Contains(output, want) {
					t.Errorf("expected output to contain %q, got %q", want, output)
				}
			}
		})
	}
}

func TestPromptsValidateFileTargetAllOk(t *testing.T) {
	wd, err := os.Getwd()
	if err != nil {
		t.Fatalf("failed to get cwd: %v", err)
	}
	tmp := t.TempDir()
	if err := os.Chdir(tmp); err != nil {
		t.Fatalf("failed to chdir: %v", err)
	}
	t.Cleanup(func() {
		_ = os.Chdir(wd)
	})
	t.Setenv("HOME", t.TempDir())

	target := filepath.Join(tmp, "review.md")
	content := "---\ndescription: Review the plan\n---\n# Review Prompt\n\n" +
		"Review the code until <promise>COMPLETE</promise>.\n"
	if err := os.WriteFile(target, []byte(content), 0644); err != nil {
		t.Fatalf("failed to write prompt file: %v", err)
	}

	cmd := cli.NewPromptsValidateCommand()
	cmd.SetArgs([]string{target})
	var out bytes.Buffer
	cmd.SetOut(&out)

	if err := cmd.Execute(); err != nil {
		t.Fatalf("expected valid file to pass, got: %v", err)
	}

	output := out.String()
	for _, want := range []string{
		target + "\n",
		"ok       resolve",
		"ok       completion-signal",
		"0 failed, 0 warnings",
	} {
		if !strings.Contains(output, want) {
			t.Errorf("expected output to contain %q, got %q", want, output)
		}
	}
}

func TestPromptsValidateNamedCustomPrompt(t *testing.T) {
	wd, err := os.Getwd()
	if err != nil {
		t.Fatalf("failed to get cwd: %v", err)
	}
	tmp := t.TempDir()
	if err := os.Chdir(tmp); err != nil {
		t.Fatalf("failed to chdir: %v", err)
	}
	t.Cleanup(func() {
		_ = os.Chdir(wd)
	})

	homeDir := t.TempDir()
	t.Setenv("HOME", homeDir)

	promptsDir := filepath.Join(homeDir, ".ralph")
	if err := os.MkdirAll(promptsDir, 0755); err != nil {
		t.Fatalf("failed to create prompts dir: %v", err)
	}
	promptFile := filepath.Join(promptsDir, "review.md")
	content := "# Review\n\nKeep going until <COMPLETION_SIGNAL>.\n"
	if err := os.WriteFile(promptFile, []byte(content), 0644); err != nil {
		t.Fatalf("failed to write prompt file: %v", err)
	}

	cmd := cli.NewPromptsValidateCommand()
	cmd.SetArgs([]string{"review"})
	var out bytes.Buffer
	cmd.SetOut(&out)

	if err := cmd.Execute(); err != nil {
		t.Fatalf("expected named prompt to validate, got: %v", err)
	}

	output := out.String()
	if !strings.Contains(output, promptFile+"\n") {
		t.Errorf("expected resolved path header, got %q", output)
	}
	if !strings.Contains(output, "ok       completion-signal") {
		t.Errorf("expected placeholder signal to pass, got %q", output)
	}
}

func TestPromptsValidateMissingSignalFails(t *testing.T) {
	wd, err := os.Getwd()
	if err != nil {
		t.Fatalf("failed to get cwd: %v", err)
	}
	tmp := t.TempDir()
	if err := os.Chdir(tmp); err != nil {
		t.Fatalf("failed to chdir: %v", err)
	}
	t.Cleanup(func() {
		_ = os.Chdir(wd)
	})
	t.Setenv("HOME", t.TempDir())

	target := filepath.Join(tmp, "review.md")
	content := "---\ndescription: Review prompt\n---\n# Review Prompt\n\nReview the code.\n"
	if err := os.WriteFile(target, []byte(content), 0644); err != nil {
		t.Fatalf("failed to write prompt file: %v", err)
	}

	cmd := cli.NewPromptsValidateCommand()
	cmd.SetArgs([]string{target})
	var out bytes.Buffer
	cmd.SetOut(&out)

	err = cmd.Execute()
	if !errors.Is(err, cli.ErrValidationFailed) {
		t.Fatalf("expected ErrValidationFailed, got: %v", err)
	}

	output := out.String()
	for _, want := range []string{
		"fail     completion-signal: neither <COMPLETION_SIGNAL> nor <promise>COMPLETE</promise> found",
		"1 failed, 0 warnings",
		"ok       description",
	} {
		if !strings.Contains(output, want) {
			t.Errorf("expected output to contain %q, got %q", want, output)
		}
	}
	if strings.Contains(output, "Review the code.") {
		t.Errorf("prompt content must not appear in output, got %q", output)
	}
}

func TestPromptsValidateUnbalancedFrontmatterWarns(t *testing.T) {
	wd, err := os.Getwd()
	if err != nil {
		t.Fatalf("failed to get cwd: %v", err)
	}
	tmp := t.TempDir()
	if err := os.Chdir(tmp); err != nil {
		t.Fatalf("failed to chdir: %v", err)
	}
	t.Cleanup(func() {
		_ = os.Chdir(wd)
	})
	t.Setenv("HOME", t.TempDir())

	target := filepath.Join(tmp, "review.md")
	content := "---\ndescription: Broken delimiters\n\n# Review\n\nUntil <promise>COMPLETE</promise>.\n"
	if err := os.WriteFile(target, []byte(content), 0644); err != nil {
		t.Fatalf("failed to write prompt file: %v", err)
	}

	cmd := cli.NewPromptsValidateCommand()
	cmd.SetArgs([]string{target})
	var out bytes.Buffer
	cmd.SetOut(&out)

	if err := cmd.Execute(); err != nil {
		t.Fatalf("expected warnings not to fail validation, got: %v", err)
	}

	output := out.String()
	for _, want := range []string{
		"warn     frontmatter: unbalanced frontmatter delimiters",
		"0 failed, 1 warning",
	} {
		if !strings.Contains(output, want) {
			t.Errorf("expected output to contain %q, got %q", want, output)
		}
	}
}

func TestPromptsValidateMissingFileTarget(t *testing.T) {
	wd, err := os.Getwd()
	if err != nil {
		t.Fatalf("failed to get cwd: %v", err)
	}
	tmp := t.TempDir()
	if err := os.Chdir(tmp); err != nil {
		t.Fatalf("failed to chdir: %v", err)
	}
	t.Cleanup(func() {
		_ = os.Chdir(wd)
	})
	t.Setenv("HOME", t.TempDir())

	target := filepath.Join(tmp, "missing.md")

	cmd := cli.NewPromptsValidateCommand()
	cmd.SetArgs([]string{target})
	var out bytes.Buffer
	cmd.SetOut(&out)

	err = cmd.Execute()
	if !errors.Is(err, cli.ErrValidationFailed) {
		t.Fatalf("expected ErrValidationFailed, got: %v", err)
	}

	output := out.String()
	if !strings.Contains(output, "fail     resolve") {
		t.Errorf("expected resolve check to fail, got %q", output)
	}
	if strings.Contains(output, "completion-signal") {
		t.Errorf("content checks must not run without content, got %q", output)
	}
}

func TestPromptsValidateUnknownNameErrors(t *testing.T) {
	wd, err := os.Getwd()
	if err != nil {
		t.Fatalf("failed to get cwd: %v", err)
	}
	tmp := t.TempDir()
	if err := os.Chdir(tmp); err != nil {
		t.Fatalf("failed to chdir: %v", err)
	}
	t.Cleanup(func() {
		_ = os.Chdir(wd)
	})
	t.Setenv("HOME", t.TempDir())

	cmd := cli.NewPromptsValidateCommand()
	cmd.SetArgs([]string{"nonexistent"})
	var out bytes.Buffer
	cmd.SetOut(&out)

	err = cmd.Execute()
	if err == nil || !strings.Contains(err.Error(), `prompt "nonexistent" not found`) {
		t.Fatalf("expected not-found error, got: %v", err)
	}
}

func TestPromptsGuidePrintsAuthoringContract(t *testing.T) {
	cmd := cli.NewPromptsGuideCommand()
	var out bytes.Buffer
	cmd.SetOut(&out)

	if err := cmd.Execute(); err != nil {
		t.Fatalf("expected guide to print without error, got: %v", err)
	}

	output := out.String()
	for _, want := range []string{
		"$HOME/.ralph",                // file location
		"ralph run <name>",            // invocation
		"description:",                // frontmatter convention
		"<COMPLETION_SIGNAL>",         // placeholder signal
		"<promise>COMPLETE</promise>", // literal signal
		"Objective",                   // recommended structure
		"Stop Condition",              // recommended structure
		"scope",                       // scope argument behavior
		"ralph skill install",         // distribution line
	} {
		if !strings.Contains(output, want) {
			t.Errorf("expected guide to contain %q, got %q", want, output)
		}
	}
}

func TestPromptsGuideRegistered(t *testing.T) {
	cmd := cli.NewPromptsCommand()
	for _, sub := range cmd.Commands() {
		if sub.Name() == "guide" {
			return
		}
	}

	t.Error("expected prompts command to register a guide subcommand")
}
