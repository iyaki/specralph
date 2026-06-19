// Package cli provides CLI commands and execution flow for Specralph.
package cli

import (
	"github.com/spf13/cobra"

	"github.com/iyaki/specralph/internal/config"
)

const maxPositionalArgs = 2

// NewRalphCommand creates the root command for Specralph.
func NewRalphCommand() *cobra.Command {
	var cfg config.Config

	cmd := &cobra.Command{
		Use:   "ralph [options] [prompt] [scope]",
		Short: "POSIX-compliant AI Agentic Loop runner for spec-driven development",
		Long: `Specralph is a POSIX-compliant AI agentic loop shell runner.
It is a Ralph-Wiggum inspired spec-driven agent runner.
It loads prompts from files (with optional inline overrides) and comes with build/plan presets.

The loop runs until the agent emits <promise>COMPLETE</promise> or max iterations is reached.
When writing custom prompts, include <promise>COMPLETE</promise> or <COMPLETION_SIGNAL> at the end to signal completion.

Custom Prompt Example:
  ralph --prompt "Implement feature X. When done, output <promise>COMPLETE</promise>"
  ralph --prompt "Task X, Y and Z. Output <COMPLETION_SIGNAL> when everything is done"

For extended documentation, examples, and configuration options, visit https://github.com/iyaki/specralph.`,
		SilenceUsage:  true,
		SilenceErrors: true,
		Example: `  ralph build
  ralph plan my-feature
  ralph --max-iterations 10 build
  ralph --prompt "Custom prompt text"
  echo "prompt from stdin" | ralph -`,
		Args: cobra.MaximumNArgs(maxPositionalArgs),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runCommandLogic(cmd, args, &cfg)
		},
	}

	setupSharedFlags(cmd, &cfg)

	// Register subcommands
	cmd.AddCommand(NewInitCommand())
	cmd.AddCommand(NewRunCommand())
	cmd.AddCommand(NewVersionCommand())
	cmd.AddCommand(NewPromptsCommand())

	return cmd
}
