package controller

import (
	"context"
	"testing"

	"github.com/straitgateway/straitgateway/internal/tgwd/api"
	"go.uber.org/zap"
)

func TestControllerReconcilers(t *testing.T) {
	log := zap.NewNop()
	ctx := context.Background()

	// 1. Node Reconciler
	nodeRec := newNodeReconciler(log)
	nodeRec.RegisterNode("worker-1", []string{"198.51.100.0/24"})
	if err := nodeRec.reconcile(ctx); err != nil {
		t.Fatalf("nodeReconciler.reconcile failed: %v", err)
	}

	// 2. Service Reconciler
	svcRec := newServiceReconciler(log, nil)
	svcRec.StageService("10.96.0.1", 443, "TCP", []*api.BackendEntry{
		{Addr: "198.51.100.10", Port: 6443, Active: true},
	})
	if err := svcRec.reconcile(ctx); err != nil {
		t.Fatalf("serviceReconciler.reconcile failed: %v", err)
	}

	// 3. Gateway & Policy Reconcilers
	gwRec := newGatewayReconciler(log, nil)
	if err := gwRec.reconcile(ctx); err != nil {
		t.Fatalf("gatewayReconciler.reconcile failed: %v", err)
	}

	polRec := newPolicyReconciler(log, nil)
	if err := polRec.reconcile(ctx); err != nil {
		t.Fatalf("policyReconciler.reconcile failed: %v", err)
	}
}
