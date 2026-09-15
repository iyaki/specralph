package cli

import (
	_ "embed"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"

	"github.com/iyaki/specralph/internal/config"
	"github.com/iyaki/specralph/internal/prompt"
)

// guideText is the static authoring contract printed by `prompts guide`.
//
//go:embed guide.md
var guideText string

// NewPromptsCommand creates the prompts command for listing and viewing prompts.
func NewPromptsCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "prompts",
		Short: "List and view available prompts",
		Long:  `List and view available built-in and custom prompts.`,
	}

	cmd.AddCommand(NewPromptsListCommand())
	cmd.AddCommand(NewPromptsShowCommand())
	cmd.AddCommand(NewPromptsValidateCommand())

	cmd.AddCommand(NewPromptsGuideCommand())

	return cmd
}

// NewPromptsGuideCommand creates the prompts guide subcommand.
func NewPromptsGuideCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "guide",
		Short: "Show the prompt authoring guide",
		Long:  `Print the authoring contract for custom prompt files.`,
		RunE: func(cmd *cobra.Command, _ []string) error {
			_, err := io.WriteString(cmd.OutOrStdout(), guideText)

			return err
		},
	}
}

// NewPromptsListCommand creates the prompts list subcommand.
func NewPromptsListCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List available prompts",
		Long:  `List all available built-in and custom prompts with descriptions.`,
		RunE: func(cmd *cobra.Command, _ []string) error {
			var cfg config.Config
			if err := cfg.LoadConfig(); err != nil {
				return fmt.Errorf("failed to load config: %w", err)
			}

			return runPromptsList(cmd.OutOrStdout(), &cfg)
		},
	}
}

// NewPromptsShowCommand creates the prompts show subcommand.
func NewPromptsShowCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "show <name>",
		Short: "Show full prompt content",
		Long:  `Display the full content of a built-in or custom prompt.`,
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			var cfg config.Config
			if err := cfg.LoadConfig(); err != nil {
				return fmt.Errorf("failed to load config: %w", err)
			}

			promptName := args[0]

			return runPromptsShow(cmd.OutOrStdout(), &cfg, promptName)
		},
	}
}

// ErrValidationFailed signals that prompts validate reported at least one
// failing check. The report itself is on stdout, so the entrypoint exits 1
// without printing an extra error line.
var ErrValidationFailed = errors.New("validation failed")

// errTargetNotFound marks a named target that matches neither a built-in nor
// a custom prompt file.
var errTargetNotFound = errors.New("target not found")

// NewPromptsValidateCommand creates the prompts validate subcommand.
func NewPromptsValidateCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "validate <target>",
		Short: "Validate a prompt file or named prompt",
		Long:  `Run static checks on a prompt file or named prompt and report the outcome of each check.`,
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			var cfg config.Config
			if err := cfg.LoadConfig(); err != nil {
				return fmt.Errorf("failed to load config: %w", err)
			}

			return runPromptsValidate(cmd.OutOrStdout(), &cfg, args[0])
		},
	}
}

func runPromptsValidate(output io.Writer, cfg *config.Config, target string) error {
	path, content, resolveErr := resolveValidationTarget(cfg, target)
	if errors.Is(resolveErr, errTargetNotFound) {
		return fmt.Errorf("prompt %q not found", target)
	}

	_, _ = fmt.Fprintln(output, path)

	var checks []prompt.Check
	if resolveErr != nil {
		checks = append(checks, prompt.Check{Name: "resolve", Status: prompt.StatusFail, Message: resolveErr.Error()})
	} else {
		checks = append(checks, prompt.Check{Name: "resolve", Status: prompt.StatusOK})
		checks = append(checks, prompt.ValidatePromptText(content)...)
	}

	failed, warnings := 0, 0
	for _, check := range checks {
		line := fmt.Sprintf("%-9s%s", check.Status, check.Name)
		if check.Message != "" {
			line += ": " + check.Message
		}
		_, _ = fmt.Fprintln(output, line)

		switch check.Status {
		case prompt.StatusFail:
			failed++
		case prompt.StatusWarn:
			warnings++
		}
	}

	summary := fmt.Sprintf("%d failed, %d warning", failed, warnings)
	if warnings != 1 {
		summary += "s"
	}
	_, _ = fmt.Fprintln(output, summary)

	if failed > 0 {
		return ErrValidationFailed
	}

	return nil
}

