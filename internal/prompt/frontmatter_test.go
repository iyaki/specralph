package prompt_test

import (
	"testing"

	"github.com/iyaki/specralph/internal/prompt"
)

func TestParseFrontMatter_HappyPaths(t *testing.T) {
	tests := []struct {
		name          string
		content       string
		expectedModel string
		expectedMode  string
		expectedBody  string
	}{
		{
			name: "Valid front matter with model and agent-mode",
			content: `---
model: gpt-4
agent-mode: planner
---
Prompt body`,
			expectedModel: "gpt-4",
			expectedMode:  "planner",
			expectedBody:  "Prompt body",
		},
		{
			name: "Valid front matter with only model",
			content: `---
model: gpt-3.5
---
Prompt body`,
			expectedModel: "gpt-3.5",
			expectedMode:  "",
			expectedBody:  "Prompt body",
		},
		{
			name: "Valid front matter with only agent-mode",
			content: `---
agent-mode: coder
---
Prompt body`,
			expectedModel: "",
			expectedMode:  "coder",
			expectedBody:  "Prompt body",
		},
		{
			name: "Front matter with unknown keys",
			content: `---
model: gpt-4
unknown: value
---
Prompt body`,
			expectedModel: "gpt-4",
			expectedMode:  "",
			expectedBody:  "Prompt body",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			runFrontMatterTest(t, tt.content, tt.expectedModel, tt.expectedMode, tt.expectedBody)
		})
	}
}

func TestParseFrontMatter_EdgeCases(t *testing.T) {
	tests := []struct {
		name          string
		content       string
		expectedModel string
		expectedMode  string
		expectedBody  string
	}{
		{
			name:          "No front matter",
			content:       `Just a prompt body`,
			expectedModel: "",
			expectedMode:  "",
			expectedBody:  "Just a prompt body",
		},
		{
			name: "Empty front matter",
			content: `---
---
Prompt body`,
			expectedModel: "",
			expectedMode:  "",
			expectedBody:  "Prompt body",
		},
		{
			name: "Front matter like structure but not at start",
			content: `
---
model: gpt-4
---
Prompt body`,
			expectedModel: "",
			expectedMode:  "",
			expectedBody: `
---
model: gpt-4
---
Prompt body`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			runFrontMatterTest(t, tt.content, tt.expectedModel, tt.expectedMode, tt.expectedBody)
		})
	}
}

func TestParseFrontMatter_Invalid(t *testing.T) {
	tests := []struct {
		name    string
		content string
	}{
		{
			name: "Invalid front matter (malformed YAML)",
			content: `---
model: [
---
Prompt body`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, _, err := prompt.ParseFrontMatter(tt.content)
			if err == nil {
				t.Errorf("Expected error, got nil")
			}
		})
	}
}

func runFrontMatterTest(t *testing.T, content, expectedModel, expectedMode, expectedBody string) {
	t.Helper()

	settings, body, err := prompt.ParseFrontMatter(content)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)

		return
	}

	if settings == nil {
		t.Errorf("Expected settings, got nil")

		return
	}

	if settings.Model != expectedModel {
		t.Errorf("Expected model %q, got %q", expectedModel, settings.Model)
	}

	if settings.AgentMode != expectedMode {
		t.Errorf("Expected agent-mode %q, got %q", expectedMode, settings.AgentMode)
	}

	if body != expectedBody {
		t.Errorf("Expected body %q, got %q", expectedBody, body)
	}
}
