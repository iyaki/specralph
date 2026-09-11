package prompt

import (
	"strings"

	"gopkg.in/yaml.v3"
)

const (
	frontMatterDelim    = "---"
	frontMatterDelimLen = len(frontMatterDelim)
)

// FrontMatterSettings holds the configuration extracted from the front matter.
type FrontMatterSettings struct {
	Model       string `yaml:"model,omitempty"`
	AgentMode   string `yaml:"agent-mode,omitempty"`
	Description string `yaml:"description,omitempty"`
}

// ParseFrontMatter parses the YAML front matter from a markdown string.
// It returns the parsed settings, the body with front matter stripped, and an error if parsing fails.
func ParseFrontMatter(content string) (*FrontMatterSettings, string, error) {
	frontMatter, body, ok := splitFrontMatter(content)
	if !ok {
		return &FrontMatterSettings{}, content, nil
	}

	var settings FrontMatterSettings
	if err := yaml.Unmarshal([]byte(frontMatter), &settings); err != nil {
		return nil, "", err
	}

	return &settings, strings.TrimSpace(body), nil
}

// splitFrontMatter splits content into raw front matter and body. ok is false
// when content does not open with a front matter delimiter or the closing
// delimiter is missing (both cases are treated as plain text).
func splitFrontMatter(content string) (frontMatter, body string, ok bool) {
	// Front matter must start with "---" at the very beginning of the file.
	if !strings.HasPrefix(content, frontMatterDelim) {
		return "", content, false
	}

	// It must be followed by a newline. Content is exactly "---" (no closing),
	// or the delimiter is followed by something else (e.g. "---foo"): plain text.
	if len(content) <= frontMatterDelimLen {
		return "", content, false
	}
	if charAfter := content[frontMatterDelimLen]; charAfter != '\n' && charAfter != '\r' {
		return "", content, false
	}

	// Search for the closing delimiter.
	// We look for "\n---". This covers both "\n---" and "\r\n---" (partially).
	rest := content[frontMatterDelimLen:]
	idxLF := strings.Index(rest, "\n---")

	if idxLF == -1 {
		// No closing delimiter found
		return "", content, false
	}

	closeIdx := idxLF
	delimLen := 4

	// Check if it is CRLF
	if idxLF > 0 && rest[idxLF-1] == '\r' {
		closeIdx = idxLF - 1
		delimLen = 5
	}

	return rest[:closeIdx], content[frontMatterDelimLen+closeIdx+delimLen:], true
}

// opensFrontMatter reports whether content begins with a front matter opening
// delimiter, even if it is never closed.
func opensFrontMatter(content string) bool {
	if !strings.HasPrefix(content, frontMatterDelim) {
		return false
	}
	if len(content) == frontMatterDelimLen {
		return true
	}
	charAfter := content[frontMatterDelimLen]

	return charAfter == '\n' || charAfter == '\r'
}
