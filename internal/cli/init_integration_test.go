package cli_test

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/iyaki/specralph/internal/cli"
	"github.com/spf13/cobra"
)

// TestExecuteInitCommand_FullSuccessPath tests the complete happy path.
func TestExecuteInitCommand_FullSuccessPath(t *testing.T) {
	// Override terminal check
	original := cli.GetIsInteractiveTerminalForTest()
	cli.SetIsInteractiveTerminalForTest(true)
	t.Cleanup(func() { cli.SetIsInteractiveTerminalForTest(original) })

	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "ralph.toml")

	// All defaults, confirm write
	input := strings.Repeat("\n", 10) + "yes\n"

	cmd := &cobra.Command{}
	cmd.SetIn(strings.NewReader(input))
	output := &bytes.Buffer{}
	cmd.SetOut(output)

	err := cli.ExecuteInitCommandForTest(cmd, configPath, false)
	if err != nil {
		t.Fatalf("executeInitCommand failed: %v", err)
	}

	// Verify success
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		t.Fatal("config file should exist")
	}
	if !strings.Contains(output.String(), "Initialized Specralph configuration") {
		t.Error("expected success message")
	}
}

// TestExecuteInitCommand_NewInitSessionError tests when newInitSession fails (non-interactive).
func TestExecuteInitCommand_NewInitSessionError(t *testing.T) {
	cli.SetIsInteractiveTerminalForTest(false)
	t.Cleanup(func() { cli.SetIsInteractiveTerminalForTest(true) })

	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "ralph.toml")

	cmd := &cobra.Command{}
	cmd.SetIn(strings.NewReader(""))
	cmd.SetOut(&bytes.Buffer{})

	err := cli.ExecuteInitCommandForTest(cmd, configPath, false)
	if err == nil {
		t.Fatal("expected error for non-interactive terminal")
	}
	if !strings.Contains(err.Error(), "interactive terminal") {
		t.Errorf("expected interactive terminal error, got: %v", err)
	}
}

// TestExecuteInitCommand_PrepareInitSessionDecline tests declining to overwrite.
func TestExecuteInitCommand_PrepareInitSessionDecline(t *testing.T) {
	cli.SetIsInteractiveTerminalForTest(true)

	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "ralph.toml")

	// Create existing config
	if err := os.WriteFile(configPath, []byte(`agent = "opencode"`), 0644); err != nil {
		t.Fatalf("failed to create config: %v", err)
	}

	cmd := &cobra.Command{}
	// Decline overwrite
	cmd.SetIn(strings.NewReader("no\n"))
	output := &bytes.Buffer{}
	cmd.SetOut(output)

	err := cli.ExecuteInitCommandForTest(cmd, configPath, false)
	if err != nil {
		t.Fatalf("executeInitCommand failed: %v", err)
	}

	// Should exit early without running questionnaire
	if strings.Contains(output.String(), "AI agent") {
		t.Error("should not have shown questions")
	}
}

// TestExecuteInitCommand_RunInitQuestionnaireError tests questionnaire failure.
func TestExecuteInitCommand_RunInitQuestionnaireError(t *testing.T) {
	cli.SetIsInteractiveTerminalForTest(true)

	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "ralph.toml")

	cmd := &cobra.Command{}
	// EOF immediately during questions
	cmd.SetIn(strings.NewReader(""))
	output := &bytes.Buffer{}
	cmd.SetOut(output)

	err := cli.ExecuteInitCommandForTest(cmd, configPath, false)
	if err == nil {
		t.Fatal("expected error from questionnaire")
	}
	if !strings.Contains(err.Error(), "unexpected end of input") {
		t.Errorf("expected EOF error, got: %v", err)
	}
}

// TestExecuteInitCommand_ConfirmInitWriteDecline tests declining to write.
func TestExecuteInitCommand_ConfirmInitWriteDecline(t *testing.T) {
	cli.SetIsInteractiveTerminalForTest(true)

	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "ralph.toml")

	cmd := &cobra.Command{}
	// Answer all questions, then decline write
	input := strings.Repeat("\n", 9) + "no\n"
	cmd.SetIn(strings.NewReader(input))
	output := &bytes.Buffer{}
	cmd.SetOut(output)

	err := cli.ExecuteInitCommandForTest(cmd, configPath, false)
	if err != nil {
		t.Fatalf("executeInitCommand failed: %v", err)
	}

	// Config should NOT be created
	if _, err := os.Stat(configPath); !os.IsNotExist(err) {
		t.Fatal("config file should not exist")
	}
	if !strings.Contains(output.String(), "cancelled") {
		t.Error("expected cancellation message")
	}
}

// TestExecuteInitCommand_WriteInitConfigError tests write failure with read-only directory.
func TestExecuteInitCommand_WriteInitConfigError(t *testing.T) {
	if os.Geteuid() == 0 {
		t.Skip("skipping when running as root")
	}
	cli.SetIsInteractiveTerminalForTest(true)

	tmpDir := t.TempDir()
	// Create a read-only subdirectory
	readOnlyDir := filepath.Join(tmpDir, "readonly")
	if err := os.MkdirAll(readOnlyDir, 0555); err != nil {
		t.Fatalf("failed to create dir: %v", err)
	}
	configPath := filepath.Join(readOnlyDir, "ralph.toml")

	cmd := &cobra.Command{}
	input := strings.Repeat("\n", 10) + "yes\n"
	cmd.SetIn(strings.NewReader(input))
	output := &bytes.Buffer{}
	cmd.SetOut(output)

	err := cli.ExecuteInitCommandForTest(cmd, configPath, false)
	if err == nil {
		t.Fatal("expected error for read-only directory")
	}
	if !strings.Contains(err.Error(), "failed to write") {
		t.Logf("got error: %v", err)
	}
}
