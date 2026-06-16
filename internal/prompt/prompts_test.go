package prompt_test

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/iyaki/ralphex/internal/config"
	"github.com/iyaki/ralphex/internal/prompt"
)

func TestBuildPromptIncludesConfiguredReferences(t *testing.T) {
	cfg := &config.Config{
		SpecsDir:               "specs",
		SpecsIndexFile:         "README.md",
		ImplementationPlanName: "PLAN.md",
	}
	p := prompt.BuildPrompt(cfg)
	if !strings.Contains(p, "specs/README.md") {
		t.Fatalf("expected specs index reference, got %q", p)
	}
	if !strings.Contains(p, "PLAN.md") {
		t.Fatalf("expected implementation plan name, got %q", p)
	}
}

func TestPlanPromptIncludesScopeAndPlanName(t *testing.T) {
	cfg := &config.Config{ImplementationPlanName: "PLAN.md", SpecsDir: "specs"}
	p := prompt.PlanPrompt(cfg, "API")
	if !strings.Contains(p, "Scope: API") {
		t.Fatalf("expected scope in prompt, got %q", p)
	}
	if !strings.Contains(p, "PLAN.md") {
		t.Fatalf("expected plan name in prompt, got %q", p)
	}
}

func TestGetPromptCustomPrompt(t *testing.T) {
	cfg := &config.Config{CustomPrompt: "inline custom"}
	var out bytes.Buffer
	p, _, err := prompt.GetPrompt(cfg, "build", "scope", &out)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if p != "inline custom" {
		t.Fatalf("expected custom prompt, got %q", p)
	}
}

func TestGetPromptFromStdin(t *testing.T) {
	cfg := &config.Config{PromptFile: "-"}
	var out bytes.Buffer

	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("pipe creation failed: %v", err)
	}
	if _, err := w.Write([]byte("from-stdin")); err != nil {
		t.Fatalf("pipe write failed: %v", err)
	}
	_ = w.Close()

	oldStdin := os.Stdin
	os.Stdin = r
	t.Cleanup(func() {
		os.Stdin = oldStdin
	})

	p, _, err := prompt.GetPrompt(cfg, "build", "scope", &out)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if p != "from-stdin" {
		t.Fatalf("expected stdin content, got %q", p)
	}
}

