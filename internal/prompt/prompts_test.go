package prompt_test

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/iyaki/specralph/internal/config"
	"github.com/iyaki/specralph/internal/prompt"
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

func TestBuildSubagentsPromptMatchesSpecOutline(t *testing.T) {
	cfg := &config.Config{
		SpecsDir:               "specs",
		SpecsIndexFile:         "README.md",
		ImplementationPlanName: "IMPLEMENTATION_PLAN.md",
	}
	want := joinSpecOutlineLines(
		"# Agent Instructions (Build Mode with Subagents)",
		"",
		"- Study `specs/*` (including `specs/README.md` and related specs).",
		"- Study `IMPLEMENTATION_PLAN.md` and select the next pending tasks, up to a maximum of 10.",
		"- If no tasks are pending, verify the plan is complete and reply with `<COMPLETION_SIGNAL>`.",
		"",
		"## Task Execution",
		"",
		"- Dispatch at least one subagent per selected task using the agent CLI's native",
		"  subagent mechanism.",
		"- Give each subagent a self-contained brief: the task, relevant spec paths, and",
		"  validation commands.",
		"- Orchestrate only: tasks are implemented by subagents, not directly in the main context.",
		"- After each task: validate, update `IMPLEMENTATION_PLAN.md`, and commit code and",
		"  plan update together.",
		"- If a task fails, record it in the plan as not complete and continue with the",
		"  remaining selected tasks.",
		"",
		"## Stop Condition",
		"",
		"- After the selected batch, stop. Do NOT pick up more tasks in the same run.",
		"- If and only if ALL stories are complete and passing, reply with `<COMPLETION_SIGNAL>`.",
		"",
		"## IMPORTANT",
		"",
		"- Before changes, search the codebase. Do NOT assume functionality is missing.",
		"- Use the verification log format: `YYYY-MM-DD: <command or URL> - <result>`.",
		"- Keep a `Manual Deployment Tasks` section in the plan and use `None` when there are no tasks.",
		"- You may add temporary logging as needed and remove if no longer needed.",
		"",
	)
	if got := prompt.BuildSubagentsPrompt(cfg); got != want {
		t.Fatalf("build-subagents prompt does not match spec outline:\n got: %q\nwant: %q", got, want)
	}
}

func TestBuildSubagentsPromptNoSpecsIndexOmitsParenthetical(t *testing.T) {
	cfg := &config.Config{
		SpecsDir:               "specs",
		SpecsIndexFile:         "README.md",
		NoSpecsIndex:           true,
		ImplementationPlanName: "IMPLEMENTATION_PLAN.md",
	}
	p := prompt.BuildSubagentsPrompt(cfg)
	if strings.Contains(p, "including") {
		t.Fatalf("expected specs index parenthetical to be omitted, got:\n%s", p)
	}
	if !strings.Contains(p, "- Study `specs/*`.") {
		t.Fatalf("expected bare specs study line, got:\n%s", p)
	}
}

func TestBuiltInPromptSpecMarkers(t *testing.T) {
	cfg := &config.Config{
		SpecsDir:               "specs",
		SpecsIndexFile:         "README.md",
		ImplementationPlanName: "IMPLEMENTATION_PLAN.md",
	}
	classic := prompt.BuildPrompt(cfg)
	subagents := prompt.BuildSubagentsPrompt(cfg)
	tests := []struct {
		name    string
		content string
		marker  string
	}{
		// specs/prompts/build-classic.md verifications
		{"build-classic title", classic, "# Agent Instructions (Build Mode)"},
		{"build-classic single task", classic, "pick the single most important task"},
		{"build-classic stop condition", classic, "Do NOT start another task in the same run"},
		{"build-classic verification log format", classic, "YYYY-MM-DD: <command or URL> - <result>"},
		{"build-classic manual deployment tasks", classic, "Manual Deployment Tasks"},
		{"build-classic completion signal", classic, "<COMPLETION_SIGNAL>"},
		// specs/prompts/build-subagents.md verifications
		{"build-subagents title", subagents, "# Agent Instructions (Build Mode with Subagents)"},
		{"build-subagents 10-task cap", subagents, "up to a maximum of 10"},
		{"build-subagents at-least-one-subagent rule", subagents, "at least one subagent per selected task"},
		{"build-subagents orchestrate-only", subagents, "Orchestrate only"},
		{"build-subagents per-task validate/commit/plan-update", subagents,
			"After each task: validate, update `IMPLEMENTATION_PLAN.md`"},
		{"build-subagents failed-task record-and-continue", subagents, "record it in the plan as not complete"},
		{"build-subagents stop after batch", subagents, "Do NOT pick up more tasks in the same run"},
		{"build-subagents signal only when all complete", subagents, "If and only if ALL stories are complete and passing"},
		{"build-subagents completion signal", subagents, "<COMPLETION_SIGNAL>"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if !strings.Contains(tt.content, tt.marker) {
				t.Fatalf("expected marker %q in content:\n%s", tt.marker, tt.content)
			}
		})
	}
}

