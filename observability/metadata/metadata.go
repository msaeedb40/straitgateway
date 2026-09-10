// Copyright 2026 straitgateway Authors
// SPDX-License-Identifier: Apache-2.0

// Package metadata defines the canonical 11-attribute observability model
// for straitgateway. All telemetry signals (metrics, traces, logs, flows)
// MUST include these attributes for correlation.
package metadata

import "fmt"

// CanonicalAttributes is the 11-attribute canonical observability model.
// All straitgateway telemetry (metrics, traces, logs, flow events) must
// include these attributes to enable correlated analysis.
type CanonicalAttributes struct {
	// 1. cluster_name: Kubernetes cluster name.
	ClusterName string `json:"cluster_name"`
	// 2. node_name: Kubernetes node name.
	NodeName string `json:"node_name"`
	// 3. namespace: Kubernetes namespace of the source/destination workload.
	Namespace string `json:"namespace"`
	// 4. pod_name: Kubernetes pod name.
	PodName string `json:"pod_name"`
	// 5. workload_name: Kubernetes workload name (Deployment/StatefulSet/DaemonSet).
	WorkloadName string `json:"workload_name"`
	// 6. identity: 32-bit straitgateway security identity.
	Identity uint32 `json:"identity"`
	// 7. segment_id: 32-bit transit segment ID.
	SegmentID uint32 `json:"segment_id"`
	// 8. direction: "ingress" or "egress".
	Direction string `json:"direction"`
	// 9. protocol: L4 protocol ("TCP", "UDP", "ICMP", "SCTP").
	Protocol string `json:"protocol"`
	// 10. action: Policy action ("allow", "deny", "reject").
	Action string `json:"action"`
	// 11. drop_reason: Drop reason code (empty string if allowed).
	DropReason string `json:"drop_reason,omitempty"`
}

// OTelAttributes returns the CanonicalAttributes as OpenTelemetry key-value pairs.
// Used when exporting traces and metrics via OTLP.
func (a *CanonicalAttributes) OTelAttributes() map[string]string {
	return map[string]string{
		"sg.cluster_name":  a.ClusterName,
		"sg.node_name":     a.NodeName,
		"sg.namespace":     a.Namespace,
		"sg.pod_name":      a.PodName,
		"sg.workload_name": a.WorkloadName,
		"sg.identity":      fmt.Sprintf("%d", a.Identity),
		"sg.segment_id":    fmt.Sprintf("%d", a.SegmentID),
		"sg.direction":     a.Direction,
		"sg.protocol":      a.Protocol,
		"sg.action":        a.Action,
		"sg.drop_reason":   a.DropReason,
	}
}

// PrometheusLabels returns the CanonicalAttributes as Prometheus label key-value pairs.
// All straitgateway metrics use these labels with the "straitgateway_" prefix.
func (a *CanonicalAttributes) PrometheusLabels() []string {
	return []string{
		"cluster_name", a.ClusterName,
		"node_name", a.NodeName,
		"namespace", a.Namespace,
		"pod_name", a.PodName,
		"workload_name", a.WorkloadName,
		"direction", a.Direction,
		"protocol", a.Protocol,
		"action", a.Action,
	}
}

// LogFields returns the CanonicalAttributes as zap structured log fields.
func (a *CanonicalAttributes) LogFields() map[string]interface{} {
	return map[string]interface{}{
		"cluster_name":  a.ClusterName,
		"node_name":     a.NodeName,
		"namespace":     a.Namespace,
		"pod_name":      a.PodName,
		"workload_name": a.WorkloadName,
		"identity":      a.Identity,
		"segment_id":    a.SegmentID,
		"direction":     a.Direction,
		"protocol":      a.Protocol,
		"action":        a.Action,
		"drop_reason":   a.DropReason,
	}
}
