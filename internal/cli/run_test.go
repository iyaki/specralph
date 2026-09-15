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

func TestNewRunCommandBasicProperties(t *testing.T) {
	cmd := cli.NewRunCommand()
	if !strings.Contains(cmd.Use, "run") {
		t.Fatalf("unexpected use string: %q", cmd.Use)
	}
	if cmd.Flags().Lookup("max-iterations") == nil {
		t.Fatal("expected max-iterations flag to exist")
	}

	cmd.SetArgs([]string{"--help"})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("help should execute successfully: %v", err)
	}
}

func TestRunCommandExecuteDebugHappyPath(t *testing.T) {
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

	cmd := cli.NewRunCommand()
	cmd.SetArgs([]string{"build-classic"})

	// Capture output
	var out bytes.Buffer
	cmd.SetOut(&out)

	if err := cmd.Execute(); err != nil {
		t.Fatalf("expected execute success in debug mode, got: %v", err)
	}

	output := out.String()
	if !strings.Contains(output, "[build-classic]") {
		t.Errorf("expected output to contain [build-classic], got %q", output)
	}
}

func TestRunMissingSignalWarning(t *testing.T) {
	const wantWarning = "warning: prompt file has no completion signal " +
		"(<COMPLETION_SIGNAL>); the loop will only stop at max iterations"

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

	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("DEBUG", "1")

	binDir := t.TempDir()
	writeExecutable(t, binDir, "opencode", "#!/bin/sh\necho \"ok\"\n")
	t.Setenv("PATH", binDir)

	promptsDir := filepath.Join(home, ".ralph")
	if err := os.MkdirAll(promptsDir, 0o755); err != nil {
		t.Fatalf("failed to create prompts dir: %v", err)
	}
	promptFile := filepath.Join(promptsDir, "review.md")
	noSignal := "Review the implementation plan.\nStop when done.\n"
	if err := os.WriteFile(promptFile, []byte(noSignal), 0o644); err != nil {
		t.Fatalf("failed to write prompt file: %v", err)
	}
	stdinFile, err := os.Create(filepath.Join(tmp, "stdin.md"))
	if err != nil {
		t.Fatalf("failed to create stdin file: %v", err)
	}
	if _, err := stdinFile.WriteString(noSignal); err != nil {
		t.Fatalf("failed to write stdin content: %v", err)
	}
	if _, err := stdinFile.Seek(0, 0); err != nil {
		t.Fatalf("failed to rewind stdin file: %v", err)
	}

	tests := []struct {
		name      string
		args      []string
		stdin     bool
		wantWarns bool
	}{
		{name: "prompt dir file without signal warns", args: []string{"review"}, wantWarns: true},
		{name: "explicit prompt file without signal warns", args: []string{"--prompt-file", promptFile}, wantWarns: true},
		{name: "inline prompt without signal is silent", args: []string{"--prompt", noSignal}},
		{name: "stdin prompt without signal is silent", args: []string{"-"}, stdin: true},
		{name: "built-in prompt is silent", args: []string{"build-classic"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assertRunWarningCase(t, tt.args, tt.wantWarns, wantWarning, stdinFile)
		})
	}
}

func assertRunWarningCase(t *testing.T, args []string, wantWarns bool, wantWarning string, stdinFile *os.File) {
	t.Helper()

	if stdinFile != nil {
		old := os.Stdin
		os.Stdin = stdinFile
		t.Cleanup(func() { os.Stdin = old })
	}

	cmd := cli.NewRunCommand()
	cmd.SetArgs(args)

	var out, errOut bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(&errOut)

	if err := cmd.Execute(); err != nil {
		t.Fatalf("expected execute success in debug mode, got: %v", err)
	}
	if !strings.Contains(out.String(), "All planned tasks completed") {
		t.Fatalf("expected run to proceed normally, got %q", out.String())
	}

	got := strings.Count(errOut.String(), wantWarning)
	if wantWarns && got != 1 {
		t.Errorf("expected warning exactly once on stderr, got %d in %q", got, errOut.String())
	}
	if !wantWarns && got != 0 {
		t.Errorf("expected no warning, got %d in %q", got, errOut.String())
	}
}

