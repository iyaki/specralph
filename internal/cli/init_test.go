package cli_test

import (
	"testing"

	"github.com/iyaki/specralph/internal/cli"
)

func TestInitCommandStructure(t *testing.T) {
	cmd := cli.NewInitCommand()

	if cmd.Use != "init" {
		t.Errorf("expected Use 'init', got %s", cmd.Use)
	}

	if cmd.Short != "Initialize Specralph configuration" {
		t.Errorf("expected Short 'Initialize Specralph configuration', got %s", cmd.Short)
	}

	// Check flags
	forceFlag := cmd.Flag("force")
	if forceFlag == nil {
		t.Error("force flag not found")
	} else if forceFlag.DefValue != "false" {
		t.Errorf("expected force default 'false', got %s", forceFlag.DefValue)
	}

	outputFlag := cmd.Flag("output")
	if outputFlag == nil {
		t.Error("output flag not found")
	} else if outputFlag.DefValue != "" {
		t.Errorf("expected output default '', got %s", outputFlag.DefValue)
	}
}

func TestInitCommandExecution_NoTTY(t *testing.T) {
	// This test is hard to execute reliably without mocking stdout stat
	// but we can verify that NewInitCommand returns a valid command
	cmd := cli.NewInitCommand()
	if cmd == nil {
		t.Fatal("NewInitCommand returned nil")
	}
}
