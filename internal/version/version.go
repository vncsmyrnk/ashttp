package version

import (
	"fmt"
	"runtime"
)

var (
	// Version is the semantic version of the application.
	Version = "dev"
	// GitCommit is the git commit hash from which the binary was built.
	GitCommit = "n/a"
	// BuildDate is the date when the binary was built.
	BuildDate = "unknown"
)

func Info() string {
	return fmt.Sprintf("ashttp %s (commit: %s, built: %s, go: %s)", Version, GitCommit, BuildDate, runtime.Version())
}

func Short() string {
	return Version
}
