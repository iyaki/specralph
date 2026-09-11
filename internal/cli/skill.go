package cli

import (
	_ "embed"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
)

// skillText is the prompt-authoring skill embedded at build time from the
// canonical source at .agents/skills/specralph-prompts/SKILL.md (a symlink to
// skill.md in this package, which go:embed can read).
//
//go:embed skill.md
var skillText string

// NewSkillCommand creates the skill command for installing the bundled
// prompt-authoring skill.
func NewSkillCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "skill",
		Short: "Install the bundled prompt-authoring skill",
		Long:  `Install the bundled prompt-authoring skill into a project directory.`,
	}

	var force bool
	install := &cobra.Command{
		Use:   "install [dir]",
		Short: "Install the prompt-authoring skill into a project",
		Long: `Write the embedded specralph-prompts skill to <dir>/specralph-prompts/SKILL.md.
The directory defaults to .agents/skills. Use --force to overwrite an existing installation.`,
		Args: cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			dir := ".agents/skills"
			if len(args) == 1 {
				dir = args[0]
			}

			return runSkillInstall(cmd.OutOrStdout(), dir, force)
		},
	}
	install.Flags().BoolVar(&force, "force", false, "Overwrite an existing skill file")

	cmd.AddCommand(install)

	return cmd
}

// runSkillInstall writes the embedded skill to <dir>/specralph-prompts/SKILL.md.
func runSkillInstall(output io.Writer, dir string, force bool) error {
	const (
		dirPerm  = 0755
		filePerm = 0644
	)

	target := filepath.Join(dir, "specralph-prompts", "SKILL.md")
	if _, err := os.Stat(target); err == nil && !force {
		return fmt.Errorf("skill already exists at %s (use --force to overwrite)", target)
	}

	if err := os.MkdirAll(filepath.Dir(target), dirPerm); err != nil {
		return fmt.Errorf("failed to create skill directory: %w", err)
	}

	// #nosec G304 -- path is the user-provided install target
	if err := os.WriteFile(target, []byte(skillText), filePerm); err != nil {
		_ = os.Remove(target) // no partial file left behind

		return fmt.Errorf("failed to write skill file: %w", err)
	}

	_, err := fmt.Fprintf(output, "Installed specralph-prompts skill to %s\n", target)

	return err
}
