package config

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/BurntSushi/toml"
)

// WriteConfig writes the given Config to the specified path in TOML format.
// It uses an atomic write pattern (write to temp file, then rename) to ensure data integrity.
func WriteConfig(path string, cfg *Config) error {
	// Create a temporary file in the same directory as the target file
	// to ensure we can rename it atomically (same filesystem).
	dir := filepath.Dir(path)
	const dirPerm = 0750
	if err := os.MkdirAll(dir, dirPerm); err != nil {
		return fmt.Errorf("failed to create directory %s: %w", dir, err)
	}

	tmpFile, err := os.CreateTemp(dir, "ralph-config-*.toml")
	if err != nil {
		return fmt.Errorf("failed to create temp file: %w", err)
	}
	defer func() {
		// Clean up temp file if rename fails
		_ = os.Remove(tmpFile.Name())
	}()

	// Encode config to TOML
	encoder := toml.NewEncoder(tmpFile)
	if err := encoder.Encode(cfg); err != nil {
		_ = tmpFile.Close()

		return fmt.Errorf("failed to encode config to TOML: %w", err)
	}

	// Close the file explicitly to flush writes
	if err := tmpFile.Close(); err != nil {
		return fmt.Errorf("failed to close temp file: %w", err)
	}

	// Rename temp file to target path (atomic operation)
	if err := os.Rename(tmpFile.Name(), path); err != nil {
		return fmt.Errorf("failed to rename temp file to %s: %w", path, err)
	}

	return nil
}
