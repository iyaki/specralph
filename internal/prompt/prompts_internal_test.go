package prompt

import (
	"os"
	"path/filepath"
	"testing"
)

func TestFindFileUpwardsInternal(t *testing.T) {
	t.Run("absolute path exists returns path", func(t *testing.T) {
		tmpDir := t.TempDir()
		testFile := filepath.Join(tmpDir, "test.txt")
		if err := os.WriteFile(testFile, []byte("test"), 0644); err != nil {
			t.Fatalf("failed to create test file: %v", err)
		}

		result := findFileUpwards(testFile)
		if result != testFile {
			t.Errorf("expected %q, got %q", testFile, result)
		}
	})

	t.Run("absolute path missing returns empty", func(t *testing.T) {
		result := findFileUpwards("/nonexistent/file.txt")
		if result != "" {
			t.Errorf("expected empty string, got %q", result)
		}
	})

	t.Run("relative file in cwd returns found path", func(t *testing.T) {
		tmpDir := t.TempDir()
		testFile := "test.txt"
		fullPath := filepath.Join(tmpDir, testFile)
		if err := os.WriteFile(fullPath, []byte("test"), 0644); err != nil {
			t.Fatalf("failed to create test file: %v", err)
		}

		origDir, _ := os.Getwd()
		if err := os.Chdir(tmpDir); err != nil {
			t.Fatal(err)
		}
		defer func() {
			if err := os.Chdir(origDir); err != nil {
				t.Fatal(err)
			}
		}()

		result := findFileUpwards(testFile)
		if result != fullPath {
			t.Errorf("expected %q, got %q", fullPath, result)
		}
	})

	t.Run("file nowhere returns empty", func(t *testing.T) {
		result := findFileUpwards("nonexistent-file-12345.txt")
		if result != "" {
			t.Errorf("expected empty string, got %q", result)
		}
	})
}
