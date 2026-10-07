package version

import (
	"fmt"
	"runtime"
)

var (
	// Version holds the semantic version of StraitGateway.
	Version = "1.0.0"
	// GitCommit holds the git commit hash at build time.
	GitCommit = "unknown"
	// BuildDate holds the UTC date string of the build.
	BuildDate = "unknown"
)

// Info returns human-readable build and version details.
func Info() string {
	return fmt.Sprintf("StraitGateway v%s (commit: %s, built: %s, go: %s, os/arch: %s/%s)",
		Version, GitCommit, BuildDate, runtime.Version(), runtime.GOOS, runtime.GOARCH)
}
