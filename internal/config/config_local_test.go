package config_test

import (
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/iyaki/specralph/internal/config"
)

func TestLoadConfigWithOverlayScalars(t *testing.T) {
	dir := t.TempDir()
	configFile := filepath.Join(dir, "ralph.toml")
	overlayFile := filepath.Join(dir, "ralph-local.toml")

	baseContent := `
model = "gpt-4"
agent = "opencode"
max-iterations = 10
`
	if err := os.WriteFile(configFile, []byte(baseContent), 0644); err != nil {
		t.Fatalf("failed to write base config: %v", err)
	}

	overlayContent := `
model = "claude-3-opus"
max-iterations = 20
`
	if err := os.WriteFile(overlayFile, []byte(overlayContent), 0644); err != nil {
		t.Fatalf("failed to write overlay config: %v", err)
	}

	c := &config.Config{
		ConfigFile: configFile,
	}

	if err := c.LoadConfig(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// scalar overrides
	if c.Model != "claude-3-opus" {
		t.Errorf("expected model 'claude-3-opus' (from overlay), got %q", c.Model)
	}
	if c.MaxIterations != 20 {
		t.Errorf("expected max-iterations 20 (from overlay), got %d", c.MaxIterations)
	}
	// base value preserved
	if c.AgentName != "opencode" {
		t.Errorf("expected agent 'opencode' (from base), got %q", c.AgentName)
	}
}

func TestLoadConfigWithOverlayPrompts(t *testing.T) {
	dir := t.TempDir()
	configFile := filepath.Join(dir, "ralph.toml")
	overlayFile := filepath.Join(dir, "ralph-local.toml")

	baseContent := `
[prompt-overrides.build]
model = "gpt-5-preview"
agent-mode = "code"

[prompt-overrides.test]
model = "gpt-3.5-turbo"
`
	if err := os.WriteFile(configFile, []byte(baseContent), 0644); err != nil {
		t.Fatalf("failed to write base config: %v", err)
	}

	overlayContent := `
[prompt-overrides.build]
agent-mode = "planner"

[prompt-overrides.deploy]
model = "gpt-4-turbo"
`
	if err := os.WriteFile(overlayFile, []byte(overlayContent), 0644); err != nil {
		t.Fatalf("failed to write overlay config: %v", err)
	}

	c := &config.Config{ConfigFile: configFile}
	if err := c.LoadConfig(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	assertPromptOverrides(t, c)
}

func assertPromptOverrides(t *testing.T, c *config.Config) {
	t.Helper()

	// 1. modified in overlay
	buildOverride, ok := c.PromptOverrides["build"]
	if !ok {
		t.Fatal("expected override for 'build'")
	}
	if buildOverride.Model != "gpt-5-preview" {
		t.Errorf("expected build.model 'gpt-5-preview' (from base), got %q", buildOverride.Model)
	}
	if buildOverride.AgentMode != "planner" {
		t.Errorf("expected build.agent-mode 'planner' (from overlay), got %q", buildOverride.AgentMode)
	}

	// 2. only in base
	testOverride, ok := c.PromptOverrides["test"]
	if !ok {
		t.Fatal("expected override for 'test'")
	}
	if testOverride.Model != "gpt-3.5-turbo" {
		t.Errorf("expected test.model 'gpt-3.5-turbo', got %q", testOverride.Model)
	}

	// 3. only in overlay
	deployOverride, ok := c.PromptOverrides["deploy"]
	if !ok {
		t.Fatal("expected override for 'deploy'")
	}
	if deployOverride.Model != "gpt-4-turbo" {
		t.Errorf("expected deploy.model 'gpt-4-turbo', got %q", deployOverride.Model)
	}
}

func TestLoadConfigOverlayDiscoveryEnv(t *testing.T) {
	// Test that when RALPH_CONFIG is set, we look for ralph-local.toml in that directory
	dir := t.TempDir()
	configFile := filepath.Join(dir, "my-config.toml")
	overlayFile := filepath.Join(dir, "ralph-local.toml") // Name must be ralph-local.toml

	if err := os.WriteFile(configFile, []byte(`model="base"`), 0644); err != nil {
		t.Fatalf("failed to write base config: %v", err)
	}
	if err := os.WriteFile(overlayFile, []byte(`model="overlay"`), 0644); err != nil {
		t.Fatalf("failed to write overlay config: %v", err)
	}

	t.Setenv("RALPH_CONFIG", configFile)

	c := &config.Config{}
	if err := c.LoadConfig(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if c.Model != "overlay" {
		t.Errorf("expected model 'overlay', got %q", c.Model)
	}
}

func TestLoadConfigOverlayDiscoveryDefault(t *testing.T) {
	// Test that default discovery (ralph.toml in cwd) finds overlay
	dir := t.TempDir()
	wd, err := os.Getwd()
	if err != nil {
		t.Fatalf("failed to get cwd: %v", err)
	}
	if err := os.Chdir(dir); err != nil {
		t.Fatalf("failed to chdir: %v", err)
	}
	t.Cleanup(func() {
		_ = os.Chdir(wd)
	})

	if err := os.WriteFile("ralph.toml", []byte(`model="base"`), 0644); err != nil {
		t.Fatalf("failed to write base config: %v", err)
	}
	if err := os.WriteFile("ralph-local.toml", []byte(`model="overlay"`), 0644); err != nil {
		t.Fatalf("failed to write overlay config: %v", err)
	}

	c := &config.Config{}
	if err := c.LoadConfig(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if c.Model != "overlay" {
		t.Errorf("expected model 'overlay', got %q", c.Model)
	}
}

func TestLoadConfigInvalidOverlay(t *testing.T) {
	dir := t.TempDir()
	configFile := filepath.Join(dir, "ralph.toml")
	overlayFile := filepath.Join(dir, "ralph-local.toml")

	if err := os.WriteFile(configFile, []byte(`model="base"`), 0644); err != nil {
		t.Fatalf("failed to write base config: %v", err)
	}
	// Invalid TOML in overlay
	if err := os.WriteFile(overlayFile, []byte(`model = "overlay" broken`), 0644); err != nil {
		t.Fatalf("failed to write overlay config: %v", err)
	}

	c := &config.Config{ConfigFile: configFile}
	if err := c.LoadConfig(); err == nil {
		t.Fatal("expected error for invalid overlay config")
	}
}

func TestLoadConfigRejectsConfigFileKeyInOverlay(t *testing.T) {
	dir := t.TempDir()
	configFile := filepath.Join(dir, "ralph.toml")
	overlayFile := filepath.Join(dir, "ralph-local.toml")

	if err := os.WriteFile(configFile, []byte(`model="base"`), 0644); err != nil {
		t.Fatalf("failed to write base config: %v", err)
	}
	if err := os.WriteFile(overlayFile, []byte(`config-file = "./other.toml"`), 0644); err != nil {
		t.Fatalf("failed to write overlay config: %v", err)
	}

	c := &config.Config{ConfigFile: configFile}
	err := c.LoadConfig()
	if err == nil {
		t.Fatal("expected error for unsupported config-file key in overlay")
	}

	if !strings.Contains(err.Error(), "unsupported config key 'config-file'") {
		t.Fatalf("expected unsupported key error, got %v", err)
	}
}

func TestLoadConfigWithOverlayEnvDeepMerge(t *testing.T) {
	dir := t.TempDir()
	configFile := filepath.Join(dir, "ralph.toml")
	overlayFile := filepath.Join(dir, "ralph-local.toml")

	baseContent := `
[env]
OPENAI_API_KEY = "from-base"
HTTP_PROXY = "http://127.0.0.1:8080"
`
	if err := os.WriteFile(configFile, []byte(baseContent), 0644); err != nil {
		t.Fatalf("failed to write base config: %v", err)
	}

	overlayContent := `
[env]
OPENAI_API_KEY = "from-overlay"
HTTPS_PROXY = "http://127.0.0.1:8081"
`
	if err := os.WriteFile(overlayFile, []byte(overlayContent), 0644); err != nil {
		t.Fatalf("failed to write overlay config: %v", err)
	}

	c := &config.Config{ConfigFile: configFile}
	if err := c.LoadConfig(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	expected := map[string]string{
		"OPENAI_API_KEY": "from-overlay",
		"HTTP_PROXY":     "http://127.0.0.1:8080",
		"HTTPS_PROXY":    "http://127.0.0.1:8081",
	}

	if !reflect.DeepEqual(c.Env, expected) {
		t.Fatalf("expected merged env map %+v, got %+v", expected, c.Env)
	}
}
