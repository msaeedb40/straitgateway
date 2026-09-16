// Copyright 2026 straitgateway Authors
// SPDX-License-Identifier: Apache-2.0

// Package service reconciles Kubernetes Service + EndpointSlice resources
// into ServiceIR objects consumed by the dataplane compiler.
//
// Architectural invariant: this package ONLY produces ServiceIR.
// It NEVER touches BPF maps directly.
package service

import (
	"context"
	"fmt"
	"net/netip"
	"sync"

	"go.uber.org/zap"
	corev1 "k8s.io/api/core/v1"
	discoveryv1 "k8s.io/api/discovery/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/msaeedb40/straitgateway/dataplane/ir"
	sgtypes "github.com/msaeedb40/straitgateway/pkg/types"
)

// Manager reconciles Service + EndpointSlice resources into ServiceIR.
type Manager struct {
	client client.Client
	log    *zap.Logger
	mu     sync.RWMutex

	// services is the current set of computed ServiceIRs, keyed by namespace/name.
	services map[string]*ir.ServiceIR
	// generation is incremented on each reconcile cycle.
	generation ir.Generation
}

// New creates a new service Manager.
func New(c client.Client, log *zap.Logger) *Manager {
	return &Manager{
		client:   c,
		log:      log,
		services: make(map[string]*ir.ServiceIR),
	}
}

// Reconcile re-reads all Services and EndpointSlices and updates the ServiceIR set.
// Called by the Service and EndpointSlice controllers on any change.
func (m *Manager) Reconcile(ctx context.Context) ([]ir.ServiceIR, error) {
	var svcList corev1.ServiceList
	if err := m.client.List(ctx, &svcList); err != nil {
		return nil, fmt.Errorf("listing services: %w", err)
	}

	var epsList discoveryv1.EndpointSliceList
	if err := m.client.List(ctx, &epsList); err != nil {
		return nil, fmt.Errorf("listing endpoint slices: %w", err)
	}

	// Build a map of service key → endpoint slices.
	epsMap := make(map[string][]discoveryv1.EndpointSlice)
	for _, eps := range epsList.Items {
		svcName := eps.Labels["kubernetes.io/service-name"]
		if svcName == "" {
			continue
		}
		key := eps.Namespace + "/" + svcName
		epsMap[key] = append(epsMap[key], eps)
	}

	m.mu.Lock()
	defer m.mu.Unlock()
	m.generation++

	result := make([]ir.ServiceIR, 0, len(svcList.Items))

	for _, svc := range svcList.Items {
		if svc.Spec.ClusterIP == "None" {
			continue // skip headless services
		}
		key := svc.Namespace + "/" + svc.Name

		svcIR := m.buildServiceIR(svc, epsMap[key])
		svcIR.Generation = m.generation
		m.services[key] = svcIR
		result = append(result, *svcIR)
	}

	m.log.Info("service reconciliation complete",
		zap.Int("services", len(result)),
		zap.Int64("generation", int64(m.generation)),
	)

	return result, nil
}

// buildServiceIR builds a ServiceIR from a Kubernetes Service and its EndpointSlices.
func (m *Manager) buildServiceIR(svc corev1.Service, epss []discoveryv1.EndpointSlice) *ir.ServiceIR {
	svcIR := &ir.ServiceIR{
		ID: sgtypes.ServiceID{
			Namespace: svc.Namespace,
			Name:      svc.Name,
		},
		Algorithm:    sgtypes.LBAlgorithmMaglev,
		IsNodePort:   svc.Spec.Type == corev1.ServiceTypeNodePort,
		IsExternalLB: svc.Spec.Type == corev1.ServiceTypeLoadBalancer,
	}

	// Set service VIP.
	if svc.Spec.ClusterIP != "" && svc.Spec.ClusterIP != "None" {
		if addr, err := netip.ParseAddr(svc.Spec.ClusterIP); err == nil {
			if addr.Is4() {
				svcIR.VIPv4 = addr
			} else {
				svcIR.VIPv6 = addr
			}
		}
	}

	// Set service port.
	if len(svc.Spec.Ports) > 0 {
		svcIR.Port = uint16(svc.Spec.Ports[0].Port)
		switch svc.Spec.Ports[0].Protocol {
		case corev1.ProtocolUDP:
			svcIR.Protocol = sgtypes.ProtocolUDP
		case corev1.ProtocolSCTP:
			svcIR.Protocol = sgtypes.ProtocolSCTP
		default:
			svcIR.Protocol = sgtypes.ProtocolTCP
		}
		if svc.Spec.Type == corev1.ServiceTypeNodePort {
			svcIR.NodePort = uint16(svc.Spec.Ports[0].NodePort)
		}
	}

	// Collect backends from EndpointSlices.
	var backendID sgtypes.BackendID = 1
	for _, eps := range epss {
		var targetPort uint16 = svcIR.Port
		if len(eps.Ports) > 0 && eps.Ports[0].Port != nil {
			targetPort = uint16(*eps.Ports[0].Port)
		}
		for _, ep := range eps.Endpoints {
			if ep.Conditions.Ready != nil && !*ep.Conditions.Ready {
				continue
			}
			for _, addrStr := range ep.Addresses {
				addr, err := netip.ParseAddr(addrStr)
				if err != nil {
					continue
				}
				be := ir.BackendIR{
					ID:      backendID,
					Address: netip.AddrPortFrom(addr, targetPort),
					State:   sgtypes.BackendStateActive,
					Weight:  1,
				}
				svcIR.Backends = append(svcIR.Backends, be)
				backendID++
			}
		}
	}

	return svcIR
}
