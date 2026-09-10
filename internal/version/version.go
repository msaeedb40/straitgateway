// Copyright 2026 straitgateway Authors
// SPDX-License-Identifier: Apache-2.0

// Package version holds build-time version information.
// Populated via -ldflags during compilation.
package version

import (
	"fmt"
	"runtime"
)

var (
	// Version is the semantic version (e.g. v1.0.0).
	Major int = 1
	Minor int = 0
	Patch int = 0

	Version = fmt.Sprintf("v%d.%d.%d", Major, Minor, Patch)
	// Commit is the git commit SHA.
	Commit = "none"
	// BuildDate is the build timestamp.
	BuildDate = "unknown"
	// GoVersion is the Go runtime version used to build the binary.
	GoVersion = runtime.Version()
)

// String returns a human-readable version string.
func String() string {
	return fmt.Sprintf("straitgateway %s (commit: %s, built: %s, go: %s)", Version, Commit, BuildDate, GoVersion)
}
