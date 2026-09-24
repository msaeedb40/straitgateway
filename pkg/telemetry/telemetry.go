// Package telemetry provides shared metric names and telemetry constants for StraitGateway.
package telemetry

// Metric names published across StraitGateway subsystems.
const (
	MetricDatapathPacketsTotal = "straitgateway_datapath_packets_total"
	MetricDatapathBytesTotal   = "straitgateway_datapath_bytes_total"
	MetricDatapathDropsTotal   = "straitgateway_datapath_drops_total"
	MetricCNIAddDuration       = "straitgateway_cni_add_duration_seconds"
	MetricCNIDelDuration       = "straitgateway_cni_del_duration_seconds"
	MetricIPAMAllocatedTotal   = "straitgateway_ipam_allocated_ips"
	MetricIPAMAvailableTotal   = "straitgateway_ipam_available_ips"
	MetricServiceBackendsTotal = "straitgateway_service_backends_count"
	MetricTransitPeersTotal    = "straitgateway_transit_peers_count"
)

// Label names used in telemetry metrics.
const (
	LabelNode      = "node"
	LabelNamespace = "namespace"
	LabelPod       = "pod"
	LabelProtocol  = "protocol"
	LabelDirection = "direction"
	LabelReason    = "reason"
)
