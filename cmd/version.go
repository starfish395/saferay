package cmd

import "fmt"

// These variables are set at build time via ldflags
var (
	version = "dev"
	commit  = "none"
	date    = "unknown"
	builtBy = "unknown"
)

func cmdVersion() {
	fmt.Printf("saferay %s\n", version)
	if version != "dev" {
		fmt.Printf("  commit:   %s\n", commit)
		fmt.Printf("  built:    %s\n", date)
		fmt.Printf("  built by: %s\n", builtBy)
	}
}

// Version returns the current version string
func Version() string {
	return version
}
