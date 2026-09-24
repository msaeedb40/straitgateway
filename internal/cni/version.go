// Package cni implements the StraitGateway CNI plugin.
package cni

import (
	"encoding/json"
	"io"
	"os"
)

// VersionInfo holds version information returned by the CNI VERSION command.
type VersionInfo struct {
	CNIVersion        string   `json:"cniVersion"`
	SupportedVersions []string `json:"supportedVersions"`
}

// SupportedVersions lists all CNI specification versions supported by strait-cni.
var SupportedVersions = []string{"0.4.0", "1.0.0", "1.1.0"}

// CmdVersion handles the CNI VERSION command, encoding version info to stdout.
func CmdVersion(stdout io.Writer) error {
	if stdout == nil {
		stdout = os.Stdout
	}
	info := VersionInfo{
		CNIVersion:        CNIVersion,
		SupportedVersions: SupportedVersions,
	}
	return json.NewEncoder(stdout).Encode(info)
}
