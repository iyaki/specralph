package logger

import (
	"os"
	"testing"
)

func TestGetGitBranchInternal(t *testing.T) {
	t.Run("git command fails returns N/A", func(t *testing.T) {
		tmpDir := t.TempDir()
		origDir, _ := os.Getwd()
		if err := os.Chdir(tmpDir); err != nil {
			t.Fatal(err)
		}
		defer func() {
			if err := os.Chdir(origDir); err != nil {
				t.Fatal(err)
			}
		}()

		result := getGitBranch()
		if result != "N/A" {
			t.Errorf("expected \"N/A\" in non-git directory, got %q", result)
		}
	})
}

func TestGetGitCommitInternal(t *testing.T) {
	t.Run("git command fails returns N/A", func(t *testing.T) {
		tmpDir := t.TempDir()
		origDir, _ := os.Getwd()
		if err := os.Chdir(tmpDir); err != nil {
			t.Fatal(err)
		}
		defer func() {
			if err := os.Chdir(origDir); err != nil {
				t.Fatal(err)
			}
		}()

		result := getGitCommit()
		if result != "N/A" {
			t.Errorf("expected \"N/A\" in non-git directory, got %q", result)
		}
	})
}
