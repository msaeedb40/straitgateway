package observability

import (
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/straitgateway/straitgateway/internal/datapath/maps"
	"github.com/straitgateway/straitgateway/internal/routing/bfd"
	"github.com/straitgateway/straitgateway/internal/routing/bgp"
	"github.com/straitgateway/straitgateway/internal/wireguard"
)

// MetricsProvider interfaces subsystem queries
type MetricsProvider interface {
	Collect() DataplaneMetrics
}

// MetricsCollector compiles dataplane counters and system metrics
type MetricsCollector struct {
	mu         sync.RWMutex
	mapMgr     *maps.MapManager
	bgpSpeaker *bgp.Speaker
	bfdMonitor *bfd.Monitor
	wgManager  *wireguard.Manager
	nodeName   string
}

// NewMetricsCollector creates a MetricsCollector
func NewMetricsCollector(
	mapMgr *maps.MapManager,
	bgpSpeaker *bgp.Speaker,
	bfdMonitor *bfd.Monitor,
	wgManager *wireguard.Manager,
	nodeName string,
) *MetricsCollector {
	if nodeName == "" {
		nodeName = "localhost"
	}
	return &MetricsCollector{
		mapMgr:     mapMgr,
		bgpSpeaker: bgpSpeaker,
		bfdMonitor: bfdMonitor,
		wgManager:  wgManager,
		nodeName:   nodeName,
	}
}

// Collect returns the aggregated dataplane metrics
func (c *MetricsCollector) Collect() DataplaneMetrics {
	c.mu.RLock()
	defer c.mu.RUnlock()

	var totalRx, totalTx, totalXdp, totalRouted, totalPodTx, totalDrops, totalLb uint64

	if c.mapMgr != nil {
		totalRx, _ = c.mapMgr.GetCounter(0)     // RX
		totalTx, _ = c.mapMgr.GetCounter(1)     // TX
		totalXdp, _ = c.mapMgr.GetCounter(2)    // XDP
		totalRouted, _ = c.mapMgr.GetCounter(3) // Routed
		totalPodTx, _ = c.mapMgr.GetCounter(4)  // Pod TX
		totalDrops, _ = c.mapMgr.GetCounter(5)  // Policy Drops
		totalLb, _ = c.mapMgr.GetCounter(6)     // Service LB
	}

	bgpActive := 0
	if c.bgpSpeaker != nil {
		for _, p := range c.bgpSpeaker.ListPeers() {
			if p.State == bgp.StateEstablished {
				bgpActive++
			}
		}
	}

	bfdActive := 0
	if c.bfdMonitor != nil {
		for _, s := range c.bfdMonitor.ListSessions() {
			if s.State == bfd.StateUp {
				bfdActive++
			}
		}
	}

	wgPeers := 0
	if c.wgManager != nil {
		wgPeers = len(c.wgManager.GetPeers())
	}

	totalPackets := totalRx + totalTx + totalXdp + totalPodTx
	totalBytes := totalPackets * 128 // Approximation when raw byte counting is per-flow

	return DataplaneMetrics{
		PacketsTotal:        totalPackets,
		BytesTotal:          totalBytes,
		DropsTotal:          totalDrops,
		PolicyDeniedTotal:   totalDrops,
		RouteLookupTotal:    totalRouted,
		RouteMissTotal:      0,
		ConntrackEntries:    0,
		NATTranslations:     0,
		BGPSessionsActive:   bgpActive,
		BFDSessionsActive:   bfdActive,
		WireGuardPeersCount: wgPeers,
		WireGuardRxBytes:    0,
		WireGuardTxBytes:    0,
		ServiceLBRequests:   totalLb,
		Timestamp:           time.Now().UTC(),
	}
}

// PrometheusExport formats metrics in standard Prometheus exposition text
func (c *MetricsCollector) PrometheusExport() string {
	m := c.Collect()
	var b strings.Builder

	fmt.Fprintf(&b, "# HELP straitgateway_packets_total Total packets processed by dataplane\n")
	fmt.Fprintf(&b, "# TYPE straitgateway_packets_total counter\n")
	fmt.Fprintf(&b, "straitgateway_packets_total{node=\"%s\"} %d\n", c.nodeName, m.PacketsTotal)

	fmt.Fprintf(&b, "# HELP straitgateway_bytes_total Total bytes processed by dataplane\n")
	fmt.Fprintf(&b, "# TYPE straitgateway_bytes_total counter\n")
	fmt.Fprintf(&b, "straitgateway_bytes_total{node=\"%s\"} %d\n", c.nodeName, m.BytesTotal)

	fmt.Fprintf(&b, "# HELP straitgateway_drops_total Total packet drops\n")
	fmt.Fprintf(&b, "# TYPE straitgateway_drops_total counter\n")
	fmt.Fprintf(&b, "straitgateway_drops_total{node=\"%s\"} %d\n", c.nodeName, m.DropsTotal)

	fmt.Fprintf(&b, "# HELP straitgateway_policy_denied_total Total packets blocked by security policy\n")
	fmt.Fprintf(&b, "# TYPE straitgateway_policy_denied_total counter\n")
	fmt.Fprintf(&b, "straitgateway_policy_denied_total{node=\"%s\"} %d\n", c.nodeName, m.PolicyDeniedTotal)

	fmt.Fprintf(&b, "# HELP straitgateway_route_lookup_total Total L3 route lookups performed\n")
	fmt.Fprintf(&b, "# TYPE straitgateway_route_lookup_total counter\n")
	fmt.Fprintf(&b, "straitgateway_route_lookup_total{node=\"%s\"} %d\n", c.nodeName, m.RouteLookupTotal)

	fmt.Fprintf(&b, "# HELP straitgateway_bgp_sessions Active BGP sessions\n")
	fmt.Fprintf(&b, "# TYPE straitgateway_bgp_sessions gauge\n")
	fmt.Fprintf(&b, "straitgateway_bgp_sessions{node=\"%s\"} %d\n", c.nodeName, m.BGPSessionsActive)

	fmt.Fprintf(&b, "# HELP straitgateway_bfd_sessions Active BFD sessions\n")
	fmt.Fprintf(&b, "# TYPE straitgateway_bfd_sessions gauge\n")
	fmt.Fprintf(&b, "straitgateway_bfd_sessions{node=\"%s\"} %d\n", c.nodeName, m.BFDSessionsActive)

	fmt.Fprintf(&b, "# HELP straitgateway_wireguard_peers Configured WireGuard overlay peers\n")
	fmt.Fprintf(&b, "# TYPE straitgateway_wireguard_peers gauge\n")
	fmt.Fprintf(&b, "straitgateway_wireguard_peers{node=\"%s\"} %d\n", c.nodeName, m.WireGuardPeersCount)

	fmt.Fprintf(&b, "# HELP straitgateway_service_lb_requests Total Service ClusterIP load balancer requests\n")
	fmt.Fprintf(&b, "# TYPE straitgateway_service_lb_requests counter\n")
	fmt.Fprintf(&b, "straitgateway_service_lb_requests{node=\"%s\"} %d\n", c.nodeName, m.ServiceLBRequests)

	return b.String()
}
