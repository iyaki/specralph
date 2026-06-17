package cli_test

import (
	"bytes"
	"strings"
	"testing"

	"github.com/iyaki/ralphex/internal/cli"
)

func TestNewVersionCommandBasicProperties(t *testing.T) {
	cmd := cli.NewVersionCommand()

	if cmd.Use != "version" {
		t.Errorf("expected Use to be 'version', got %q", cmd.Use)
	}

	if cmd.Short == "" {
		t.Error("expected Short description to be set")
	}

	if cmd.Long == "" {
		t.Error("expected Long description to be set")
	}
}

func TestVersionCommandHelp(t *testing.T) {
	cmd := cli.NewVersionCommand()
	cmd.SetArgs([]string{"--help"})

	// Execute help - should not error
	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(&out)

	err := cmd.Execute()
	if err != nil {
		t.Errorf("expected no error with --help, got %v", err)
	}

	output := out.String()
	if !strings.Contains(output, "version") {
		t.Error("expected help to contain 'version'")
	}
}
