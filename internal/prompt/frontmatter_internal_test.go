//nolint:cyclop
package prompt

import (
	"testing"
)

//nolint:funlen
func TestParseFrontMatterEdgeCasesInternal(t *testing.T) {
	t.Run("--- alone returns empty settings and content unchanged", func(t *testing.T) {
		content := "---"
		settings, body, err := ParseFrontMatter(content)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
		if settings == nil {
			t.Fatal("expected non-nil settings")
		}
		if settings.Model != "" || settings.AgentMode != "" {
			t.Errorf("expected empty settings, got %+v", settings)
		}
		if body != "---" {
			t.Errorf("expected body \"---\", got %q", body)
		}
	})

	t.Run("--- followed by non-newline treated as text", func(t *testing.T) {
		content := "---foo"
		settings, body, err := ParseFrontMatter(content)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
		if body != "---foo" {
			t.Errorf("expected body unchanged, got %q", body)
		}
		_ = settings
	})

	t.Run("no closing --- returns content unchanged", func(t *testing.T) {
		content := "---\nmodel: gpt-4\nSome content without closing"
		settings, body, err := ParseFrontMatter(content)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
		if body != content {
			t.Errorf("expected content unchanged, got %q", body)
		}
		_ = settings
	})

	t.Run("CRLF line endings handled", func(t *testing.T) {
		content := "---\r\nmodel: gpt-4\r\n---\r\nBody text"
		settings, body, err := ParseFrontMatter(content)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
		if settings.Model != "gpt-4" {
			t.Errorf("expected Model \"gpt-4\", got %q", settings.Model)
		}
		if body != "Body text" {
			t.Errorf("expected body \"Body text\", got %q", body)
		}
	})

	t.Run("empty front matter block", func(t *testing.T) {
		content := "---\n---\nBody"
		settings, body, err := ParseFrontMatter(content)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
		if settings == nil {
			t.Fatal("expected non-nil settings")
		}
		if body != "Body" {
			t.Errorf("expected body \"Body\", got %q", body)
		}
	})
}
