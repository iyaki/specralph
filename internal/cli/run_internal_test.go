package cli

import (
	"os"
	"testing"

	"github.com/iyaki/ralphex/internal/config"
	"github.com/spf13/cobra"
)

func TestReadBoolFlagOverride(t *testing.T) {
	t.Run("flag not changed returns unchanged false", func(t *testing.T) {
		cmd := &cobra.Command{}
		cmd.Flags().Bool("test-flag", false, "")

		result, err := readBoolFlagOverride(cmd, "test-flag")
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
		if result.changed {
			t.Error("expected changed=false")
		}
		if result.value {
			t.Error("expected value=false")
		}
	})

	t.Run("flag changed to true returns changed true with value", func(t *testing.T) {
		cmd := &cobra.Command{}
		cmd.Flags().Bool("test-flag", false, "")
		if err := cmd.Flags().Set("test-flag", "true"); err != nil {
			t.Fatalf("failed to set flag: %v", err)
		}

		result, err := readBoolFlagOverride(cmd, "test-flag")
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
		if !result.changed {
			t.Error("expected changed=true")
		}
		if !result.value {
			t.Error("expected value=true")
		}
	})
}

func TestApplyBoolFlagOverrides(t *testing.T) {
	t.Run("changed=true updates config", func(t *testing.T) {
		cfg := &config.Config{}
		override := boolFlagOverride{changed: true, value: true}

		applyBoolFlagOverrides(cfg, override)

		if !cfg.LogTruncate {
			t.Error("expected LogTruncate to be true")
		}
	})

	t.Run("changed=false leaves config unchanged", func(t *testing.T) {
		cfg := &config.Config{LogTruncate: false}
		override := boolFlagOverride{changed: false, value: true}

		applyBoolFlagOverrides(cfg, override)

		if cfg.LogTruncate {
			t.Error("expected LogTruncate to remain false")
		}
	})
}

func TestReadEnvFlagOverrides(t *testing.T) {
	t.Run("no --env flags returns nil", func(t *testing.T) {
		cmd := &cobra.Command{}
		cmd.Flags().StringArray("env", []string{}, "")

		result, err := readEnvFlagOverrides(cmd)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
		if result != nil {
			t.Errorf("expected nil, got %v", result)
		}
	})

	t.Run("single KEY=value returns map", func(t *testing.T) {
		cmd := &cobra.Command{}
		cmd.Flags().StringArray("env", []string{}, "")
		if err := cmd.Flags().Set("env", "KEY=value"); err != nil {
			t.Fatalf("failed to set flag: %v", err)
		}

		result, err := readEnvFlagOverrides(cmd)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
		if result["KEY"] != "value" {
			t.Errorf("expected KEY=value, got %v", result)
		}
	})

	t.Run("invalid format returns error", func(t *testing.T) {
		cmd := &cobra.Command{}
		cmd.Flags().StringArray("env", []string{}, "")
		if err := cmd.Flags().Set("env", "INVALID_NO_EQUALS"); err != nil {
			t.Fatalf("failed to set flag: %v", err)
		}

		_, err := readEnvFlagOverrides(cmd)
		if err == nil {
			t.Error("expected error for invalid format")
		}
	})

	t.Run("invalid key format returns error", func(t *testing.T) {
		cmd := &cobra.Command{}
		cmd.Flags().StringArray("env", []string{}, "")
		if err := cmd.Flags().Set("env", "invalid-key=value"); err != nil {
			t.Fatalf("failed to set flag: %v", err)
		}

		_, err := readEnvFlagOverrides(cmd)
		if err == nil {
			t.Error("expected error for invalid key")
		}
	})
}

func TestApplyEnvFlagOverrides(t *testing.T) {
	t.Run("nil overrides leaves config unchanged", func(t *testing.T) {
		cfg := &config.Config{}

		applyEnvFlagOverrides(cfg, nil)

		if cfg.Env != nil {
			t.Errorf("expected Env to remain nil, got %v", cfg.Env)
		}
	})

	t.Run("with overrides populates config", func(t *testing.T) {
		cfg := &config.Config{}
		overrides := map[string]string{"KEY1": "val1", "KEY2": "val2"}

		applyEnvFlagOverrides(cfg, overrides)

		if cfg.Env["KEY1"] != "val1" {
			t.Errorf("expected KEY1=val1, got %v", cfg.Env)
		}
		if cfg.Env["KEY2"] != "val2" {
			t.Errorf("expected KEY2=val2, got %v", cfg.Env)
		}
	})
}