func joinSpecOutlineLines(lines ...string) string {
	return strings.Join(lines, "\n") + "\n"
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

func TestGetPromptBundledBuiltIns(t *testing.T) {
	cfg := &config.Config{SpecsDir: "specs", SpecsIndexFile: "README.md", ImplementationPlanName: "PLAN.md"}

	var classicOut bytes.Buffer
	classicPrompt, _, err := prompt.GetPrompt(cfg, "build-classic", "scope", &classicOut)
	if err != nil {
		t.Fatalf("unexpected error for build-classic: %v", err)
	}
	if classicPrompt != prompt.BuildPrompt(cfg) {
		t.Fatalf("build-classic output must be byte-identical to BuildPrompt:\n got: %q\nwant: %q",
			classicPrompt, prompt.BuildPrompt(cfg))
	}
	if !strings.Contains(classicOut.String(), "USING DEFAULT 'BUILD-CLASSIC' PROMPT") {
		t.Fatalf("expected build-classic banner, got: %q", classicOut.String())
	}

	var subagentsOut bytes.Buffer
	subagentsPrompt, _, err := prompt.GetPrompt(cfg, "build-subagents", "scope", &subagentsOut)
	if err != nil {
		t.Fatalf("unexpected error for build-subagents: %v", err)
	}
	if subagentsPrompt != prompt.BuildSubagentsPrompt(cfg) {
		t.Fatalf("build-subagents output must equal BuildSubagentsPrompt:\n got: %q\nwant: %q",
			subagentsPrompt, prompt.BuildSubagentsPrompt(cfg))
	}
	if !strings.Contains(subagentsOut.String(), "USING DEFAULT 'BUILD-SUBAGENTS' PROMPT") {
		t.Fatalf("expected build-subagents banner, got: %q", subagentsOut.String())
	}

	planPrompt, _, err := prompt.GetPrompt(cfg, "plan", "My Scope", &bytes.Buffer{})
	if err != nil {
		t.Fatalf("unexpected error for plan: %v", err)
	}
	if !strings.Contains(planPrompt, "Scope: My Scope") {
		t.Fatalf("unexpected plan prompt: %q", planPrompt)
	}
}

func TestGetPromptBuildNameIsNotBuiltIn(t *testing.T) {
	cfg := &config.Config{SpecsDir: "specs", SpecsIndexFile: "README.md", ImplementationPlanName: "PLAN.md"}
	_, _, err := prompt.GetPrompt(cfg, "build", "scope", &bytes.Buffer{})
	if err == nil {
		t.Fatal("expected error for plain build name; alias rewrite lives in the CLI layer")
	}
	for _, name := range []string{"build-classic", "build-subagents", "plan"} {
		if !strings.Contains(err.Error(), name) {
			t.Fatalf("expected error to list built-in %q, got: %v", name, err)
		}
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

func TestHasCompletionSignal(t *testing.T) {
	tests := []struct {
		name     string
		text     string
		expected bool
	}{
		{name: "placeholder", text: "Stop when done with <COMPLETION_SIGNAL>", expected: true},
		{name: "literal", text: "<promise>COMPLETE</promise>", expected: true},
		{name: "literal inside sentence", text: "Reply with <promise>COMPLETE</promise> when finished.", expected: true},
		{name: "absent", text: "# Objective\n\nDo the work.", expected: false},
		{name: "empty", text: "", expected: false},
		{name: "partial placeholder only", text: "COMPLETION_SIGNAL without brackets", expected: false},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if got := prompt.HasCompletionSignal(tc.text); got != tc.expected {
				t.Errorf("HasCompletionSignal(%q) = %v, want %v", tc.text, got, tc.expected)
			}
		})
	}
}

func TestExtractDescription(t *testing.T) {
	tests := []struct {
		name     string
		text     string
		expected string
	}{
		{
			name:     "frontmatter description wins",
			text:     "---\ndescription: Author prompts with ease\n---\n\n# Objective\n\nBody line here.\n",
			expected: "Author prompts with ease",
		},
		{
			name:     "empty frontmatter description falls back to body",
			text:     "---\ndescription:\n---\n\n# Objective\n\nFirst body line.\n",
			expected: "First body line.",
		},
		{
			name:     "no frontmatter uses first non-empty non-heading body line",
			text:     "# Objective\n\n   \nFirst real line.\nSecond line.\n",
			expected: "First real line.",
		},
		{
			name:     "no description anywhere",
			text:     "# Only A Heading\n\n## Another\n",
			expected: "",
		},
		{
			name:     "empty text",
			text:     "",
			expected: "",
		},
		{
			name:     "invalid frontmatter yaml",
			text:     "---\ndescription: [unclosed\n---\n\nBody line.\n",
			expected: "",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if got := prompt.ExtractDescription(tc.text); got != tc.expected {
				t.Errorf("ExtractDescription(%q) = %q, want %q", tc.text, got, tc.expected)
			}
		})
	}
}