func TestRunCommandExecuteInitAsPrompt(t *testing.T) {
	// This tests that `ralph run init` treats "init" as a prompt name, not the init command
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

	cmd := cli.NewRunCommand()
	cmd.SetArgs([]string{"init"}) // "init" as prompt name

	// Capture output
	var out bytes.Buffer
	cmd.SetOut(&out)

	err = cmd.Execute()
	if err == nil {
		t.Fatal("expected error for missing 'init' prompt")
	}
	if !strings.Contains(err.Error(), "prompt file not found for 'init'") {
		t.Fatalf("expected prompt not found error, got: %v", err)
	}
}

func TestReadBoolFlagOverride(t *testing.T) {
	cmd := cli.NewRunCommand()
	cmd.Flags().Bool("test-flag", false, "")

	// Flag not changed
	override, err := cli.ReadBoolFlagOverride(cmd, "test-flag")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if override.Changed {
		t.Fatal("expected Changed to be false")
	}

	// Flag changed to true
	cmd.SetArgs([]string{"--test-flag=true"})
	if err := cmd.ParseFlags([]string{"--test-flag=true"}); err != nil {
		t.Fatalf("failed to parse flags: %v", err)
	}
	override, err = cli.ReadBoolFlagOverride(cmd, "test-flag")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !override.Changed || !override.Value {
		t.Fatal("expected Changed=true and Value=true")
	}
}

func TestReadEnvFlagOverrides(t *testing.T) {
	cmd := cli.NewRunCommand()

	// No env flags
	overrides, err := cli.ReadEnvFlagOverrides(cmd)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if overrides != nil {
		t.Fatal("expected nil overrides")
	}

	// With env flags
	cmd.SetArgs([]string{"--env", "KEY1=value1", "--env", "KEY2=value2"})
	if err := cmd.ParseFlags([]string{"--env", "KEY1=value1", "--env", "KEY2=value2"}); err != nil {
		t.Fatalf("failed to parse flags: %v", err)
	}
	overrides, err = cli.ReadEnvFlagOverrides(cmd)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if overrides["KEY1"] != "value1" || overrides["KEY2"] != "value2" {
		t.Fatalf("expected KEY1=value1, KEY2=value2, got %v", overrides)
	}
}

func TestReadEnvFlagOverridesInvalidEntry(t *testing.T) {
	cmd := cli.NewRunCommand()
	cmd.SetArgs([]string{"--env", "invalid-no-value"})
	if err := cmd.ParseFlags([]string{"--env", "invalid-no-value"}); err != nil {
		t.Fatalf("failed to parse flags: %v", err)
	}
	_, err := cli.ReadEnvFlagOverridesForTest(cmd)
	if err == nil {
		t.Fatal("expected error for invalid env entry")
	}
	if !strings.Contains(err.Error(), "expected KEY=VALUE") {
		t.Fatalf("expected KEY=VALUE error, got %v", err)
	}
}

func TestReadEnvFlagOverridesInvalidKey(t *testing.T) {
	cmd := cli.NewRunCommand()
	cmd.SetArgs([]string{"--env", "invalid-key=value"})
	if err := cmd.ParseFlags([]string{"--env", "invalid-key=value"}); err != nil {
		t.Fatalf("failed to parse flags: %v", err)
	}
	_, err := cli.ReadEnvFlagOverridesForTest(cmd)
	if err == nil {
		t.Fatal("expected error for invalid env key")
	}
	if !strings.Contains(err.Error(), "invalid --env key") {
		t.Fatalf("expected invalid key error, got %v", err)
	}
}

