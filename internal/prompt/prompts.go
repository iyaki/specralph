// Package prompt builds and resolves prompt text for Ralph commands.
package prompt

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/iyaki/specralph/internal/config"
)

const (
	planPromptLinesCapacity   = 128
	outputFormatLinesCapacity = 80
)

// GetPrompt returns the prompt to use based on configuration and arguments.
func GetPrompt(
	cfg *config.Config,
	promptName, scope string,
	output io.Writer,
) (string, *config.PromptConfigOverride, error) {
	if promptText, ok, err := customPrompt(cfg, output); ok || err != nil {
		return promptText, nil, err
	}

	if promptText, ok, err := stdinPrompt(cfg, promptName, output); ok || err != nil {
		return promptText, nil, err
	}

	if promptText, override, ok, err := explicitPromptFile(cfg, output); ok || err != nil {
		return promptText, override, err
	}

	if promptText, override, ok, err := promptFromDir(cfg, promptName, output); ok || err != nil {
		return promptText, override, err
	}

	promptText, err := bundledPrompt(cfg, promptName, scope, output)

	return promptText, nil, err
}

func customPrompt(cfg *config.Config, output io.Writer) (string, bool, error) {
	if cfg.CustomPrompt == "" {
		return "", false, nil
	}

	writeBanner(output, "               USING INLINE CUSTOM PROMPT")

	return cfg.CustomPrompt, true, nil
}

func stdinPrompt(cfg *config.Config, promptName string, output io.Writer) (string, bool, error) {
	if cfg.PromptFile != "-" && promptName != "-" {
		return "", false, nil
	}

	content, err := io.ReadAll(os.Stdin)
	if err != nil {
		return "", true, fmt.Errorf("failed to read from stdin: %w", err)
	}

	writeBanner(output, "               USING PROMPT FROM STDIN")

	return string(content), true, nil
}

func explicitPromptFile(cfg *config.Config, output io.Writer) (string, *config.PromptConfigOverride, bool, error) {
	if cfg.PromptFile == "" {
		return "", nil, false, nil
	}

	content, err := os.ReadFile(cfg.PromptFile)
	if err != nil {
		return "", nil, true, fmt.Errorf("failed to read prompt file %s: %w", cfg.PromptFile, err)
	}

	writePromptFileBanner(output, cfg.PromptFile)

	fm, body, err := ParseFrontMatter(string(content))
	if err != nil {
		return "", nil, true, fmt.Errorf("failed to parse front matter in %s: %w", cfg.PromptFile, err)
	}

	override := &config.PromptConfigOverride{
		Model:     fm.Model,
		AgentMode: fm.AgentMode,
	}

	return body, override, true, nil
}

func promptFromDir(
	cfg *config.Config,
	promptName string,
	output io.Writer,
) (string, *config.PromptConfigOverride, bool, error) {
	promptFilePath := filepath.Join(cfg.PromptsDir, promptName+".md")
	foundPath := findFileUpwards(promptFilePath)
	if foundPath == "" {
		return "", nil, false, nil
	}

	content, err := os.ReadFile(foundPath) // #nosec G304 -- path is discovered within project tree
	if err != nil {
		return "", nil, true, fmt.Errorf("failed to read prompt file %s: %w", foundPath, err)
	}

	writePromptFileBanner(output, foundPath)

	fm, body, err := ParseFrontMatter(string(content))
	if err != nil {
		return "", nil, true, fmt.Errorf("failed to parse front matter in %s: %w", foundPath, err)
	}

	override := &config.PromptConfigOverride{
		Model:     fm.Model,
		AgentMode: fm.AgentMode,
	}

	return body, override, true, nil
}

func bundledPrompt(cfg *config.Config, promptName, scope string, output io.Writer) (string, error) {
	switch promptName {
	case "build":
		writeBanner(output, "               USING DEFAULT 'BUILD' PROMPT")

		return BuildPrompt(cfg), nil
	case "plan":
		writeBanner(output, "               USING DEFAULT 'PLAN' PROMPT")

		return PlanPrompt(cfg, scope), nil
	default:
		return "", fmt.Errorf(
			"prompt file not found for '%s'. Use a valid prompt file or one of the pre-bundled prompts (build, plan)",
			promptName,
		)
	}
}

