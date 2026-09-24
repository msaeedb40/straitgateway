// Package controller implements the StraitGateway Kubernetes controller.
//
// The SG Controller watches Kubernetes resources (Gateway API, Services,
// Endpoints, Nodes, StraitGateway CRDs) and translates desired networking
// state into configuration consumed by StraitD on each node.
package controller

import (
	"context"
	"fmt"

	"go.uber.org/zap"
)

// Controller is the StraitGateway Kubernetes controller.
type Controller struct {
	log *zap.Logger
}

// NewController creates and initializes the SG Controller.
func NewController() (*Controller, error) {
	log, err := zap.NewProduction()
	if err != nil {
		return nil, fmt.Errorf("failed to initialize logger: %w", err)
	}
	return &Controller{log: log}, nil
}

// Run starts the controller manager and blocks until ctx is cancelled.
func (c *Controller) Run(ctx context.Context) error {
	c.log.Info("sg-controller starting")

	// Start the reconcile loop (drives node, service, gateway, policy reconcilers).
	go c.reconcileLoop(ctx)

	c.log.Info("sg-controller running")
	<-ctx.Done()
	c.log.Info("sg-controller stopped")
	return nil
}

