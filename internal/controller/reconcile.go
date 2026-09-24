// Package controller — reconcile.go
// Wires all reconcilers into the main controller loop.
package controller

import (
	"context"
	"os"
	"time"

	"github.com/straitgateway/straitgateway/internal/tgwd/api"
	"go.uber.org/zap"
)

// reconcileLoop runs all reconcilers on a configurable interval.
// In production this will be replaced by controller-runtime event-driven
// reconciliation, but the interval loop is useful for initial bringup.
func (c *Controller) reconcileLoop(ctx context.Context) {
	socketPath := os.Getenv("STRAITD_API_SOCKET")
	if socketPath == "" {
		socketPath = "unix:///run/straitd/api.sock"
	}

	// Connect to StraitD API client (Finding #2)
	dialCtx, cancel := context.WithTimeout(ctx, 2*time.Second)
	client, err := api.NewClient(dialCtx, c.log, socketPath)
	cancel()
	if err != nil {
		c.log.Warn("reconcile: straitd client connection unavailable, running in disconnected mode", zap.Error(err))
	} else {
		defer client.Close()
	}

	reconcilers := []reconciler{
		newNodeReconciler(c.log),
		newServiceReconciler(c.log, client),
		newGatewayReconciler(c.log, client),
		newPolicyReconciler(c.log, client),
	}

	ticker := time.NewTicker(15 * time.Second)
	defer ticker.Stop()

	c.log.Info("reconcile loop started", zap.Duration("interval", 15*time.Second))

	// Run once immediately on startup
	c.runReconcilers(ctx, reconcilers)

	for {
		select {
		case <-ctx.Done():
			c.log.Info("reconcile loop stopped")
			return
		case <-ticker.C:
			c.runReconcilers(ctx, reconcilers)
		}
	}
}

func (c *Controller) runReconcilers(ctx context.Context, reconcilers []reconciler) {
	for _, r := range reconcilers {
		if err := r.reconcile(ctx); err != nil {
			c.log.Error("reconciler error", zap.Error(err))
		}
	}
}
