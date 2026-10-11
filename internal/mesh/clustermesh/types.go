package clustermesh

import (
	"net"
	"sync"
	"time"

	"github.com/straitgateway/straitgateway/api/v1alpha1"
)

// Config defines the configuration for ClusterMesh
type Config struct {
	LocalClusterID         string `json:"localClusterId"`
	LocalWireGuardEndpoint string `json:"localWireguardEndpoint"`
	LocalWireGuardPubkey   string `json:"localWireguardPubkey"`
	GlobalCIDR             string `json:"globalCIDR"`     // e.g. "240.0.0.0/8"
	DomainSuffix           string `json:"domainSuffix"`   // e.g. "global.straitgateway"
	DNSPort                int    `json:"dnsPort"`        // e.g. 5353
}

// DefaultConfig provides standard ClusterMesh defaults
func DefaultConfig() Config {
	return Config{
		LocalClusterID:         "cluster-local",
		LocalWireGuardEndpoint: "0.0.0.0:51820",
		GlobalCIDR:             "240.0.0.0/8",
		DomainSuffix:           "global.straitgateway",
		DNSPort:                5353,
	}
}

// ClusterPeerState tracks the active runtime state of a remote peered cluster
type ClusterPeerState struct {
	ClusterID         string
	WireGuardEndpoint string
	WireGuardPubkey   string
	PodCIDRs          []net.IPNet
	ServiceCIDRs      []net.IPNet
	VRFID             uint32
	Phase             string // Connected, Degraded, Disconnected
	LatencyMillis     float64
	LastHeartbeat     time.Time
	ActiveRoutes      int
}

// GlobalServiceRecord tracks an exported/imported service across the ClusterMesh
type GlobalServiceRecord struct {
	mu            sync.RWMutex
	ServiceName   string
	Namespace     string
	GlobalName    string // e.g. "orders.default.global.straitgateway"
	GlobalVIP     net.IP
	Ports         []v1alpha1.GlobalServicePort
	TrafficPolicy string // ClusterMeshPreferred, ClusterLocal, RoundRobin
	// ClusterBackends maps ClusterID -> list of endpoint IP:port strings
	ClusterBackends map[string][]string
	rrCounter       uint64
}
