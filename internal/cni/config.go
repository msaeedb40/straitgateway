package cni

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

// ConflistPlugin represents an individual plugin entry within a CNI conflist
type ConflistPlugin struct {
	Type         string                 `json:"type"`
	SubnetV4     string                 `json:"subnetV4,omitempty"`
	SubnetV6     string                 `json:"subnetV6,omitempty"`
	MTU          int                    `json:"mtu,omitempty"`
	StateFile    string                 `json:"stateFile,omitempty"`
	IPAM         map[string]interface{} `json:"ipam,omitempty"`
	Capabilities map[string]interface{} `json:"capabilities,omitempty"`
}

// CNIConflist represents the top-level standard Kubernetes CNI .conflist document
type CNIConflist struct {
	CNIVersion string           `json:"cniVersion"`
	Name       string           `json:"name"`
	Plugins    []ConflistPlugin `json:"plugins"`
}

// DefaultCNIConfigDir is standard location for Kubernetes CNI conflist files
const DefaultCNIConfigDir = "/etc/cni/net.d"

// DefaultCNIConfigName is standard filename for StraitGateway CNI conflist
const DefaultCNIConfigName = "10-straitgateway.conflist"

// GenerateConflist creates formatted JSON bytes for standard StraitGateway CNI deployment
func GenerateConflist(subnetV4, subnetV6 string, mtu int) ([]byte, error) {
	if subnetV4 == "" {
		subnetV4 = "10.244.0.0/16"
	}
	if mtu == 0 {
		mtu = 1500
	}

	conf := CNIConflist{
		CNIVersion: "1.0.0",
		Name:       "straitgateway",
		Plugins: []ConflistPlugin{
			{
				Type:     "sg-cni",
				SubnetV4: subnetV4,
				SubnetV6: subnetV6,
				MTU:      mtu,
				IPAM: map[string]interface{}{
					"subnetV4":  subnetV4,
					"subnetV6":  subnetV6,
					"stateFile": "/var/lib/cni/straitgateway/ipam.json",
				},
			},
			{
				Type: "portmap",
				Capabilities: map[string]interface{}{
					"portMappings": true,
				},
			},
		},
	}

	return json.MarshalIndent(conf, "", "  ")
}

// InstallCNIConfig writes 10-straitgateway.conflist into the target directory
func InstallCNIConfig(targetDir, subnetV4, subnetV6 string, mtu int) error {
	if targetDir == "" {
		targetDir = DefaultCNIConfigDir
	}
	if err := os.MkdirAll(targetDir, 0755); err != nil {
		return fmt.Errorf("create cni config dir %s: %w", targetDir, err)
	}

	data, err := GenerateConflist(subnetV4, subnetV6, mtu)
	if err != nil {
		return fmt.Errorf("generate conflist: %w", err)
	}

	targetPath := filepath.Join(targetDir, DefaultCNIConfigName)
	return os.WriteFile(targetPath, data, 0644)
}
