package observability

import (
	"encoding/json"
	"net/http"
)

// ClusterOverview represents cluster-wide dataplane status
type ClusterOverview struct {
	Name                 string `json:"name"`
	Version              string `json:"version"`
	KubernetesVersion    string `json:"kubernetesVersion"`
	KernelVersion        string `json:"kernelVersion"`
	EbpfVersion          string `json:"ebpfVersion"`
	Status               string `json:"status"`
	NodesCount           int    `json:"nodesCount"`
	PodsCount            int    `json:"podsCount"`
	ServicesCount        int    `json:"servicesCount"`
	GatewaysCount        int    `json:"gatewaysCount"`
	ActiveWireguardPeers int    `json:"activeWireguardPeers"`
	BGPSessionsUp        int    `json:"bgpSessionsUp"`
	UptimeSeconds        int    `json:"uptimeSeconds"`
}

// NodeInfo represents a Kubernetes node running StraitGateway
type NodeInfo struct {
	Name               string  `json:"name"`
	IP                 string  `json:"ip"`
	PodCIDR            string  `json:"podCIDR"`
	Role               string  `json:"role"`
	Status             string  `json:"status"`
	KernelVersion      string  `json:"kernelVersion"`
	NetkitSupported    bool    `json:"netkitSupported"`
	TCXAttached        bool    `json:"tcxAttached"`
	EbpfLoadedCount    int     `json:"ebpfLoadedCount"`
	WireguardPubKey    string  `json:"wireguardPubKey"`
	WireguardEndpoint  string  `json:"wireguardEndpoint"`
	CPUUsagePercent    float64 `json:"cpuUsagePercent"`
	MemoryUsagePercent float64 `json:"memoryUsagePercent"`
}

// PodInfo represents a workload pod attached via netkit
type PodInfo struct {
	Namespace        string `json:"namespace"`
	Name             string `json:"name"`
	IP               string `json:"ip"`
	NodeName         string `json:"nodeName"`
	NetkitInterface  string `json:"netkitInterface"`
	Status           string `json:"status"`
	SecurityIdentity uint32 `json:"securityIdentity"`
	PolicyEnforced   bool   `json:"policyEnforced"`
	CreatedAt        string `json:"createdAt"`
}

// ServiceInfo represents a ClusterIP / LoadBalancer service
type ServiceInfo struct {
	Namespace      string        `json:"namespace"`
	Name           string        `json:"name"`
	Type           string        `json:"type"`
	ClusterIP      string        `json:"clusterIP"`
	Ports          []ServicePort `json:"ports"`
	EndpointsCount int           `json:"endpointsCount"`
	Algorithm      string        `json:"algorithm"`
}

// ServicePort represents a port mapping
type ServicePort struct {
	Name       string      `json:"name"`
	Protocol   string      `json:"protocol"`
	Port       int         `json:"port"`
	TargetPort interface{} `json:"targetPort"`
}

// GatewayResource represents a Gateway API Gateway
type GatewayResource struct {
	Name             string            `json:"name"`
	Namespace        string            `json:"namespace"`
	GatewayClassName string            `json:"gatewayClassName"`
	Listeners        []GatewayListener `json:"listeners"`
	Addresses        []string          `json:"addresses"`
	Programmed       bool              `json:"programmed"`
	Ready            bool              `json:"ready"`
}

// GatewayListener represents a listener on a Gateway
type GatewayListener struct {
	Name        string `json:"name"`
	Protocol    string `json:"protocol"`
	Port        int    `json:"port"`
	RoutesCount int    `json:"routesCount"`
}

// TransitGatewayResource represents a multi-tenant TransitGateway CR
type TransitGatewayResource struct {
	Name         string              `json:"name"`
	Namespace    string              `json:"namespace"`
	ASN          uint32              `json:"asn"`
	RouterID     string              `json:"routerId"`
	VRFs         []TransitGatewayVRF `json:"vrfs"`
	PeersCount   int                 `json:"peersCount"`
	TunnelsCount int                 `json:"tunnelsCount"`
	Status       string              `json:"status"`
}

// TransitGatewayVRF represents a VRF instance
type TransitGatewayVRF struct {
	Name         string   `json:"name"`
	ID           uint32   `json:"id"`
	RouteTableID uint32   `json:"routeTableId"`
	CIDRs        []string `json:"cidrs"`
}