var validatePromptTextTests = []struct {
	name     string
	text     string
	expected []prompt.Check
}{
	{
		name: "fully valid prompt",
		text: "---\ndescription: A good prompt\n---\n\n# Objective\n\nRun tasks until <promise>COMPLETE</promise>.\n",
		expected: []prompt.Check{
			{Name: "frontmatter", Status: prompt.StatusOK, Message: ""},
			{Name: "description", Status: prompt.StatusOK, Message: ""},
			{Name: "completion-signal", Status: prompt.StatusOK, Message: ""},
		},
	},
	{
		name: "missing completion signal fails",
		text: "---\ndescription: No signal\n---\n\n# Objective\n\nDo work forever.\n",
		expected: []prompt.Check{
			{Name: "frontmatter", Status: prompt.StatusOK, Message: ""},
			{Name: "description", Status: prompt.StatusOK, Message: ""},
			{Name: "completion-signal", Status: prompt.StatusFail,
				Message: "neither <COMPLETION_SIGNAL> nor <promise>COMPLETE</promise> found"},
		},
	},
	{
		name: "unbalanced frontmatter warns and later checks still run",
		text: "---\ndescription: Broken delimiters\n\n# Objective\n\n<promise>COMPLETE</promise>\n",
		expected: []prompt.Check{
			{Name: "frontmatter", Status: prompt.StatusWarn, Message: "unbalanced frontmatter delimiters"},
			{Name: "description", Status: prompt.StatusOK, Message: ""},
			{Name: "completion-signal", Status: prompt.StatusOK, Message: ""},
		},
	},
	{
		name: "empty frontmatter description warns frontmatter but description ok via body",
		text: "---\ndescription:\n---\n\n# Objective\n\nBody line.\n<COMPLETION_SIGNAL>\n",
		expected: []prompt.Check{
			{Name: "frontmatter", Status: prompt.StatusWarn, Message: "frontmatter description is empty"},
			{Name: "description", Status: prompt.StatusOK, Message: ""},
			{Name: "completion-signal", Status: prompt.StatusOK, Message: ""},
		},
	},
	{
		name: "signal placeholder line counts as body line",
		text: "# Objective\n\n<COMPLETION_SIGNAL>\n",
		expected: []prompt.Check{
			{Name: "frontmatter", Status: prompt.StatusOK, Message: ""},
			{Name: "description", Status: prompt.StatusOK, Message: ""},
			{Name: "completion-signal", Status: prompt.StatusOK, Message: ""},
		},
	},
	{
		name: "heading-only prompt warns description and fails signal",
		text: "# Only A Heading\n",
		expected: []prompt.Check{
			{Name: "frontmatter", Status: prompt.StatusOK, Message: ""},
			{Name: "description", Status: prompt.StatusWarn, Message: "no description line found"},
			{Name: "completion-signal", Status: prompt.StatusFail,
				Message: "neither <COMPLETION_SIGNAL> nor <promise>COMPLETE</promise> found"},
		},
	},
	{
		name: "no frontmatter at all is not a frontmatter problem",
		text: "# Objective\n\nFirst line.\n<COMPLETION_SIGNAL>\n",
		expected: []prompt.Check{
			{Name: "frontmatter", Status: prompt.StatusOK, Message: ""},
			{Name: "description", Status: prompt.StatusOK, Message: ""},
			{Name: "completion-signal", Status: prompt.StatusOK, Message: ""},
		},
	},
	{
		name: "non-string frontmatter description does not warn frontmatter",
		text: "---\ndescription: [not, a, string]\n---\n\nBody line.\n<COMPLETION_SIGNAL>\n",
		expected: []prompt.Check{
			{Name: "frontmatter", Status: prompt.StatusOK, Message: ""},
			{Name: "description", Status: prompt.StatusWarn, Message: "no description line found"},
			{Name: "completion-signal", Status: prompt.StatusOK, Message: ""},
		},
	},
	{
		name: "bare opening delimiter is unbalanced frontmatter",
		text: "---",
		expected: []prompt.Check{
			{Name: "frontmatter", Status: prompt.StatusWarn, Message: "unbalanced frontmatter delimiters"},
			{Name: "description", Status: prompt.StatusWarn, Message: "no description line found"},
			{Name: "completion-signal", Status: prompt.StatusFail,
				Message: "neither <COMPLETION_SIGNAL> nor <promise>COMPLETE</promise> found"},
		},
	},
}

func TestValidatePromptText(t *testing.T) {
	for _, tc := range validatePromptTextTests {
		t.Run(tc.name, func(t *testing.T) {
			assertChecks(t, prompt.ValidatePromptText(tc.text), tc.expected)
		})
	}
}

func assertChecks(t *testing.T, got, want []prompt.Check) {
	t.Helper()

	if len(got) != len(want) {
		t.Fatalf("expected %d checks, got %d: %+v", len(want), len(got), got)
	}

	for i := range want {
		if got[i] != want[i] {
			t.Errorf("check[%d] = %+v, want %+v", i, got[i], want[i])
		}
	}
}
