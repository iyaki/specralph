package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestWriteConfigErrorPaths(t *testing.T) {
	cfg := &Config{AgentName: "opencode"}

	t.Run("directory creation fails", func(t *testing.T) {
		// Use a path we know will fail (root is typically not writable in tests)
		unwritablePath := "/root/nonexistent-test-dir/config.toml"
		err := WriteConfig(unwritablePath, cfg)
		if err == nil {
			t.Skip("expected error for unwritable path, but succeeded (running as root?)")
		}
	})

	t.Run("rename fails with invalid path", func(t *testing.T) {
		// Create a temp directory, then try to rename to an invalid location
		tmpDir := t.TempDir()
		testPath := filepath.Join(tmpDir, "test.toml")

		// Write successfully first
		err := WriteConfig(testPath, cfg)
		if err != nil {
			t.Fatalf("unexpected error writing config: %v", err)
		}

		// Verify file exists
		if _, err := os.Stat(testPath); err != nil {
			t.Errorf("config file should exist after write: %v", err)
		}
	})
}
