package cli_test

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/iyaki/specralph/internal/cli"
	"github.com/iyaki/specralph/internal/config"
)

func writeExecutable(t *testing.T, dir, name, content string) string {
	t.Helper()
	path := filepath.Join(dir, name)
	if err := os.WriteFile(path, []byte(content), 0755); err != nil {
		t.Fatalf("failed to write executable: %v", err)
	}

	return path
}

func TestNewRalphCommandBasicProperties(t *testing.T) {
	cmd := cli.NewRalphCommand()
	if cmd.Use != "ralph [options] [prompt] [scope]" {
		t.Fatalf("unexpected use string: %q", cmd.Use)
	}
	if !strings.Contains(cmd.Long, "Specralph") {
		t.Fatalf("expected long help to mention Specralph, got %q", cmd.Long)
	}
	if !strings.Contains(cmd.Long, "Ralph-Wiggum inspired spec-driven agent runner") {
		t.Fatalf("expected long help to include the new subtitle, got %q", cmd.Long)
	}
	if !strings.Contains(cmd.Long, "https://github.com/iyaki/specralph") {
		t.Fatalf("expected long help to point to the specralph repo, got %q", cmd.Long)
	}
	if cmd.Flags().Lookup("max-iterations") == nil {
		t.Fatal("expected max-iterations flag to exist")
	}

	cmd.SetArgs([]string{"--help"})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("help should execute successfully: %v", err)
	}
}

func TestNewRalphCommandExecuteDebugHappyPath(t *testing.T) {
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
	t.Setenv("DEBUG", "1")

	binDir := t.TempDir()
	writeExecutable(t, binDir, "opencode", "#!/bin/sh\necho \"ok\"\n")
	t.Setenv("PATH", binDir)

	cmd := cli.NewRalphCommand()
	cmd.SetArgs([]string{"build"})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("expected execute success in debug mode, got: %v", err)
	}
}

func TestNewRalphCommandExecuteConfigError(t *testing.T) {
	cmd := cli.NewRalphCommand()
	cmd.SetArgs([]string{"--config", "missing-config.toml", "build"})
	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected config loading error")
	}
	if !strings.Contains(err.Error(), "failed to load config") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestNewRalphCommandExecutePromptError(t *testing.T) {
	cmd := cli.NewRalphCommand()
	cmd.SetArgs([]string{"--prompt-file", "missing-prompt.md", "build"})
	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected prompt loading error")
	}
	if !strings.Contains(err.Error(), "failed to get prompt") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestRunLoopCompletesOnSignal(t *testing.T) {
	tmp := t.TempDir()
	writeExecutable(t, tmp, "opencode", "#!/bin/sh\necho \"$*\"\necho \"<promise>COMPLETE</promise>\"\n")
	t.Setenv("PATH", tmp)
	t.Setenv("DEBUG", "")

	cfg := &config.Config{MaxIterations: 3, AgentName: "opencode"}
	var out bytes.Buffer
	err := cli.RunLoop(cfg, "task <COMPLETION_SIGNAL>", "build", &out)
	if err != nil {
		t.Fatalf("expected completion success, got %v", err)
	}

	output := out.String()
	if !strings.Contains(output, "Starting Specralph - Max iterations: 3") {
		t.Fatalf("expected branded startup message, got %q", output)
	}
	if !strings.Contains(output, "All planned tasks completed!") {
		t.Fatalf("expected completion output, got %q", output)
	}
	if !strings.Contains(output, "<promise>COMPLETE</promise>") {
		t.Fatalf("expected replaced completion signal in agent input/output, got %q", output)
	}
}

