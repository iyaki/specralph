package buildversion

import (
	"strings"
	"testing"
)

func TestGetReturnsVersionInfo(t *testing.T) {
	info := Get()

	if info.Version == "" {
		t.Error("expected Version to be non-empty")
	}
	if info.Commit == "" {
		t.Error("expected Commit to be non-empty")
	}
	if info.Date == "" {
		t.Error("expected Date to be non-empty")
	}
}

func TestStringReturnsFormattedVersion(t *testing.T) {
	result := String()

	if !strings.HasPrefix(result, "ralph v") {
		t.Errorf("expected to start with 'ralph v', got %q", result)
	}

	// Should not contain newlines in basic output
	if strings.Contains(result, "\n") {
		t.Errorf("expected single line output, got %q", result)
	}
}

func TestVersionInfoStruct(t *testing.T) {
	info := Get()

	// Verify struct fields match package variables
	if info.Version != "dev" && info.Version == "" {
		t.Error("Version should be non-empty")
	}
	if info.Commit != "unknown" && info.Commit == "" {
		t.Error("Commit should be non-empty")
	}
	if info.Date != "unknown" && info.Date == "" {
		t.Error("Date should be non-empty")
	}
}
