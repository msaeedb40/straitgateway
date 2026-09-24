// Package integration provides integration tests for StraitGateway running
// on a live kind cluster. These tests require a running cluster and are
// skipped unless the SG_INTEGRATION_TEST environment variable is set.
package integration

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/straitgateway/straitgateway/internal/cni/ipam"
	"github.com/straitgateway/straitgateway/internal/tgwd/api"
	"go.uber.org/zap"
)

// requireIntegration skips the test if not in integration mode.
func requireIntegration(t *testing.T) {
	t.Helper()
	if os.Getenv("SG_INTEGRATION_TEST") == "" {
		t.Skip("set SG_INTEGRATION_TEST=1 to run integration tests")
	}
}

func apiAddr() string {
	if a := os.Getenv("STRAITD_API_ADDR"); a != "" {
		return a
	}
	return "unix:///run/straitd/api.sock"
}

// TestStraitdPing tests that straitd is running and reachable.
func TestStraitdPing(t *testing.T) {
	requireIntegration(t)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	log, _ := zap.NewDevelopment()
	defer log.Sync()

	client, err := api.NewClient(ctx, log, apiAddr())
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}
	defer client.Close()

	if err := client.Ping(ctx); err != nil {
		t.Fatalf("Ping: %v — is straitd running?", err)
	}
	t.Log("✓ straitd is reachable and healthy")
}

// TestIPAMAllocateRelease is an integration test for the IPAM allocator
// against real node CIDRs. Can run without a cluster.
func TestIPAMAllocateRelease(t *testing.T) {
	cidr := os.Getenv("SG_POD_CIDR")
	if cidr == "" {
		cidr = "10.244.0.0/24"
	}

	alloc, err := ipam.NewAllocator([]string{cidr})
	if err != nil {
		t.Fatalf("NewAllocator(%q): %v", cidr, err)
	}

	var addrs []string
	for i := 0; i < 10; i++ {
		addr, err := alloc.Allocate()
		if err != nil {
			t.Fatalf("Allocate #%d: %v", i, err)
		}
		addrs = append(addrs, addr)
	}

	for _, addr := range addrs {
		if err := alloc.Release(addr); err != nil {
			t.Errorf("Release(%q): %v", addr, err)
		}
	}
	t.Logf("✓ allocated and released %d addresses from %s", len(addrs), cidr)
}
