package cni_test

import (
	"bytes"
	"testing"

	"github.com/straitgateway/straitgateway/internal/cni"
)

func TestHostInterfaceName(t *testing.T) {
	name1 := cni.HostInterfaceName("abcdef1234567890")
	if name1 != "sg-abcdef12345" {
		t.Errorf("got %s, want sg-abcdef12345", name1)
	}

	name2 := cni.HostInterfaceName("short")
	if name2 != "sg-short" {
		t.Errorf("got %s, want sg-short", name2)
	}
}

func TestParseCNIEnv(t *testing.T) {
	t.Setenv("CNI_COMMAND", "ADD")
	t.Setenv("CNI_CONTAINERID", "container-123")
	t.Setenv("CNI_NETNS", "/proc/1/ns/net")
	t.Setenv("CNI_IFNAME", "eth0")
	t.Setenv("CNI_ARGS", "K8S_POD_NAME=test-pod;K8S_POD_NAMESPACE=kube-system")

	env, err := cni.ParseCNIEnv()
	if err != nil {
		t.Fatalf("ParseCNIEnv failed: %v", err)
	}

	if env.Command != "ADD" {
		t.Errorf("Command: got %s, want ADD", env.Command)
	}
	if env.ContainerID != "container-123" {
		t.Errorf("ContainerID: got %s, want container-123", env.ContainerID)
	}
	if env.ParsedArgs.PodName != "test-pod" {
		t.Errorf("PodName: got %s, want test-pod", env.ParsedArgs.PodName)
	}
	if env.ParsedArgs.PodNamespace != "kube-system" {
		t.Errorf("PodNamespace: got %s, want kube-system", env.ParsedArgs.PodNamespace)
	}
}

func TestCmdDelIdempotent(t *testing.T) {
	t.Setenv("CNI_COMMAND", "DEL")
	t.Setenv("CNI_CONTAINERID", "nonexistent-container")
	t.Setenv("CNI_NETNS", "/nonexistent/netns")

	netConfJSON := `{"cniVersion": "1.0.0", "name": "strait-net", "type": "strait-cni"}`
	stdin := bytes.NewBufferString(netConfJSON)

	// CNI DEL must always succeed idempotently even if interface and netns do not exist
	if err := cni.CmdDel(stdin); err != nil {
		t.Fatalf("CmdDel should be idempotent and return nil, got: %v", err)
	}
}

func TestCmdAddRule3NoRFC1918Assumption(t *testing.T) {
	t.Setenv("CNI_COMMAND", "ADD")
	t.Setenv("CNI_CONTAINERID", "container-test-add")
	t.Setenv("CNI_NETNS", "/proc/1/ns/net")
	t.Setenv("CNI_IFNAME", "eth0")

	// When no IPAM ranges and no POD_CIDR are specified, CmdAdd must fail with Rule 3 error, NOT assume 10.244.0.0/16
	netConfJSON := `{"cniVersion": "1.0.0", "name": "strait-net", "type": "strait-cni"}`
	stdin := bytes.NewBufferString(netConfJSON)

	err := cni.CmdAdd(stdin)
	if err == nil {
		t.Fatalf("expected error due to Rule 3 (no CIDRs provided), got nil")
	}
}

