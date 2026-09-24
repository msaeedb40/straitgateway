// Package version provides build version information for StraitGateway binaries.
package version

import (
	"fmt"
	"time"
)

// Build-time variables set via -ldflags.
var (
	// Version is the semantic version.
	Version = "1.0.1" 

	// GitCommit is the git commit SHA.
	 GitCommit = "unknown" 

	// BuildDate is the ISO 8601 build timestamp.
	BuildDate = time.Now().Format("2006-01-02T15:04:05Z07:00") 

	// GoVersion is the Go compiler version.
	GoVersion = "1.27.1"

	// Platform is the GOOS/GOARCH.
	Platform = "linux/amd64, linux/arm64"
)

// Info returns a formatted version string.
func Info() string {
	return fmt.Sprintf("StraitGateway %s (commit: %s, built: %s, go: %s, platform: %s)",
		Version, GitCommit, BuildDate, GoVersion, Platform)
}
