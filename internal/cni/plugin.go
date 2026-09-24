// Package cni implements the StraitGateway CNI plugin.
//
// strait-cni integrates StraitGateway with the Kubernetes pod lifecycle.
// It handles CNI ADD, DEL, CHECK, and VERSION operations, creates Netkit
// devices for pod-to-node connectivity, and performs IPAM allocation.
//
// CNI Plugin lifecycle:
//
//	Pod creation → ADD → IPAM allocation → Netkit creation → routes → policy → READY
//	Pod deletion → DEL → remove routes → release IP → destroy Netkit
package cni

import (
	"fmt"
	"os"
)

// CNIVersion is the CNI specification version implemented.
const CNIVersion = "1.0.0"

// PluginName is the name of the CNI plugin.
const PluginName = "strait-cni"

// NetConf is the CNI network configuration passed by the runtime.
type NetConf struct {
	CNIVersion string     `json:"cniVersion"`
	Name       string     `json:"name"`
	Type       string     `json:"type"`
	IPAM       IPAMConfig `json:"ipam,omitempty"`
	MTU           int        `json:"mtu,omitempty"`
	DatapathMode  string     `json:"datapathMode,omitempty"`
	StraitDSocket string     `json:"straitdSocket,omitempty"`
}

// IPAMConfig holds IPAM configuration.
type IPAMConfig struct {
	Type   string `json:"type,omitempty"`
	Ranges []struct {
		Subnet  string `json:"subnet"`
		Gateway string `json:"gateway,omitempty"`
	} `json:"ranges,omitempty"`
}

// Run is the CNI plugin entry point. It dispatches to ADD, DEL, CHECK, or VERSION
// based on the CNI_COMMAND environment variable.
func Run() error {
	cmd := os.Getenv("CNI_COMMAND")

	switch cmd {
	case "ADD":
		return CmdAdd(os.Stdin)
	case "DEL":
		return CmdDel(os.Stdin)
	case "CHECK":
		return CmdCheck(os.Stdin)
	case "VERSION":
		return CmdVersion(os.Stdout)
	case "GC":
		gc := NewGarbageCollector(nil)
		_, err := gc.Collect()
		return err
	case "STATUS":
		return OutputStatus()
	default:
		return fmt.Errorf("unknown CNI_COMMAND: %q", cmd)
	}
}