func writeBanner(output io.Writer, title string) {
	_, _ = fmt.Fprintln(output, "")
	_, _ = fmt.Fprintln(output, "===============================================================")
	_, _ = fmt.Fprintln(output, title)
	_, _ = fmt.Fprintln(output, "===============================================================")
	_, _ = fmt.Fprintln(output, "")
}

func writePromptFileBanner(output io.Writer, path string) {
	_, _ = fmt.Fprintln(output, "")
	_, _ = fmt.Fprintln(output, "===============================================================")
	_, _ = fmt.Fprintf(output, " USING PROMPT FILE: %s\n", path)
	_, _ = fmt.Fprintln(output, "===============================================================")
	_, _ = fmt.Fprintln(output, "")
}

func joinPromptLines(lines ...string) string {
	return strings.Join(lines, "\n") + "\n"
}

// BuildPrompt generates the default build prompt.
func BuildPrompt(cfg *config.Config) string {
	specsIndexFileReference := ""
	if cfg.SpecsIndexFile != "" && !cfg.NoSpecsIndex {
		specsIndexFileReference = filepath.Join(cfg.SpecsDir, cfg.SpecsIndexFile)
	}

	specsIndexFileReferenceText := ""
	if specsIndexFileReference != "" {
		specsIndexFileReferenceText = fmt.Sprintf(" (including `%s` and related specs)", specsIndexFileReference)
	}

	return joinPromptLines(
		"# Agent Instructions (Build Mode)",
		"",
		fmt.Sprintf("- Study `%s/*`%s.", cfg.SpecsDir, specsIndexFileReferenceText),
		fmt.Sprintf("- Study `%s` and pick the single most important task.", cfg.ImplementationPlanName),
		"- Implement the task",
		"- Validate the implementation",
		"- Commit the changes",
		"- Update the plan",
		"- Commit the update plan",
		"- Stop after the commit",
		"",
		"## Stop Condition",
		"",
		"- After completing the selected task, stop. Do NOT start another task in the same run.",
		"- If ALL stories are complete and passing, reply with:",
		"  `<COMPLETION_SIGNAL>`",
		"",
		"## IMPORTANT",
		"",
		"- Before changes, search the codebase. Do NOT assume functionality is missing.",
		"- Implement ONLY one task. Stop after committing.",
		fmt.Sprintf("- Update `%s` when the task is done.", cfg.ImplementationPlanName),
		"- Use the verification log format: `YYYY-MM-DD: <command or URL> - <result>`.",
		"- Keep a `Manual Deployment Tasks` section in implementation",
		"  the plan and use `None` when there are no tasks.",
		fmt.Sprintf(
			"- You may implement missing functionality if required, but study relevant `%s/*` first.",
			cfg.SpecsDir,
		),
		"- You may add temporary logging as needed and remove if no longer needed.",
		"",
	)
}

// PlanPrompt generates the default plan prompt.
func PlanPrompt(cfg *config.Config, scope string) string {
	return joinPromptLines(planPromptLines(cfg, scope)...)
}

func planPromptLines(cfg *config.Config, scope string) []string {
	lines := make([]string, 0, planPromptLinesCapacity)
	lines = append(lines,
		"# Agent Instructions (Planning Mode)",
		"",
		fmt.Sprintf("Scope: %s", scope),
		"",
		"## Objective",
		"",
		fmt.Sprintf(
			"Generate or update `%s` in a structured, phase-based format with:",
			cfg.ImplementationPlanName,
		),
		"",
		"- Clear status metadata",
		"- Quick reference tables",
		"- Phase sections with paths and checklists",
		"- Verification log entries",
		"- Summary tables and remaining effort",
		"",
		"Plan only. Do NOT implement anything.",
		"",
	)

	lines = append(lines, planStudyAndGapLines(cfg)...)
	lines = append(lines, planOutputFormatLines(cfg)...)
	lines = append(lines, planStopConditionLines(cfg)...)

	return lines
}

func planStudyAndGapLines(cfg *config.Config) []string {
	return []string{
		"## Study and Gap Analysis",
		"",
		fmt.Sprintf("- Study `%s/*` to learn application requirements.", cfg.SpecsDir),
		fmt.Sprintf("- Study `%s` (if present; it may be incorrect).", cfg.ImplementationPlanName),
		"- Study relevant source code to compare against specs.",
		"- Use `git` to study recent changes on the specs related to the specified current scope.",
		"",
		"Rules:",
		"",
		"- Do NOT assume missing; confirm via code search first.",
		"- Identify where work already exists, partial implementations, TODOs, placeholders,",
		"  skipped/flaky tests, or inconsistent patterns.",
		"- Keep the plan concise but complete; prefer lists and tables over paragraphs.",
		"- Use `[x]` only when verified in code. Use `[ ]` if missing or unverified.",
		"- Regenerate the plan if it becomes stale, contradictory,",
		"  or significantly out of sync with code.",
		"- If the specified scope has relationships with other domain areas,",
		"  implementation may be needed in those areas as well",
		"  (always study the related specs and code). Include this in the plan.",
		"",
	}
}

