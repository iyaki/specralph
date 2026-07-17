package main

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestRunHelpReturnsZero(t *testing.T) {
	var errBuf bytes.Buffer
	exitCode := run([]string{"--help"}, &errBuf)

	if exitCode != 0 {
		t.Fatalf("expected exit code 0, got %d", exitCode)
	}

	if errBuf.Len() != 0 {
		t.Fatalf("expected empty stderr, got %q", errBuf.String())
	}
}

func TestRunInvalidFlagReturnsOne(t *testing.T) {
	var errBuf bytes.Buffer
	exitCode := run([]string{"--unknown-flag"}, &errBuf)

	if exitCode != 1 {
		t.Fatalf("expected exit code 1, got %d", exitCode)
	}

	if !strings.Contains(errBuf.String(), "Error:") {
		t.Fatalf("expected error output, got %q", errBuf.String())
	}
}

func TestMainProcessHelpExitCode(t *testing.T) {
	cmd := exec.Command(os.Args[0], "-test.run=TestMainProcessHelper", "--", "--help")
	cmd.Env = append(os.Environ(), "GO_WANT_HELPER_PROCESS=1")
	err := cmd.Run()
	if err != nil {
		t.Fatalf("expected main helper process to exit 0, got: %v", err)
	}
}

func TestMainProcessHelper(t *testing.T) {
	t.Helper()

	if os.Getenv("GO_WANT_HELPER_PROCESS") != "1" {
		return
	}

	args := []string{}
	for i, arg := range os.Args {
		if arg == "--" {
			args = os.Args[i+1:]

			break
		}
	}
	os.Args = append([]string{"ralph-test"}, args...)
	main()
}

func TestReadmeDocumentsSpecralphRepoAndRalphCli(t *testing.T) {
	content, err := os.ReadFile(filepath.Join("..", "..", "README.md"))
	if err != nil {
		t.Fatalf("failed to read README: %v", err)
	}

	readme := string(content)
	if !strings.Contains(readme, "https://github.com/iyaki/specralph/releases") {
		t.Fatalf("expected README to point releases to iyaki/specralph")
	}
	if !strings.Contains(readme, "https://github.com/iyaki/specralph/releases/latest") {
		t.Fatalf("expected README to point latest release URL to iyaki/specralph")
	}
	if !strings.Contains(readme, "npx skills add https://github.com/iyaki/specralph/ --skill spec-creator") {
		t.Fatalf("expected README to point skill install to iyaki/specralph")
	}
	if !strings.Contains(readme, "The repository is `iyaki/specralph`, but the CLI command remains `ralph`.") {
		t.Fatalf("expected README to explain the repo and CLI naming split")
	}
	if !strings.Contains(readme, "<promise>COMPLETE</promise>") {
		t.Fatalf("expected README to document the completion signal")
	}
	if !strings.Contains(readme, "Creating Custom Prompts") {
		t.Fatal("expected README to include Creating Custom Prompts section")
	}
}

func TestGoModulePathUsesSpecralph(t *testing.T) {
	content, err := os.ReadFile(filepath.Join("..", "..", "go.mod"))
	if err != nil {
		t.Fatalf("failed to read go.mod: %v", err)
	}

	if !strings.Contains(string(content), "module github.com/iyaki/specralph") {
		t.Fatalf("expected go.mod to use module github.com/iyaki/specralph")
	}
}

func TestSkillsLockPointsToExternalRepos(t *testing.T) {
	content, err := os.ReadFile(filepath.Join("..", "..", "skills-lock.json"))
	if err != nil {
		t.Fatalf("failed to read skills-lock.json: %v", err)
	}

	expected := []string{
		"github/awesome-copilot",
		"iyaki/opencode-base-template",
		"anthropics/skills",
		"obra/superpowers",
	}
	for _, exp := range expected {
		if !strings.Contains(string(content), exp) {
			t.Fatalf("expected skills-lock.json to contain %q", exp)
		}
	}
}
