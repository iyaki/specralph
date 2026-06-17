package cli

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/iyaki/ralphex/internal/buildversion"
)

// NewVersionCommand creates the version subcommand.
func NewVersionCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "Print the version number of Ralph CLI",
		Long:  `Print the version number of Ralph CLI.`,
		Run: func(_ *cobra.Command, _ []string) {
			fmt.Println(buildversion.String())
		},
	}
}