func planOutputFormatLines(cfg *config.Config) []string {
	lines := make([]string, 0, outputFormatLinesCapacity)
	lines = append(lines,
		"## Output Format Requirements",
		"",
		fmt.Sprintf("Write `%s` using this structure and level of detail:", cfg.ImplementationPlanName),
		"",
	)

	lines = append(lines, planOutputHeaderLines()...)
	lines = append(lines, planOutputPhasedAndVerificationLines()...)
	lines = append(lines, planOutputSummaryAndManualLines()...)

	return lines
}

func planOutputHeaderLines() []string {
	return []string{
		"Header",
		"",
		"- Title: `Implementation Plan (<Scope>)`",
		"- Status line: `**Status:** <summary (e.g., \"UI Components Complete (39/39)\")>`",
		"- Last Updated date: `YYYY-MM-DD`",
		"- Reference to primary spec(s)",
		"",
		"Quick Reference",
		"",
		"- A table mapping systems/subsystems to:",
		"  - Specs",
		"  - Modules/packages",
		"  - Web packages",
		"  - Migrations or other artifacts",
		"- Use `✅` to mark items already implemented.",
		"",
	}
}

func planOutputPhasedAndVerificationLines() []string {
	return []string{
		"Phased Plan",
		"",
		"- Use numbered phases (e.g., Phase 9, Phase 10) aligned to the spec's domain.",
		"- Each phase includes:",
		"  - Goal",
		"  - Status (if applicable)",
		"  - Paths (directories or file patterns)",
		"  - Checklist with `[x]` for verified complete and `[ ]` for missing",
		"  - Definition of Done (tests run, commands/URLs, files touched)",
		"  - Risks/Dependencies (brief)",
		"- Break phases into subsections (e.g., 9.1, 9.2)",
		"  with scope-specific paths and item lists.",
		"- Include \"Reference pattern\" links",
		"  when there's a canonical directory or file to follow.",
		"",
		"Verification Log",
		"",
		"- A chronological log of verification steps with dates.",
		"- Each entry includes:",
		"  - What was verified (endpoints, commands, builds, tests, UI routes)",
		"  - Exact commands or URLs used",
		"  - Tests run and results",
		"  - Bug fixes discovered (if any)",
		"  - Files touched (if known from code search)",
		"  - Use format: `YYYY-MM-DD: <command or URL> - <result>`",
		"",
	}
}

func planOutputSummaryAndManualLines() []string {
	return []string{
		"Summary",
		"",
		"- Table of phases with completion status",
		"  - \"Remaining effort\" line summarizing unfinished sections",
		"",
		"Known Existing Work",
		"",
		"- Brief section listing confirmed existing implementations to prevent duplicate work",
		"",
		"Manual Deployment Tasks",
		"",
		"- Required section to document manual steps needed before or during",
		"  production deployment (manual configuration, third-party service setup,",
		"  API key acquisition, etc).",
		"- If not applicable, write exactly: `None`.",
		"",
	}
}

func planStopConditionLines(cfg *config.Config) []string {
	return []string{
		"## Stop Condition",
		"",
		fmt.Sprintf(
			"**IMPORTANT**: After writing/updating if `%s` already reflects the current gaps, reply with:",
			cfg.ImplementationPlanName,
		),
		"`<COMPLETION_SIGNAL>`",
		"",
	}
}

// findFileUpwards searches for a file from the current directory upwards.
func findFileUpwards(path string) string {
	if filepath.IsAbs(path) {
		if _, err := os.Stat(path); err == nil {
			return path
		}

		return ""
	}

	cwd, err := os.Getwd()
	if err != nil {
		return ""
	}

	dir := cwd
	for {
		testPath := filepath.Join(dir, path)
		if _, err := os.Stat(testPath); err == nil {
			return testPath
		}

		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}

		dir = parent
	}

	return ""
}