func TestRunLoopDoesNotCompleteWhenSignalOnlyAppearsInEchoedPrompt(t *testing.T) {
	tmp := t.TempDir()
	writeExecutable(t, tmp, "opencode", "#!/bin/sh\necho \"$*\"\n")
	t.Setenv("PATH", tmp)
	t.Setenv("DEBUG", "")

	cfg := &config.Config{MaxIterations: 2, AgentName: "opencode"}
	var out bytes.Buffer
	err := cli.RunLoop(cfg, "task <COMPLETION_SIGNAL>", "build", &out)
	if err == nil {
		t.Fatal("expected max iterations error when completion signal is not explicitly returned")
	}
	if !strings.Contains(err.Error(), "max iterations reached") {
		t.Fatalf("unexpected error: %v", err)
	}

	output := out.String()
	if strings.Contains(output, "All planned tasks completed!") {
		t.Fatalf("expected no completion message when signal is only in echoed prompt, got %q", output)
	}
}

func TestRunLoopDebugMode(t *testing.T) {
	tmp := t.TempDir()
	writeExecutable(t, tmp, "opencode", "#!/bin/sh\necho \"should-not-run\"\n")
	t.Setenv("PATH", tmp)
	t.Setenv("DEBUG", "1")

	cfg := &config.Config{MaxIterations: 2, AgentName: "opencode"}
	var out bytes.Buffer
	err := cli.RunLoop(cfg, "hello <COMPLETION_SIGNAL>", "plan", &out)
	if err != nil {
		t.Fatalf("expected debug mode to finish without error, got %v", err)
	}

	output := out.String()
	if !strings.Contains(output, "hello <promise>COMPLETE</promise>") {
		t.Fatalf("expected prompt with replaced signal in debug output, got %q", output)
	}
}

func TestRunLoopWarnsWhenAgentUnavailable(t *testing.T) {
	t.Setenv("PATH", t.TempDir())
	t.Setenv("DEBUG", "1")

	cfg := &config.Config{MaxIterations: 1, AgentName: "opencode"}
	var out bytes.Buffer
	err := cli.RunLoop(cfg, "debug", "build", &out)
	if err != nil {
		t.Fatalf("expected debug mode success, got %v", err)
	}

	if !strings.Contains(out.String(), "agent not found in PATH") {
		t.Fatalf("expected unavailable-agent warning, got %q", out.String())
	}
}

func TestRunLoopAppliesEffectiveEnvOverridesToAgentProcess(t *testing.T) {
	tmp := t.TempDir()
	script := "#!/bin/sh\n" +
		"printf 'INHERITED_ONLY:%s\\n' \"$INHERITED_ONLY\"\n" +
		"printf 'OVERRIDE_ME:%s\\n' \"$OVERRIDE_ME\"\n" +
		"printf 'COMPLEX:%s\\n' \"$COMPLEX\"\n" +
		"printf '<promise>COMPLETE</promise>\\n'\n"
	writeExecutable(t, tmp, "opencode", script)
	t.Setenv("PATH", tmp)
	t.Setenv("DEBUG", "")
	t.Setenv("INHERITED_ONLY", "from-parent")
	t.Setenv("OVERRIDE_ME", "from-parent")

	cfg := &config.Config{
		MaxIterations: 1,
		AgentName:     "opencode",
		Env: map[string]string{
			"OVERRIDE_ME": "from-config",
			"COMPLEX":     "a=b=c",
		},
	}

	var out bytes.Buffer
	err := cli.RunLoop(cfg, "task", "build", &out)
	if err != nil {
		t.Fatalf("expected completion success, got %v", err)
	}

	output := out.String()
	if !strings.Contains(output, "INHERITED_ONLY:from-parent") {
		t.Fatalf("expected inherited environment variable in output, got %q", output)
	}
	if !strings.Contains(output, "OVERRIDE_ME:from-config") {
		t.Fatalf("expected config env override in output, got %q", output)
	}
	if !strings.Contains(output, "COMPLEX:a=b=c") {
		t.Fatalf("expected env value containing '=' to be preserved, got %q", output)
	}
}