// resolveValidationTarget resolves the validate target to a display path and
// its content: a file path when the target ends in .md or contains a path
// separator, otherwise a prompt name (built-in first, then custom files in
// PromptsDir). The build alias resolves to its configured target first.
func resolveValidationTarget(cfg *config.Config, target string) (string, string, error) {
	isFilePath := strings.HasSuffix(target, ".md") ||
		strings.Contains(target, "/") ||
		strings.Contains(target, `\`)
	if isFilePath {
		content, err := os.ReadFile(target) // #nosec G304 -- path is the user-provided validate target

		return target, string(content), err
	}

	target = rewriteBuildAlias(target, cfg.BuildAliasPrompt)

	switch target {
	case "build-classic":
		return target, prompt.BuildPrompt(cfg), nil
	case "build-subagents":
		return target, prompt.BuildSubagentsPrompt(cfg), nil
	case "plan":
		return target, prompt.PlanPrompt(cfg, ""), nil
	}

	promptPath := cfg.PromptsDir + "/" + target + ".md"
	foundPath := findFileUpwards(promptPath)
	if foundPath == "" {
		return "", "", errTargetNotFound
	}

	content, err := os.ReadFile(foundPath) // #nosec G304 -- path is from findFileUpwards

	return foundPath, string(content), err
}

type promptInfo struct {
	Name string
	Desc string
	Path string
}

func runPromptsList(output io.Writer, cfg *config.Config) error {
	// Display built-in prompts
	const (
		maxDescLen  = 80
		maxShortLen = 70
	)
	buildClassicDesc := "Implement a single task from IMPLEMENTATION_PLAN.md after studying specs, " +
		"then validate, commit, and update the plan."
	buildSubagentsDesc := "Pick up to 10 tasks from IMPLEMENTATION_PLAN.md and execute them " +
		"with at least one subagent per task."
	planDesc := "Generate or update IMPLEMENTATION_PLAN.md with a phase-based plan after " +
		"studying specs, existing code, and identifying gaps."

	_, _ = fmt.Fprintln(output, "Built-in Prompts:")
	_, _ = fmt.Fprintf(output, "  %-15s %s\n", "build-classic", truncateString(buildClassicDesc, maxDescLen))
	_, _ = fmt.Fprintf(output, "  %-15s %s\n", "build-subagents", truncateString(buildSubagentsDesc, maxDescLen))
	_, _ = fmt.Fprintf(output, "  %-15s %s\n", "plan", truncateString(planDesc, maxDescLen))

	_, _ = fmt.Fprintf(output, "\nAliases:\n  build -> %s (configurable via build-alias-prompt)\n", cfg.BuildAliasPrompt)

	// Discover custom prompts
	customPrompts := discoverCustomPrompts(cfg)
	if len(customPrompts) > 0 {
		_, _ = fmt.Fprintln(output, "")
		_, _ = fmt.Fprintln(output, "Custom Prompts:")
		for _, p := range customPrompts {
			_, _ = fmt.Fprintf(output, "  %-10s %s\n", p.Name, truncateString(p.Desc, maxShortLen))
			_, _ = fmt.Fprintf(output, "             %s\n", p.Path)
		}
	}

	_, _ = fmt.Fprintln(output, "")
	_, _ = fmt.Fprintln(output, "Use 'ralph run <prompt-name>' to execute a prompt.")

	return nil
}

func discoverCustomPrompts(cfg *config.Config) []promptInfo {
	var prompts []promptInfo

	// Read prompts directory
	entries, err := os.ReadDir(cfg.PromptsDir)
	if err != nil {
		return prompts // return empty if directory doesn't exist
	}

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		name := entry.Name()
		if len(name) < 4 || !strings.HasSuffix(name, ".md") {
			continue
		}

		promptName := name[:len(name)-3]
		promptPath := cfg.PromptsDir + "/" + name

		// Extract description from file
		desc := extractDescription(promptPath)

		prompts = append(prompts, promptInfo{
			Name: promptName,
			Desc: desc,
			Path: "./prompts/" + name,
		})
	}

	return prompts
}

func extractDescription(path string) string {
	content, err := os.ReadFile(path) // #nosec G304 -- path is from discovered prompts
	if err != nil {
		return "(cannot read file)"
	}

	// Invalid frontmatter yaml is worth distinguishing from a missing description.
	if _, _, err := prompt.ParseFrontMatter(string(content)); err != nil {
		return "(invalid frontmatter)"
	}

	if desc := prompt.ExtractDescription(string(content)); desc != "" {
		return desc
	}

	return "(no description)"
}
func truncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}

	// Add ellipsis
	return s[:maxLen-3] + "..."
}

func runPromptsShow(output io.Writer, cfg *config.Config, promptName string) error {
	// Resolve the build alias once via the shared rule; the rewritten name then
	// resolves like any other prompt name.
	promptName = rewriteBuildAlias(promptName, cfg.BuildAliasPrompt)

	// Try built-in prompts first
	switch promptName {
	case "build-classic":
		_, _ = fmt.Fprint(output, prompt.BuildPrompt(cfg))

		return nil
	case "build-subagents":
		_, _ = fmt.Fprint(output, prompt.BuildSubagentsPrompt(cfg))

		return nil
	case "plan":
		// Plan prompt needs a scope - use empty string for display
		_, _ = fmt.Fprint(output, prompt.PlanPrompt(cfg, ""))

		return nil
	}

	// Try custom prompt file
	promptPath := cfg.PromptsDir + "/" + promptName + ".md"
	foundPath := findFileUpwards(promptPath)
	if foundPath == "" {
		return fmt.Errorf("prompt %q not found", promptName)
	}

	content, err := os.ReadFile(foundPath) // #nosec G304 -- path is from findFileUpwards
	if err != nil {
		return fmt.Errorf("failed to read prompt file %q: %w", foundPath, err)
	}

	// Parse and strip frontmatter
	_, body, err := prompt.ParseFrontMatter(string(content))
	if err != nil {
		return fmt.Errorf("failed to parse front matter in %q: %w", foundPath, err)
	}

	_, _ = fmt.Fprint(output, body)

	return nil
}

// findFileUpwards searches for a file from the current directory upwards.
func findFileUpwards(path string) string {
	// Check if file exists at the given path
	if _, err := os.Stat(path); err == nil {
		return path
	}

	// Search upwards from current directory
	dir, err := os.Getwd()
	if err != nil {
		return ""
	}

	for {
		fullPath := dir + "/" + path
		if _, err := os.Stat(fullPath); err == nil {
			return fullPath
		}

		parent := filepath.Dir(dir)
		if parent == dir {
			// Reached root
			break
		}
		dir = parent
	}

	return ""
}
