// Copyright 2026 straitgateway Authors
// SPDX-License-Identifier: Apache-2.0

package gateway

import (
	"context"
	"fmt"

	"go.uber.org/zap"
	"sigs.k8s.io/controller-runtime/pkg/client"
	gwapiv1 "sigs.k8s.io/gateway-api/apis/v1"

	"github.com/msaeedb40/straitgateway/dataplane/ir"
)

// Manager reconciles Gateway API resources into GatewayIR.
type Manager struct {
	client client.Client
	log    *zap.Logger
}

// New creates a new Gateway API manager.
func New(c client.Client, log *zap.Logger) *Manager {
	return &Manager{client: c, log: log}
}

// Reconcile reads all Gateway and Route resources and produces GatewayIR.
func (m *Manager) Reconcile(ctx context.Context) ([]ir.GatewayIR, error) {
	var gwList gwapiv1.GatewayList
	if err := m.client.List(ctx, &gwList); err != nil {
		return nil, fmt.Errorf("listing gateways: %w", err)
	}

	var result []ir.GatewayIR
	for _, gw := range gwList.Items {
		if string(gw.Spec.GatewayClassName) != GatewayClassName {
			continue
		}

		gwIR := ir.GatewayIR{
			GatewayName: gw.Name,
			Namespace:   gw.Namespace,
		}

		for _, listener := range gw.Spec.Listeners {
			lIR := ir.ListenerIR{
				Name:     string(listener.Name),
				Protocol: string(listener.Protocol),
				Port:     uint16(listener.Port),
			}
			gwIR.Listeners = append(gwIR.Listeners, lIR)
		}

		result = append(result, gwIR)
	}

	m.log.Info("gateway reconciliation complete", zap.Int("gateways", len(result)))
	return result, nil
}
