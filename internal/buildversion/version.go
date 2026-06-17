// Package buildversion provides build-time version information for the Ralph CLI.
package buildversion

import "fmt"

// Build-time variables injected via linker flags.
var (
	// Version is the semantic version string (e.g., "1.2.3").
	Version = "dev"

	// Commit is the git commit hash at build time.
	Commit = "unknown"

	// Date is the build date in RFC3339 format.
	Date = "unknown"
)

// Info holds complete version information.
type Info struct {
	Version string
	Commit  string
	Date    string
}

// Get returns the current version information.
func Get() Info {
	return Info{
		Version: Version,
		Commit:  Commit,
		Date:    Date,
	}
}

// String returns the formatted version string.
func String() string {
	return fmt.Sprintf("ralph v%s", Version)
}