// HttpRouteResource represents a Gateway API HTTPRoute
type HttpRouteResource struct {
	Name       string   `json:"name"`
	Namespace  string   `json:"namespace"`
	ParentRefs []string `json:"parentRefs"`
	Hostnames  []string `json:"hostnames"`
	RulesCount int      `json:"rulesCount"`
}

// RouteEntry represents an LPM route
type RouteEntry struct {
	DestinationCIDR string `json:"destinationCIDR"`
	PrefixLen       int    `json:"prefixLen"`
	NextHop         string `json:"nextHop"`
	OutputInterface string `json:"outputInterface"`
	RouteType       string `json:"routeType"`
	VRFID           uint32 `json:"vrfId"`
	Metric          uint32 `json:"metric"`
	Status          string `json:"status"`
}

// BgpPeerSession represents a BGP peer
type BgpPeerSession struct {
	PeerAddress        string `json:"peerAddress"`
	ASN                uint32 `json:"asn"`
	State              string `json:"state"`
	UptimeSeconds      int64  `json:"uptimeSeconds"`
	PrefixesReceived   int    `json:"prefixesReceived"`
	PrefixesAdvertised int    `json:"prefixesAdvertised"`
	HoldTime           int    `json:"holdTime"`
	KeepaliveTime      int    `json:"keepaliveTime"`
}

// BfdSession represents a BFD fast-failover session
type BfdSession struct {
	RemoteAddress         string `json:"remoteAddress"`
	State                 string `json:"state"`
	DetectMult            uint8  `json:"detectMult"`
	TxIntervalMs          uint32 `json:"txIntervalMs"`
	RxIntervalMs          uint32 `json:"rxIntervalMs"`
	LastTransition        string `json:"lastTransition"`
	FastFailoverTriggered bool   `json:"fastFailoverTriggered"`
}

// WireguardPeer represents an active tunnel peer
type WireguardPeer struct {
	PublicKey           string   `json:"publicKey"`
	Endpoint            string   `json:"endpoint"`
	AllowedIPs          []string `json:"allowedIPs"`
	LatestHandshake     string   `json:"latestHandshake"`
	TransferRxBytes     uint64   `json:"transferRxBytes"`
	TransferTxBytes     uint64   `json:"transferTxBytes"`
	PersistentKeepalive int      `json:"persistentKeepalive"`
}

// SecurityPolicyRule represents an eBPF policy
type SecurityPolicyRule struct {
	ID                  string   `json:"id"`
	Name                string   `json:"name"`
	Namespace           string   `json:"namespace"`
	Action              string   `json:"action"`
	Direction           string   `json:"direction"`
	SourceSelector      string   `json:"sourceSelector"`
	DestinationSelector string   `json:"destinationSelector"`
	Ports               []string `json:"ports"`
	Protocol            string   `json:"protocol"`
	HitsCount           uint64   `json:"hitsCount"`
}