func TestRunLoopReachesMaxIterations(t *testing.T) {
	tmp := t.TempDir()
	binDir := t.TempDir()
	writeExecutable(t, binDir, "opencode", "#!/bin/sh\necho \"no completion signal\"\n")
	t.Setenv("PATH", binDir)
	t.Setenv("HOME", tmp)

	cfg := &config.Config{
		AgentName:     "opencode",
		Model:         "test-model",
		MaxIterations: 2,
		Env:           map[string]string{},
	}

	var out bytes.Buffer
	err := cli.RunLoop(cfg, "test-prompt", "test", &out)
	if err == nil || !strings.Contains(err.Error(), "max iterations reached") {
		t.Fatalf("expected max iterations error, got %v", err)
	}
}

func TestRunLoopDetectsCompletionSignal(t *testing.T) {
	tmp := t.TempDir()
	binDir := t.TempDir()
	writeExecutable(t, binDir, "opencode", "#!/bin/sh\necho \"<promise>COMPLETE</promise>\"\n")
	t.Setenv("PATH", binDir)
	t.Setenv("HOME", tmp)

	cfg := &config.Config{
		AgentName:     "opencode",
		Model:         "test-model",
		MaxIterations: 5,
		Env:           map[string]string{},
	}

	var out bytes.Buffer
	err := cli.RunLoop(cfg, "test-prompt", "test", &out)
	if err != nil {
		t.Fatalf("expected success with completion signal, got %v", err)
	}
}

func TestRunLoopAgentNotAvailable(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv("HOME", tmp)
	t.Setenv("PATH", tmp) // No agents available

	cfg := &config.Config{
		AgentName:     "opencode",
		Model:         "test-model",
		MaxIterations: 1,
		Env:           map[string]string{},
	}

	var out bytes.Buffer
	err := cli.RunLoop(cfg, "test-prompt", "test", &out)
	// Should complete but warn about agent not found
	output := out.String()
	if !strings.Contains(output, "Warning: opencode agent not found") {
		t.Fatalf("expected agent not found warning, got %q", output)
	}
	_ = err // Error expected due to max iterations
}

func TestHasCompletionSignal(t *testing.T) {
	tests := []struct {
		name     string
		result   string
		expected bool
	}{
		{"contains signal", "some text\n<promise>COMPLETE</promise>\nmore", true},
		{"no signal", "some text\nmore text", false},
		{"signal with spaces", "  <promise>COMPLETE</promise>  ", true},
		{"partial signal", "<promise>COMPLETE", false},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result := cli.HasCompletionSignal(tc.result, "<promise>COMPLETE</promise>")
			if result != tc.expected {
				t.Errorf("expected %v, got %v", tc.expected, result)
			}
		})
	}
}