func TestGetPromptFromFile(t *testing.T) {
	dir := t.TempDir()
	file := filepath.Join(dir, "custom.md")
	if err := os.WriteFile(file, []byte("from-file"), 0o644); err != nil {
		t.Fatalf("failed writing prompt file: %v", err)
	}

	cfg := &config.Config{PromptFile: file}
	p, _, err := prompt.GetPrompt(cfg, "build", "scope", &bytes.Buffer{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if p != "from-file" {
		t.Fatalf("expected file content, got %q", p)
	}
}

func TestGetPromptFromPromptsDir(t *testing.T) {
	dir := t.TempDir()
	promptsDir := filepath.Join(dir, "prompts")
	if err := os.MkdirAll(promptsDir, 0o755); err != nil {
		t.Fatalf("failed to create prompts dir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(promptsDir, "custom.md"), []byte("from-prompts-dir"), 0o644); err != nil {
		t.Fatalf("failed to write prompt: %v", err)
	}

	cfg := &config.Config{PromptsDir: promptsDir}
	p, _, err := prompt.GetPrompt(cfg, "custom", "scope", &bytes.Buffer{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if p != "from-prompts-dir" {
		t.Fatalf("expected prompts dir content, got %q", p)
	}
}

func TestGetPromptDefaultBuildAndPlan(t *testing.T) {
	cfg := &config.Config{SpecsDir: "specs", SpecsIndexFile: "README.md", ImplementationPlanName: "PLAN.md"}

	buildPrompt, _, err := prompt.GetPrompt(cfg, "build", "scope", &bytes.Buffer{})
	if err != nil {
		t.Fatalf("unexpected error for build: %v", err)
	}
	if !strings.Contains(buildPrompt, "Agent Instructions (Build Mode)") {
		t.Fatalf("unexpected build prompt: %q", buildPrompt)
	}

	planPrompt, _, err := prompt.GetPrompt(cfg, "plan", "My Scope", &bytes.Buffer{})
	if err != nil {
		t.Fatalf("unexpected error for plan: %v", err)
	}
	if !strings.Contains(planPrompt, "Scope: My Scope") {
		t.Fatalf("unexpected plan prompt: %q", planPrompt)
	}
}

func TestGetPromptUnknownReturnsError(t *testing.T) {
	cfg := &config.Config{PromptsDir: t.TempDir()}
	_, _, err := prompt.GetPrompt(cfg, "unknown", "scope", &bytes.Buffer{})
	if err == nil {
		t.Fatal("expected error for unknown prompt")
	}
}

func TestGetPromptWithFrontMatter(t *testing.T) {
	// Setup a temporary directory with a prompt file containing front matter
	dir := t.TempDir()
	promptFile := filepath.Join(dir, "override.md")
	content := []byte(`---
model: gpt-5-preview
agent-mode: architect
---
# Actual Prompt
Do something.`)
	if err := os.WriteFile(promptFile, content, 0o644); err != nil {
		t.Fatalf("failed to write prompt file: %v", err)
	}

	cfg := &config.Config{PromptFile: promptFile}
	var out bytes.Buffer

	// Call GetPrompt
	promptText, override, err := prompt.GetPrompt(cfg, "override", "scope", &out)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Verify the prompt text is stripped of front matter
	expectedPrompt := "# Actual Prompt\nDo something."
	if promptText != expectedPrompt {
		t.Errorf("expected stripped prompt %q, got %q", expectedPrompt, promptText)
	}

	// Verify the overrides are returned
	if override == nil {
		t.Fatal("expected override to be non-nil")
	}
	if override.Model != "gpt-5-preview" {
		t.Errorf("expected model override 'gpt-5-preview', got %q", override.Model)
	}
	if override.AgentMode != "architect" {
		t.Errorf("expected agent-mode override 'architect', got %q", override.AgentMode)
	}
}

func TestGetPromptFromDirWithFrontMatter(t *testing.T) {
	// Setup a temporary prompts directory
	dir := t.TempDir()
	promptsDir := filepath.Join(dir, "prompts")
	if err := os.MkdirAll(promptsDir, 0o755); err != nil {
		t.Fatalf("failed to create prompts dir: %v", err)
	}

	promptFile := filepath.Join(promptsDir, "my-task.md")
	content := []byte(`---
model: claude-3-opus
---
Task description`)
	if err := os.WriteFile(promptFile, content, 0o644); err != nil {
		t.Fatalf("failed to write prompt file: %v", err)
	}

	cfg := &config.Config{PromptsDir: promptsDir}
	var out bytes.Buffer

	// Call GetPrompt
	promptText, override, err := prompt.GetPrompt(cfg, "my-task", "scope", &out)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if promptText != "Task description" {
		t.Errorf("expected stripped prompt 'Task description', got %q", promptText)
	}

	if override == nil {
		t.Fatal("expected override to be non-nil")
	}
	if override.Model != "claude-3-opus" {
		t.Errorf("expected model override 'claude-3-opus', got %q", override.Model)
	}
	// AgentMode should be empty
	if override.AgentMode != "" {
		t.Errorf("expected empty agent-mode, got %q", override.AgentMode)
	}
}

func TestGetPromptNoFrontMatter(t *testing.T) {
	dir := t.TempDir()
	promptFile := filepath.Join(dir, "simple.md")
	content := []byte("Just a simple prompt")
	if err := os.WriteFile(promptFile, content, 0o644); err != nil {
		t.Fatalf("failed to write prompt file: %v", err)
	}

	cfg := &config.Config{PromptFile: promptFile}
	var out bytes.Buffer

	promptText, override, err := prompt.GetPrompt(cfg, "simple", "scope", &out)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if promptText != "Just a simple prompt" {
		t.Errorf("expected prompt text 'Just a simple prompt', got %q", promptText)
	}

	// Override should be nil or empty (depending on implementation choice, but nil is cleaner)
	// The implementation might return an empty struct. Let's check for emptiness.
	if override != nil && (override.Model != "" || override.AgentMode != "") {
		t.Errorf("expected empty overrides, got %v", override)
	}
}