func (s *Server) handleClusterOverview(w http.ResponseWriter, r *http.Request) {
	overview := ClusterOverview{
		Name:                 "StraitGateway Transit Mesh",
		Version:              "v1.0.0",
		KubernetesVersion:    "v1.32.2",
		KernelVersion:        "6.12.0-sg",
		EbpfVersion:          "CO-RE v1.0",
		Status:               "Healthy",
		NodesCount:           8,
		PodsCount:            142,
		ServicesCount:        24,
		GatewaysCount:        4,
		ActiveWireguardPeers: 3,
		BGPSessionsUp:        2,
		UptimeSeconds:        86400 * 3,
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(overview)
}

func (s *Server) handleClusterHealth(w http.ResponseWriter, r *http.Request) {
	health := map[string]interface{}{
		"status": "Healthy",
		"uptime": 86400 * 3,
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(health)
}

func (s *Server) handleNodes(w http.ResponseWriter, r *http.Request) {
	nodes := []NodeInfo{
		{
			Name: "sg-node-01", IP: "192.168.1.10", PodCIDR: "10.244.1.0/24", Role: "control-plane",
			Status: "Ready", KernelVersion: "6.12.0-sg", NetkitSupported: true, TCXAttached: true,
			EbpfLoadedCount: 13, WireguardPubKey: "a9F1k...PUB1", WireguardEndpoint: "192.168.1.10:51820",
			CPUUsagePercent: 12.4, MemoryUsagePercent: 28.1,
		},
		{
			Name: "sg-node-02", IP: "192.168.1.11", PodCIDR: "10.244.2.0/24", Role: "worker",
			Status: "Ready", KernelVersion: "6.12.0-sg", NetkitSupported: true, TCXAttached: true,
			EbpfLoadedCount: 13, WireguardPubKey: "b7K2p...PUB2", WireguardEndpoint: "192.168.1.11:51820",
			CPUUsagePercent: 24.8, MemoryUsagePercent: 44.5,
		},
		{
			Name: "sg-node-03", IP: "192.168.1.12", PodCIDR: "10.244.3.0/24", Role: "worker",
			Status: "Ready", KernelVersion: "6.12.0-sg", NetkitSupported: true, TCXAttached: true,
			EbpfLoadedCount: 13, WireguardPubKey: "c3L8q...PUB3", WireguardEndpoint: "192.168.1.12:51820",
			CPUUsagePercent: 31.2, MemoryUsagePercent: 51.0,
		},
		{
			Name: "sg-transit-01", IP: "192.168.1.50", PodCIDR: "10.244.0.0/24", Role: "transit-node",
			Status: "Ready", KernelVersion: "6.12.0-sg", NetkitSupported: true, TCXAttached: true,
			EbpfLoadedCount: 13, WireguardPubKey: "t9R4x...PUB4", WireguardEndpoint: "192.168.1.50:51820",
			CPUUsagePercent: 8.5, MemoryUsagePercent: 22.3,
		},
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(nodes)
}

func (s *Server) handlePods(w http.ResponseWriter, r *http.Request) {
	pods := []PodInfo{
		{Namespace: "default", Name: "frontend-78f99d6594-abc12", IP: "10.244.1.5", NodeName: "sg-node-01", NetkitInterface: "nk-pod-01", Status: "Running", SecurityIdentity: 1001, PolicyEnforced: true, CreatedAt: "2026-10-07T12:00:00Z"},
		{Namespace: "default", Name: "backend-55c84d764d-def34", IP: "10.244.2.8", NodeName: "sg-node-02", NetkitInterface: "nk-pod-02", Status: "Running", SecurityIdentity: 1002, PolicyEnforced: true, CreatedAt: "2026-10-07T12:05:00Z"},
		{Namespace: "default", Name: "database-64b59cb78-ghi56", IP: "10.244.3.12", NodeName: "sg-node-03", NetkitInterface: "nk-pod-03", Status: "Running", SecurityIdentity: 1003, PolicyEnforced: true, CreatedAt: "2026-10-07T12:10:00Z"},
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(pods)
}

func (s *Server) handleServices(w http.ResponseWriter, r *http.Request) {
	services := []ServiceInfo{
		{
			Namespace: "default", Name: "kubernetes", Type: "ClusterIP", ClusterIP: "10.96.0.1",
			Ports:          []ServicePort{{Name: "https", Protocol: "TCP", Port: 443, TargetPort: 6443}},
			EndpointsCount: 1, Algorithm: "maglev",
		},
		{
			Namespace: "kube-system", Name: "kube-dns", Type: "ClusterIP", ClusterIP: "10.96.0.10",
			Ports:          []ServicePort{{Name: "dns", Protocol: "UDP", Port: 53, TargetPort: 53}},
			EndpointsCount: 2, Algorithm: "maglev",
		},
		{
			Namespace: "default", Name: "frontend-svc", Type: "ClusterIP", ClusterIP: "10.96.10.50",
			Ports:          []ServicePort{{Name: "http", Protocol: "TCP", Port: 80, TargetPort: 8080}},
			EndpointsCount: 3, Algorithm: "maglev",
		},
		{
			Namespace: "default", Name: "backend-svc", Type: "ClusterIP", ClusterIP: "10.96.20.100",
			Ports:          []ServicePort{{Name: "grpc", Protocol: "TCP", Port: 9000, TargetPort: 9000}},
			EndpointsCount: 4, Algorithm: "maglev",
		},
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(services)
}

func (s *Server) handleGateways(w http.ResponseWriter, r *http.Request) {
	gws := []GatewayResource{
		{
			Name: "prod-external-gateway", Namespace: "straitgateway-system", GatewayClassName: "straitgateway",
			Listeners: []GatewayListener{
				{Name: "https", Protocol: "HTTPS", Port: 443, RoutesCount: 5},
				{Name: "http", Protocol: "HTTP", Port: 80, RoutesCount: 2},
			},
			Addresses: []string{"198.51.100.1"}, Programmed: true, Ready: true,
		},
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(gws)
}

func (s *Server) handleTransitGateways(w http.ResponseWriter, r *http.Request) {
	tgws := []TransitGatewayResource{
		{
			Name: "global-transit-mesh", Namespace: "straitgateway-system", ASN: 65001, RouterID: "192.168.1.50",
			VRFs: []TransitGatewayVRF{
				{Name: "prod-vpc", ID: 10, RouteTableID: 100, CIDRs: []string{"10.244.0.0/16", "172.16.0.0/16"}},
				{Name: "corp-transit", ID: 20, RouteTableID: 200, CIDRs: []string{"192.168.0.0/16"}},
			},
			PeersCount: 2, TunnelsCount: 3, Status: "Active",
		},
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(tgws)
}

func (s *Server) handleHttpRoutes(w http.ResponseWriter, r *http.Request) {
	routes := []HttpRouteResource{
		{
			Name: "frontend-route", Namespace: "default", ParentRefs: []string{"prod-external-gateway"},
			Hostnames: []string{"app.straitgateway.io"}, RulesCount: 3,
		},
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(routes)
}

func (s *Server) handleRoutes(w http.ResponseWriter, r *http.Request) {
	routes := []RouteEntry{
		{DestinationCIDR: "10.244.0.0/16", PrefixLen: 16, NextHop: "local", OutputInterface: "netkit", RouteType: "Local", VRFID: 0, Metric: 10, Status: "Active"},
		{DestinationCIDR: "10.96.0.0/12", PrefixLen: 12, NextHop: "service-vip", OutputInterface: "bpf-lb", RouteType: "Cluster", VRFID: 0, Metric: 10, Status: "Active"},
		{DestinationCIDR: "172.16.0.0/16", PrefixLen: 16, NextHop: "192.168.10.1", OutputInterface: "eth0", RouteType: "BGP", VRFID: 10, Metric: 20, Status: "Active"},
		{DestinationCIDR: "192.168.100.0/24", PrefixLen: 24, NextHop: "192.168.10.2", OutputInterface: "wg0", RouteType: "WireGuard", VRFID: 20, Metric: 50, Status: "Active"},
		{DestinationCIDR: "0.0.0.0/0", PrefixLen: 0, NextHop: "192.168.1.1", OutputInterface: "eth0", RouteType: "Static", VRFID: 0, Metric: 100, Status: "Active"},
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(routes)
}

func (s *Server) handleBgpPeers(w http.ResponseWriter, r *http.Request) {
	peers := []BgpPeerSession{
		{
			PeerAddress: "192.168.10.1", ASN: 65002, State: "Established", UptimeSeconds: 86400,
			PrefixesReceived: 14, PrefixesAdvertised: 4, HoldTime: 90, KeepaliveTime: 30,
		},
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(peers)
}

func (s *Server) handleBfdSessions(w http.ResponseWriter, r *http.Request) {
	sessions := []BfdSession{
		{
			RemoteAddress: "192.168.10.1", State: "Up", DetectMult: 3, TxIntervalMs: 50, RxIntervalMs: 50,
			LastTransition: "Up 24h", FastFailoverTriggered: false,
		},
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(sessions)
}

func (s *Server) handleWireguardPeers(w http.ResponseWriter, r *http.Request) {
	peers := []WireguardPeer{
		{
			PublicKey: "kX8p7...PUB1", Endpoint: "192.168.10.2:51820", AllowedIPs: []string{"192.168.100.0/24"},
			LatestHandshake: "12s ago", TransferRxBytes: 104857600, TransferTxBytes: 85492100, PersistentKeepalive: 25,
		},
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(peers)
}

func (s *Server) handleSecurityPolicies(w http.ResponseWriter, r *http.Request) {
	policies := []SecurityPolicyRule{
		{
			ID: "pol-01", Name: "allow-frontend-to-backend", Namespace: "default", Action: "Allow",
			Direction: "Ingress", SourceSelector: "app=frontend", DestinationSelector: "app=backend",
			Ports: []string{"8080/TCP"}, Protocol: "TCP", HitsCount: 48219,
		},
		{
			ID: "pol-02", Name: "deny-external-to-database", Namespace: "default", Action: "Deny",
			Direction: "Ingress", SourceSelector: "scope=world", DestinationSelector: "app=database",
			Ports: []string{"*"}, Protocol: "ANY", HitsCount: 142,
		},
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(policies)
}
