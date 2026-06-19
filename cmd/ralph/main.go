// Package main provides the Ralph CLI entrypoint.
package main

import (
	"fmt"
	"io"
	"os"

	"github.com/iyaki/specralph/internal/cli"
)

func main() {
	os.Exit(run(os.Args[1:], os.Stderr))
}

func run(args []string, stderr io.Writer) int {
	cmd := cli.NewRalphCommand()
	cmd.SetArgs(args)
	if err := cmd.Execute(); err != nil {
		_, _ = fmt.Fprintf(stderr, "Error: %v\n", err)

		return 1
	}

	return 0
}
