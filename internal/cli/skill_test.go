package cli_test

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/iyaki/specralph/internal/cli"
)

func TestSkillInstallDefaultDir(t *testing.T) {
	wd, err := os.Getwd()
	if err != nil {
		t.Fatalf("failed to get cwd: %v", err)
	}
	tmp := t.TempDir()
	if err := os.Chdir(tmp); err != nil {
		t.Fatalf("failed to chdir: %v", err)
	}
	t.Cleanup(func() {
		_ = os.Chdir(wd)
	})
	// A broken config must not affect the install: the command is config-independent.
	if err := os.WriteFile(filepath.Join(tmp, "ralph.toml"), []byte("this is not toml {{{"), 0644); err != nil {
		t.Fatalf("failed to write broken config: %v", err)
	}

	cmd := cli.NewSkillCommand()
	cmd.SetArgs([]string{"install"})
	var out bytes.Buffer
	cmd.SetOut(&out)

	if err := cmd.Execute(); err != nil {
		t.Fatalf("expected install to succeed, got: %v", err)
	}

	target := filepath.Join(".agents", "skills", "specralph-prompts", "SKILL.md")
	if !strings.Contains(out.String(), "Installed specralph-prompts skill to "+target) {
		t.Errorf("expected install message mentioning %q, got %q", target, out.String())
	}

	assertInstalledSkill(t, target)
}

func TestSkillInstallCustomDir(t *testing.T) {
	wd, err := os.Getwd()
	if err != nil {
		t.Fatalf("failed to get cwd: %v", err)
	}
	tmp := t.TempDir()
	if err := os.Chdir(tmp); err != nil {
		t.Fatalf("failed to chdir: %v", err)
	}
	t.Cleanup(func() {
		_ = os.Chdir(wd)
	})
	dir := filepath.Join(tmp, "custom", "skills")

	cmd := cli.NewSkillCommand()
	cmd.SetArgs([]string{"install", dir})
	var out bytes.Buffer
	cmd.SetOut(&out)

	if err := cmd.Execute(); err != nil {
		t.Fatalf("expected install to succeed, got: %v", err)
	}

	target := filepath.Join(dir, "specralph-prompts", "SKILL.md")
	if !strings.Contains(out.String(), "Installed specralph-prompts skill to "+target) {
		t.Errorf("expected install message mentioning %q, got %q", target, out.String())
	}

	assertInstalledSkill(t, target)
}

func TestSkillInstallCollisionWithoutForce(t *testing.T) {
	wd, err := os.Getwd()
	if err != nil {
		t.Fatalf("failed to get cwd: %v", err)
	}
	tmp := t.TempDir()
	if err := os.Chdir(tmp); err != nil {
		t.Fatalf("failed to chdir: %v", err)
	}
	t.Cleanup(func() {
		_ = os.Chdir(wd)
	})
	target := filepath.Join(tmp, ".agents", "skills", "specralph-prompts", "SKILL.md")
	if err := os.MkdirAll(filepath.Dir(target), 0755); err != nil {
		t.Fatalf("failed to create target dir: %v", err)
	}
	if err := os.WriteFile(target, []byte("original"), 0644); err != nil {
		t.Fatalf("failed to write existing skill: %v", err)
	}

	cmd := cli.NewSkillCommand()
	cmd.SetArgs([]string{"install"})
	relTarget := filepath.Join(".agents", "skills", "specralph-prompts", "SKILL.md")
	if err := cmd.Execute(); err == nil {
		t.Fatal("expected install without --force to fail on existing skill")
	} else if want := "skill already exists at " + relTarget +
		" (use --force to overwrite)"; !strings.Contains(err.Error(), want) {
		t.Errorf("expected collision error %q, got: %v", want, err)
	}

	content, err := os.ReadFile(target) // #nosec G304 -- test-controlled path
	if err != nil {
		t.Fatalf("failed to read existing skill: %v", err)
	}
	if string(content) != "original" {
		t.Errorf("expected original content untouched, got %q", content)
	}
}

func TestSkillInstallForceOverwrites(t *testing.T) {
	wd, err := os.Getwd()
	if err != nil {
		t.Fatalf("failed to get cwd: %v", err)
	}
	tmp := t.TempDir()
	if err := os.Chdir(tmp); err != nil {
		t.Fatalf("failed to chdir: %v", err)
	}
	t.Cleanup(func() {
		_ = os.Chdir(wd)
	})
	target := filepath.Join(tmp, ".agents", "skills", "specralph-prompts", "SKILL.md")
	if err := os.MkdirAll(filepath.Dir(target), 0755); err != nil {
		t.Fatalf("failed to create target dir: %v", err)
	}
	if err := os.WriteFile(target, []byte("original"), 0644); err != nil {
		t.Fatalf("failed to write existing skill: %v", err)
	}

	cmd := cli.NewSkillCommand()
	cmd.SetArgs([]string{"install", "--force"})
	var out bytes.Buffer
	cmd.SetOut(&out)

	if err := cmd.Execute(); err != nil {
		t.Fatalf("expected forced install to succeed, got: %v", err)
	}

	assertInstalledSkill(t, target)
}

func TestEmbeddedSkillMatchesCanonicalSource(t *testing.T) {
	canonical := "../../.agents/skills/specralph-prompts/SKILL.md"
	want, err := os.ReadFile(canonical) // #nosec G304 -- fixed repo-relative path
	if err != nil {
		t.Fatalf("failed to read canonical skill source: %v", err)
	}

	wd, err := os.Getwd()
	if err != nil {
		t.Fatalf("failed to get cwd: %v", err)
	}
	tmp := t.TempDir()
	if err := os.Chdir(tmp); err != nil {
		t.Fatalf("failed to chdir: %v", err)
	}
	t.Cleanup(func() {
		_ = os.Chdir(wd)
	})

	cmd := cli.NewSkillCommand()
	cmd.SetArgs([]string{"install"})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("expected install to succeed, got: %v", err)
	}

	target := filepath.Join(".agents", "skills", "specralph-prompts", "SKILL.md")
	got, err := os.ReadFile(target) // #nosec G304 -- test-controlled path
	if err != nil {
		t.Fatalf("failed to read installed skill: %v", err)
	}
	if string(got) != string(want) {
		t.Error("installed skill differs from the canonical .agents source (embed is stale or source moved)")
	}
}

func TestSkillHelpListsInstall(t *testing.T) {
	cmd := cli.NewSkillCommand()
	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetArgs([]string{"--help"})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("expected help to print, got: %v", err)
	}
	if !strings.Contains(out.String(), "install") {
		t.Errorf("expected help to list install, got %q", out.String())
	}
}

func assertInstalledSkill(t *testing.T, target string) {
	t.Helper()

	info, err := os.Stat(target)
	if err != nil {
		t.Fatalf("expected installed skill at %s: %v", target, err)
	}
	if info.IsDir() {
		t.Fatalf("expected a file at %s", target)
	}

	content, err := os.ReadFile(target) // #nosec G304 -- test-controlled path
	if err != nil {
		t.Fatalf("failed to read installed skill: %v", err)
	}
	if !strings.HasPrefix(string(content), "---\nname: specralph-prompts\n") {
		t.Errorf("expected installed skill to start with specralph-prompts frontmatter, got %q", content)
	}
}
