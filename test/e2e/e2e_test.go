// Package e2e provides end-to-end tests for StraitGateway.
// Tests require a live kind cluster with StraitGateway deployed.
// Run with: SG_E2E_TEST=1 go test ./test/e2e/...
package e2e

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/straitgateway/straitgateway/internal/tgwd/api"
	"go.uber.org/zap"
)

func requireE2E(t *testing.T) {
	t.Helper()
	if os.Getenv("SG_E2E_TEST") == "" {
		t.Skip("set SG_E2E_TEST=1 to run e2e tests")
	}
}

func TestE2EStraitdHealthy(t *testing.T) {
	requireE2E(t)

	addr := os.Getenv("STRAITD_API_ADDR")
	if addr == "" {
		addr = "unix:///run/straitd/api.sock"
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	log, _ := zap.NewDevelopment()
	defer log.Sync()

	client, err := api.NewClient(ctx, log, addr)
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}
	defer client.Close()

	if err := client.Ping(ctx); err != nil {
		t.Fatalf("straitd unhealthy: %v", err)
	}

	t.Log("✓ straitd healthy on kind cluster")
}
