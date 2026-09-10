// Copyright 2026 straitgateway Authors
// SPDX-License-Identifier: Apache-2.0

// Package metrics defines all Prometheus metrics for straitgateway.
// All metric names use the prefix "straitgateway_".
package metrics

import "github.com/prometheus/client_golang/prometheus"

const namespace = "straitgateway"

var (
	// BPFMapEntries is the number of entries per BPF map.
	BPFMapEntries = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: namespace,
			Name:      "bpf_map_entries",
			Help:      "Current number of entries in a BPF map.",
		},
		[]string{"map_name"},
	)

	// PacketsTotal counts packets processed by the dataplane.
	PacketsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: namespace,
			Name:      "packets_total",
			Help:      "Total packets processed by the straitgateway dataplane.",
		},
		[]string{"direction", "action", "protocol"},
	)

	// DropsTotal counts dropped packets with reason.
	DropsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: namespace,
			Name:      "drops_total",
			Help:      "Total packets dropped by the straitgateway dataplane.",
		},
		[]string{"reason"},
	)

	// PolicyEvaluations counts policy rule evaluations.
	PolicyEvaluations = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: namespace,
			Name:      "policy_evaluations_total",
			Help:      "Total number of network policy rule evaluations.",
		},
		[]string{"action"},
	)

	// ServiceBackends tracks the number of active backends per service.
	ServiceBackends = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: namespace,
			Name:      "service_backends",
			Help:      "Number of active backends per Kubernetes service.",
		},
		[]string{"namespace", "service", "algorithm"},
	)

	// CompilationDuration tracks dataplane compilation latency.
	CompilationDuration = prometheus.NewHistogram(
		prometheus.HistogramOpts{
			Namespace: namespace,
			Name:      "compilation_duration_seconds",
			Help:      "Duration of dataplane compilation cycles.",
			Buckets:   prometheus.ExponentialBuckets(0.001, 2, 12),
		},
	)

	// GatewayListenerRoutes tracks the number of routes per Gateway listener.
	GatewayListenerRoutes = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: namespace,
			Name:      "gateway_listener_routes",
			Help:      "Number of routes attached to a Gateway listener.",
		},
		[]string{"gateway", "namespace", "listener", "protocol"},
	)

	// TransitTunnelBytes tracks bytes transferred per transit tunnel.
	TransitTunnelBytes = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: namespace,
			Name:      "transit_tunnel_bytes_total",
			Help:      "Total bytes transferred through transit tunnels.",
		},
		[]string{"cluster_id", "direction"},
	)

	// BGPPeerStatus tracks the current BGP session state per peer.
	BGPSessionState = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: namespace,
			Name:      "bgp_session_state",
			Help:      "BGP session state (1=Established, 0=not Established).",
		},
		[]string{"peer_address", "peer_asn"},
	)
)

// Register registers all straitgateway metrics with the default Prometheus registry.
func Register() {
	prometheus.MustRegister(
		BPFMapEntries,
		PacketsTotal,
		DropsTotal,
		PolicyEvaluations,
		ServiceBackends,
		CompilationDuration,
		GatewayListenerRoutes,
		TransitTunnelBytes,
		BGPSessionState,
	)
}