var buildAliasRewriteCases = []struct {
	name        string
	configTOML  string
	promptsFile string
	args        []string
	wantOut     []string
	wantNoOut   []string
	wantErr     string
}{
	{
		name: "default alias resolves to build-classic",
		args: []string{"build"},
		wantOut: []string{
			"USING DEFAULT 'BUILD-CLASSIC' PROMPT",
			"[build-classic] Iteration 1",
			"Agent Instructions (Build Mode)",
		},
	},
	{
		name:       "config target build-subagents emits batch prompt",
		configTOML: "build-alias-prompt = \"build-subagents\"\n",
		args:       []string{"build"},
		wantOut: []string{
			"USING DEFAULT 'BUILD-SUBAGENTS' PROMPT",
			"[build-subagents] Iteration 1",
			"Agent Instructions (Build Mode with Subagents)",
			"at least one subagent per selected task",
		},
	},
	{
		name:       "flag overrides config alias target",
		configTOML: "build-alias-prompt = \"build-subagents\"\n",
		args:       []string{"--build-alias-prompt", "build-classic", "build"},
		wantOut: []string{
			"USING DEFAULT 'BUILD-CLASSIC' PROMPT",
			"[build-classic] Iteration 1",
		},
	},
	{
		name:        "escape hatch with prompts dir file keeps pre-alias behavior",
		configTOML:  "build-alias-prompt = \"build\"\n",
		promptsFile: "build.md",
		args:        []string{"build"},
		wantOut: []string{
			"USING PROMPT FILE",
			"[build] Iteration 1",
		},
	},
	{
		name:       "escape hatch without prompts dir file fails resolution",
		configTOML: "build-alias-prompt = \"build\"\n",
		args:       []string{"build"},
		wantErr:    "prompt file not found for 'build'",
	},
	{
		name:       "unknown alias target fails before loop start",
		configTOML: "build-alias-prompt = \"nope\"\n",
		args:       []string{"build"},
		wantErr:    "prompt file not found for 'nope'",
		wantNoOut:  []string{"Starting Specralph"},
	},
	{
		name:       "plan invocation is unaffected by alias config",
		configTOML: "build-alias-prompt = \"nope\"\n",
		args:       []string{"plan"},
		wantOut: []string{
			"USING DEFAULT 'PLAN' PROMPT",
			"[plan] Iteration 1",
		},
	},
}

func TestRunBuildAliasRewrite(t *testing.T) {
	for _, tt := range buildAliasRewriteCases {
		t.Run(tt.name, func(t *testing.T) {
			output, err := executeRunCommand(t, tt.args, tt.configTOML, tt.promptsFile)
			if tt.wantErr != "" {
				if err == nil {
					t.Fatalf("expected error containing %q, got success with output:\n%s", tt.wantErr, output)
				}
				if !strings.Contains(err.Error(), tt.wantErr) {
					t.Fatalf("expected error containing %q, got: %v", tt.wantErr, err)
				}
			} else if err != nil {
				t.Fatalf("expected execute success, got: %v", err)
			}

			for _, want := range tt.wantOut {
				if !strings.Contains(output, want) {
					t.Errorf("expected output to contain %q, got:\n%s", want, output)
				}
			}
			for _, forbidden := range tt.wantNoOut {
				if strings.Contains(output, forbidden) {
					t.Errorf("expected output NOT to contain %q, got:\n%s", forbidden, output)
				}
			}
		})
	}
}

// executeRunCommand runs `ralph <args>` in an isolated temp cwd with an empty
// HOME, an optional ralph.toml body, and an optional prompt file written to
// the default prompts dir ($HOME/.ralph/<name>). DEBUG mode keeps the loop to
// one iteration that echoes the resolved prompt; combined stdout/stderr is
// returned alongside the execute error.
func executeRunCommand(t *testing.T, args []string, configTOML, promptsFileName string) (string, error) {
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

	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("DEBUG", "1")

	binDir := t.TempDir()
	writeExecutable(t, binDir, "opencode", "#!/bin/sh\necho \"ok\"\n")
	t.Setenv("PATH", binDir)

	if configTOML != "" {
		if err := os.WriteFile(filepath.Join(tmp, "ralph.toml"), []byte(configTOML), 0o644); err != nil {
			t.Fatalf("failed to write config file: %v", err)
		}
	}
	if promptsFileName != "" {
		promptsDir := filepath.Join(home, ".ralph")
		if err := os.MkdirAll(promptsDir, 0o755); err != nil {
			t.Fatalf("failed to create prompts dir: %v", err)
		}
		content := "# Prompt from file\nReply with <COMPLETION_SIGNAL> when done.\n"
		if err := os.WriteFile(filepath.Join(promptsDir, promptsFileName), []byte(content), 0o644); err != nil {
			t.Fatalf("failed to write prompt file: %v", err)
		}
	}

	cmd := cli.NewRunCommand()
	cmd.SetArgs(args)

	var out, errOut bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(&errOut)

	err = cmd.Execute()

	return out.String() + errOut.String(), err
}