func TestParsePositionalArgs(t *testing.T) {
	tests := []struct {
		name     string
		args     []string
		wantGoal string
		wantDesc string
	}{
		{"empty args", []string{}, "build", "Whole system"},
		{"one arg", []string{"test"}, "test", "Whole system"},
		{"two args", []string{"test", "description"}, "test", "description"},
		{"three args", []string{"a", "b", "c"}, "a", "b"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			goal, desc := parsePositionalArgs(tc.args)
			if goal != tc.wantGoal {
				t.Errorf("expected goal %q, got %q", tc.wantGoal, goal)
			}
			if desc != tc.wantDesc {
				t.Errorf("expected desc %q, got %q", tc.wantDesc, desc)
			}
		})
	}
}

func TestApplyModelSettings(t *testing.T) {
	t.Run("flag changed leaves config unchanged", func(t *testing.T) {
		cmd := &cobra.Command{}
		cmd.Flags().String("model", "", "")
		if err := cmd.Flags().Set("model", "flag-model"); err != nil {
			t.Fatalf("failed to set flag: %v", err)
		}
		cfg := &config.Config{Model: "original"}
		fmOverride := &config.PromptConfigOverride{Model: "fm-model"}

		applyModelSettings(cfg, cmd, fmOverride, "test-prompt")

		if cfg.Model != "original" {
			t.Errorf("expected Model unchanged %q, got %q", "original", cfg.Model)
		}
	})

	t.Run("env set leaves config unchanged", func(t *testing.T) {
		cmd := &cobra.Command{}
		cfg := &config.Config{Model: "original"}
		t.Setenv("RALPH_MODEL", "env-model")

		applyModelSettings(cfg, cmd, nil, "test-prompt")

		if cfg.Model != "original" {
			t.Errorf("expected Model unchanged, got %q", cfg.Model)
		}
		_ = os.Unsetenv("RALPH_MODEL")
	})

	t.Run("front matter has model sets config", func(t *testing.T) {
		cmd := &cobra.Command{}
		cfg := &config.Config{}
		fmOverride := &config.PromptConfigOverride{Model: "fm-model"}

		applyModelSettings(cfg, cmd, fmOverride, "test-prompt")

		if cfg.Model != "fm-model" {
			t.Errorf("expected Model %q, got %q", "fm-model", cfg.Model)
		}
	})

	t.Run("prompt override has model sets config", func(t *testing.T) {
		cmd := &cobra.Command{}
		cfg := &config.Config{
			PromptOverrides: map[string]config.PromptConfigOverride{
				"test-prompt": {Model: "override-model"},
			},
		}

		applyModelSettings(cfg, cmd, nil, "test-prompt")

		if cfg.Model != "override-model" {
			t.Errorf("expected Model %q, got %q", "override-model", cfg.Model)
		}
	})
}

func TestApplyAgentModeSettings(t *testing.T) {
	t.Run("flag changed leaves config unchanged", func(t *testing.T) {
		cmd := &cobra.Command{}
		cmd.Flags().String("agent-mode", "", "")
		if err := cmd.Flags().Set("agent-mode", "flag-mode"); err != nil {
			t.Fatalf("failed to set flag: %v", err)
		}
		cfg := &config.Config{AgentMode: "original"}

		applyAgentModeSettings(cfg, cmd, nil, "test-prompt")

		if cfg.AgentMode != "original" {
			t.Errorf("expected AgentMode unchanged, got %q", cfg.AgentMode)
		}
	})

	t.Run("env set leaves config unchanged", func(t *testing.T) {
		cmd := &cobra.Command{}
		cfg := &config.Config{AgentMode: "original"}
		t.Setenv("RALPH_AGENT_MODE", "env-mode")

		applyAgentModeSettings(cfg, cmd, nil, "test-prompt")

		if cfg.AgentMode != "original" {
			t.Errorf("expected AgentMode unchanged, got %q", cfg.AgentMode)
		}
		_ = os.Unsetenv("RALPH_AGENT_MODE")
	})

	t.Run("front matter has agent-mode sets config", func(t *testing.T) {
		cmd := &cobra.Command{}
		cfg := &config.Config{}
		fmOverride := &config.PromptConfigOverride{AgentMode: "fm-mode"}

		applyAgentModeSettings(cfg, cmd, fmOverride, "test-prompt")

		if cfg.AgentMode != "fm-mode" {
			t.Errorf("expected AgentMode %q, got %q", "fm-mode", cfg.AgentMode)
		}
	})

	t.Run("prompt override has agent-mode sets config", func(t *testing.T) {
		cmd := &cobra.Command{}
		cfg := &config.Config{
			PromptOverrides: map[string]config.PromptConfigOverride{
				"test-prompt": {AgentMode: "override-mode"},
			},
		}

		applyAgentModeSettings(cfg, cmd, nil, "test-prompt")

		if cfg.AgentMode != "override-mode" {
			t.Errorf("expected AgentMode %q, got %q", "override-mode", cfg.AgentMode)
		}
	})
}