func TestRunLoopRejectsInvalidAgentEnvKeyBeforeExecution(t *testing.T) {
	tmp := t.TempDir()
	writeExecutable(t, tmp, "opencode", "#!/bin/sh\nprintf 'agent-ran\\n'\nprintf '<promise>COMPLETE</promise>\\n'\n")
	t.Setenv("PATH", tmp)
	t.Setenv("DEBUG", "")

	cfg := &config.Config{
		MaxIterations: 1,
		AgentName:     "opencode",
		Env: map[string]string{
			"1INVALID": "super-secret-token",
		},
	}

	var out bytes.Buffer
	err := cli.RunLoop(cfg, "task", "build", &out)
	if err == nil {
		t.Fatal("expected invalid environment key error")
	}
	if !strings.Contains(err.Error(), "invalid environment key") {
		t.Fatalf("expected invalid environment key error, got %v", err)
	}
	if strings.Contains(err.Error(), "super-secret-token") {
		t.Fatalf("expected error to redact env value, got %v", err)
	}
	if strings.Contains(out.String(), "agent-ran") {
		t.Fatalf("expected agent process not to start, got %q", out.String())
	}
}

func TestRunLoopHandlesExecutionWarningAndMaxIterations(t *testing.T) {
	tmp := t.TempDir()
	writeExecutable(t, tmp, "opencode", "#!/bin/sh\necho \"partial\"\nexit 1\n")
	t.Setenv("PATH", tmp)
	t.Setenv("DEBUG", "")

	cfg := &config.Config{MaxIterations: 1, AgentName: "opencode"}
	var out bytes.Buffer
	err := cli.RunLoop(cfg, "task", "build", &out)
	if err == nil {
		t.Fatal("expected max iterations error")
	}
	if !strings.Contains(err.Error(), "max iterations reached") {
		t.Fatalf("unexpected error: %v", err)
	}

	output := out.String()
	if !strings.Contains(output, "Command execution warning") {
		t.Fatalf("expected execution warning, got %q", output)
	}
	if !strings.Contains(output, "Reached max iterations") {
		t.Fatalf("expected max iterations message, got %q", output)
	}
}

func TestRunLoopMaxIterationsWithoutCompletion(t *testing.T) {
	tmp := t.TempDir()
	writeExecutable(t, tmp, "opencode", "#!/bin/sh\necho \"working\"\n")
	t.Setenv("PATH", tmp)
	t.Setenv("DEBUG", "")

	cfg := &config.Config{MaxIterations: 2, AgentName: "opencode"}
	err := cli.RunLoop(cfg, "task", "build", &bytes.Buffer{})
	if err == nil {
		t.Fatal("expected max iterations error")
	}
}

func TestNewRalphCommandDefaultToBuild(t *testing.T) {
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
	t.Setenv("DEBUG", "1")

	binDir := t.TempDir()
	writeExecutable(t, binDir, "opencode", "#!/bin/sh\necho \"ok\"\n")
	t.Setenv("PATH", binDir)

	cmd := cli.NewRalphCommand()
	cmd.SetArgs([]string{}) // No args

	// Capture output
	var out bytes.Buffer
	cmd.SetOut(&out)

	if err := cmd.Execute(); err != nil {
		t.Fatalf("expected execute success in debug mode, got: %v", err)
	}

	output := out.String()
	if !strings.Contains(output, "[build]") {
		t.Errorf("expected output to contain [build] (default behavior), got %q", output)
	}
}

func TestNewRalphCommandInit(t *testing.T) {
	// Tests that `ralph init` executes the init command
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

	cmd := cli.NewRalphCommand()
	cmd.SetArgs([]string{"init"})

	// Capture output
	var out bytes.Buffer
	cmd.SetOut(&out)

	err = cmd.Execute()
	if err == nil {
		t.Fatal("expected error because init requires TTY")
	}

	if !strings.Contains(err.Error(), "ralph init requires an interactive terminal") {
		t.Errorf("expected error to contain 'ralph init requires an interactive terminal', got %q", err.Error())
	}
}
