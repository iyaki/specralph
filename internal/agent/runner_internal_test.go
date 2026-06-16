//nolint:cyclop
package agent

import (
	"bytes"
	"strings"
	"testing"
)

func TestCloneStringSlice(t *testing.T) {
	t.Run("nil input returns nil", func(t *testing.T) {
		var input []string
		result := cloneStringSlice(input)
		if result != nil {
			t.Errorf("expected nil, got %v", result)
		}
	})

	t.Run("empty slice returns empty slice not nil", func(t *testing.T) {
		input := []string{}
		result := cloneStringSlice(input)
		if result == nil {
			t.Error("expected empty slice, got nil")
		}
		if len(result) != 0 {
			t.Errorf("expected length 0, got %d", len(result))
		}
	})

	t.Run("slice with values returns independent copy", func(t *testing.T) {
		input := []string{"a", "b", "c"}
		result := cloneStringSlice(input)

		if len(result) != len(input) {
			t.Errorf("expected length %d, got %d", len(input), len(result))
		}

		for i := range input {
			if result[i] != input[i] {
				t.Errorf("index %d: expected %q, got %q", i, input[i], result[i])
			}
		}

		// Modify original, verify copy unchanged
		input[0] = "modified"
		if result[0] == "modified" {
			t.Error("result was not an independent copy")
		}
		if result[0] != "a" {
			t.Errorf("expected result[0] to still be \"a\", got %q", result[0])
		}
	})
}

func TestMapFromEnvironment(t *testing.T) {
	t.Run("empty entries returns empty map", func(t *testing.T) {
		var entries []string
		result := mapFromEnvironment(entries)
		if len(result) != 0 {
			t.Errorf("expected empty map, got %v", result)
		}
	})

	t.Run("valid KEY=value entries returns map", func(t *testing.T) {
		entries := []string{"FOO=bar", "BAZ=qux", "EMPTY="}
		result := mapFromEnvironment(entries)

		expected := map[string]string{
			"FOO":   "bar",
			"BAZ":   "qux",
			"EMPTY": "",
		}

		for k, v := range expected {
			if result[k] != v {
				t.Errorf("key %q: expected %q, got %q", k, v, result[k])
			}
		}
	})

	t.Run("entries without equals are skipped", func(t *testing.T) {
		entries := []string{"VALID=value", "INVALID_NO_EQUALS", "ANOTHER=valid"}
		result := mapFromEnvironment(entries)

		if _, exists := result["INVALID_NO_EQUALS"]; exists {
			t.Error("expected INVALID_NO_EQUALS to be skipped")
		}
		if result["VALID"] != "value" {
			t.Errorf("expected VALID=value, got %q", result["VALID"])
		}
		if result["ANOTHER"] != "valid" {
			t.Errorf("expected ANOTHER=valid, got %q", result["ANOTHER"])
		}
	})
}

func TestMapToEnvironment(t *testing.T) {
	t.Run("nil map returns nil slice", func(t *testing.T) {
		var input map[string]string
		result := mapToEnvironment(input)
		if result != nil {
			t.Errorf("expected nil, got %v", result)
		}
	})

	t.Run("empty map returns nil slice", func(t *testing.T) {
		input := map[string]string{}
		result := mapToEnvironment(input)
		if result != nil {
			t.Errorf("expected nil, got %v", result)
		}
	})

	t.Run("map with values returns sorted KEY=value strings", func(t *testing.T) {
		input := map[string]string{
			"ZEBRA": "last",
			"ALPHA": "first",
			"MID":   "middle",
		}
		result := mapToEnvironment(input)

		if len(result) != 3 {
			t.Errorf("expected 3 entries, got %d: %v", len(result), result)
		}

		// Verify lexicographic order
		expected := []string{
			"ALPHA=first",
			"MID=middle",
			"ZEBRA=last",
		}
		for i, exp := range expected {
			if result[i] != exp {
				t.Errorf("index %d: expected %q, got %q", i, exp, result[i])
			}
		}
	})
}

func TestExecuteAgentCommand(t *testing.T) {
	t.Run("successful command returns stdout and nil error", func(t *testing.T) {
		var buf bytes.Buffer
		output, err := executeAgentCommand("echo", []string{"test output"}, []string{}, &buf, "")

		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
		if strings.TrimSpace(output) != "test output" {
			t.Errorf("expected \"test output\", got %q", strings.TrimSpace(output))
		}
		if strings.TrimSpace(buf.String()) != "test output" {
			t.Errorf("buffer expected \"test output\", got %q", strings.TrimSpace(buf.String()))
		}
	})

	t.Run("command with stderr returns combined output", func(t *testing.T) {
		var buf bytes.Buffer
		// bash -c to produce both stdout and stderr
		cmd := []string{"-c", "echo stdout_msg; echo stderr_msg >&2"}
		output, err := executeAgentCommand("bash", cmd, []string{}, &buf, "")

		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
		// Combined output should contain both messages
		if !strings.Contains(output, "stdout_msg") {
			t.Errorf("expected output to contain \"stdout_msg\", got %q", output)
		}
		if !strings.Contains(output, "stderr_msg") {
			t.Errorf("expected output to contain \"stderr_msg\", got %q", output)
		}
	})

	t.Run("failing command returns output and error", func(t *testing.T) {
		var buf bytes.Buffer
		output, err := executeAgentCommand("false", []string{}, []string{}, &buf, "testprefix")

		if err == nil {
			t.Error("expected error, got nil")
		}
		if !strings.Contains(err.Error(), "execution failed") {
			t.Errorf("expected error to contain \"execution failed\", got %q", err.Error())
		}
		// output may be empty depending on command
		_ = output
	})

	t.Run("nil output writer uses io.Discard without panic", func(t *testing.T) {
		output, err := executeAgentCommand("echo", []string{"nil test"}, []string{}, nil, "")

		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
		if strings.TrimSpace(output) != "nil test" {
			t.Errorf("expected \"nil test\", got %q", strings.TrimSpace(output))
		}
	})
}

func TestIsAgentAvailable(t *testing.T) {
	t.Run("available command returns true", func(t *testing.T) {
		// echo is guaranteed to exist on any POSIX system
		result := isAgentAvailable("echo")
		if !result {
			t.Error("expected echo to be available")
		}
	})

	t.Run("nonexistent command returns false", func(t *testing.T) {
		// Use a guaranteed-nonexistent command name
		result := isAgentAvailable("nonexistent-ralph-test-agent-binary")
		if result {
			t.Error("expected nonexistent-ralph-test-agent-binary to be unavailable")
		}
	})
}
