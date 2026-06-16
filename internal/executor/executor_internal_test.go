package executor

import (
	"strings"
	"testing"
)

func TestExecuteCommandWithNilOutput(t *testing.T) {
	output, err := ExecuteCommand("echo", []string{"test"}, nil)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if strings.TrimSpace(output) != "test" {
		t.Errorf("expected \"test\" output, got %q", output)
	}
}

func TestExecuteCommandCombinedOutput(t *testing.T) {
	output, err := ExecuteCommand("sh", []string{"-c", "echo stdout; echo stderr >&2"}, nil)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if !strings.Contains(output, "stdout") {
		t.Errorf("expected output to contain \"stdout\", got %q", output)
	}
	if !strings.Contains(output, "stderr") {
		t.Errorf("expected output to contain \"stderr\", got %q", output)
	}
}
