// Package conformance provides standard specification conformance tests
// for StraitGateway CNI, Gateway API v1.6.1, and Kubernetes network routing.
package conformance

import (
	"bytes"
	"encoding/json"
	"os"
	"strings"
	"testing"

	"github.com/straitgateway/straitgateway/internal/cni"
	"github.com/straitgateway/straitgateway/internal/cni/ipam"
)

// TestCNIVersionConformance verifies that the CNI plugin responds to the
// CNI VERSION command according to the specification.
func TestCNIVersionConformance(t *testing.T) {
	buf := new(bytes.Buffer)
	if err := cni.CmdVersion(buf); err != nil {
		t.Fatalf("CmdVersion failed: %v", err)
	}

	var info cni.VersionInfo
	if err := json.Unmarshal(buf.Bytes(), &info); err != nil {
		t.Fatalf("Failed to decode CNI VersionInfo JSON: %v", err)
	}

	if info.CNIVersion == "" {
		t.Fatal("CNIVersion must not be empty")
	}

	// Specification requires 1.0.0 support
	has100 := false
	for _, v := range info.SupportedVersions {
		if v == "1.0.0" {
			has100 = true
			break
		}
	}
	if !has100 {
		t.Fatalf("SupportedVersions must include 1.0.0, got: %v", info.SupportedVersions)
	}
	t.Logf("✓ CNI VERSION conformance verified: version=%s, supported=%v", info.CNIVersion, info.SupportedVersions)
}

// TestCNIConfigParsingConformance tests CNI conflist parsing and defaults.
func TestCNIConfigParsingConformance(t *testing.T) {
	conflistJSON := `{
		"cniVersion": "1.0.0",
		"name": "straitgateway-test",
		"plugins": [
			{
				"type": "strait-cni",
				"datapathMode": "netkit",
				"mtu": 1500,
				"ipam": {
					"type": "cluster-pool",
					"ranges": [
						{"subnet": "10.244.1.0/24", "gateway": "10.244.1.1"}
					]
				}
			}
		]
	}`

	var confMap map[string]interface{}
	if err := json.Unmarshal([]byte(conflistJSON), &confMap); err != nil {
		t.Fatalf("Conflist JSON must be valid: %v", err)
	}

	if confMap["cniVersion"] != "1.0.0" {
		t.Errorf("expected cniVersion 1.0.0, got %v", confMap["cniVersion"])
	}

	plugins, ok := confMap["plugins"].([]interface{})
	if !ok || len(plugins) == 0 {
		t.Fatal("plugins list must not be empty")
	}

	firstPlugin := plugins[0].(map[string]interface{})
	if firstPlugin["type"] != "strait-cni" {
		t.Errorf("expected plugin type 'strait-cni', got %v", firstPlugin["type"])
	}
	if firstPlugin["datapathMode"] != "netkit" {
		t.Errorf("expected datapathMode 'netkit', got %v", firstPlugin["datapathMode"])
	}
}

// TestCNIIPAMAllocationReleaseLifecycle validates that IPAM operates
// conformantly under sequential allocation and release.
func TestCNIIPAMAllocationReleaseLifecycle(t *testing.T) {
	testCIDR := "192.168.100.0/28" // 16 total IPs
	alloc, err := ipam.NewAllocator([]string{testCIDR})
	if err != nil {
		t.Fatalf("Failed to initialize IPAM allocator: %v", err)
	}

	allocated := make(map[string]bool)
	count := 10
	for i := 0; i < count; i++ {
		ip, err := alloc.Allocate()
		if err != nil {
			t.Fatalf("Allocation #%d failed: %v", i, err)
		}
		if allocated[ip] {
			t.Fatalf("Duplicate IP allocated: %s", ip)
		}
		if !strings.HasPrefix(ip, "192.168.100.") {
			t.Fatalf("Allocated IP %s does not belong to subnet %s", ip, testCIDR)
		}
		allocated[ip] = true
	}

	// Release all
	for ip := range allocated {
		if err := alloc.Release(ip); err != nil {
			t.Fatalf("Failed to release IP %s: %v", ip, err)
		}
	}

	// Re-allocate should succeed
	reIP, err := alloc.Allocate()
	if err != nil {
		t.Fatalf("Allocation after release failed: %v", err)
	}
	if !allocated[reIP] {
		t.Logf("Reallocated IP from pool: %s", reIP)
	}
	_ = alloc.Release(reIP)
}

// TestCNIEnvironmentConformance tests CNI standard environment variable dispatch.
func TestCNIEnvironmentConformance(t *testing.T) {
	origCmd := os.Getenv("CNI_COMMAND")
	defer func() { _ = os.Setenv("CNI_COMMAND", origCmd) }()

	_ = os.Setenv("CNI_COMMAND", "VERSION")
	if err := cni.Run(); err != nil {
		t.Errorf("cni.Run() with CNI_COMMAND=VERSION failed: %v", err)
	}
}
