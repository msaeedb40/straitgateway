# StraitGateway: Capabilities

StraitGateway provides an end-to-end networking and security substrate for cloud-native infrastructures. This document provides a comprehensive breakdown of its core capabilities and operational features.

---

## Capabilities Matrix

| Capability Area | Feature | Description | Default Status |
| :--- | :--- | :--- | :--- |
| **CNI & IPAM** | NetKit Datapath | Line-rate pod interconnect via Linux 6.7+ NetKit devices | Enabled |
| | Dynamic PerNode IPAM | Automatic per-node CIDR allocation via API server discovery | Enabled |
| | Dual-Stack | Simultaneous IPv4 and IPv6 routing and addressing | Ready |
| | Cluster & Node Sync | `ClusterNetworkConfig` and `NodeNetworkConfig` CRDs for state auditing | Enabled |
| **Service Load Balancer** | Kube-Proxy Replacement | 100% eBPF-driven service routing; zero iptables rules | Enabled |
| | Maglev Consistent Hashing | Backend selection minimizing connection churn (table size: 128) | Enabled |
| | Direct Server Return (DSR) | Backends reply directly to clients, bypassing gateway egress bottleneck | Configurable |
| | XDP NodePort Acceleration | Line-rate NodePort handling (`30000-32767`) at network card driver | Enabled |
| **Security & Policy** | Identity-Based Engine | Numeric 32-bit security identities derived from label hashing | Enabled |
| | Multi-Dimensional Selectors | Fine-grained rules spanning Namespaces, Pods, Clusters, Segments, Routes | Enabled |
| | BPF LSM Hooks | Kernel-enforced socket `connect()` and `bind()` restrictions | Enabled |
| | Default Deny Posture | Zero-trust baseline policy enforcement | Configurable |
| **Multi-Cluster Transit** | Mesh & Hub-Spoke Topologies | Interconnect clusters across VPCs, regions, and clouds | Configurable |
| | 32-bit Segment Isolation | Virtual network segmentation (Segment 0 default backbone) | Enabled |
| | WireGuard / IPsec Encryption | ChaCha20-Poly1305 and AES-GCM encrypted transit tunnels | WireGuard |
| **Gateway API v1.6.1** | GatewayClass `skgateway` | Native integration with Kubernetes Gateway API v1.6.1 | Enabled |
| | Multi-Protocol Routing | HTTPRoute, GRPCRoute, TLSRoute, TCPRoute, UDPRoute | Enabled |
| | Canary & Weighting | Fine-grained percentage-based traffic splitting | Enabled |
| **BGP & Dynamic Routing** | BGP Peering | Peering with Top-of-Rack (ToR) switches using `BGPPeer` CRD | Optional |
| | BFD Fast Detection | Sub-second link failure detection using `BFDSession` CRD | Optional |
| | VIP Advertisement | Announce Service `LoadBalancer` IPs directly via BGP | Optional |
| **Observability** | Prometheus Metrics | High-resolution Prometheus endpoint (`straitgateway_*` on port 9090) | Enabled |
| | Flow Logging | eBPF ring buffer event capture for real-time flow audits | Enabled |
| | OpenTelemetry Tracing | OTLP trace export for distributed network latency analysis | Optional |
| **Operational Tooling** | `sg-cli` | 17-subcommand CLI for diagnostics, BPF maps, and node status | Binary |
| | Angular Console UI | 15-module management dashboard with packet capture and topology map | Helm (Optional) |

---

## 1. CNI, IPAM, & State Synchronization

### NetKit Datapath (Linux Kernel >= 6.7.0)
Unlike legacy CNIs that attach containers via virtual ethernet (`veth`) pairs, StraitGateway uses **NetKit** devices. NetKit reduces packet processing overhead, eliminates sk_buff clone operations, and provides direct memory hooks into eBPF programs, achieving significantly lower latency under high packet rates.

### PerNode IPAM Engine
- Automatically discovers the cluster-wide pod CIDR from the Kubernetes API server (never hardcoded).
- Partitions the CIDR into deterministic per-node subnets.
- Eliminates central IPAM lock contention during rapid scaling events.
- Supports both IPv4 (`/24` default per node) and IPv6 (`/64` default per node).

### Cluster & Node Network Configuration CRDs
- **`ClusterNetworkConfig` (`cnc`)**: Holds cluster-wide network state, discovered PodCIDR/ServiceCIDR, and counts of ready nodes.
- **`NodeNetworkConfig` (`nnc`)**: Tracks per-node desired and observed dataplane health:
  - `status.dataplane.cniReady`: CNI socket and link status.
  - `status.dataplane.serviceReady`: eBPF service proxy health.
  - `status.dataplane.policyReady`: Identity and policy map synchronization.
  - `status.dataplane.kubeProxyReplacement`: Confirmation that kube-proxy is disabled and replaced.
  - `status.bpfMapRevision`: Monotonically increasing revision of active BPF maps on the node.

---

## 2. Kube-Proxy Replacement & Service Load Balancing

StraitGateway completely eliminates `kube-proxy` by compiling Kubernetes `Service` and `Endpoints` definitions directly into eBPF maps:

### Load Balancing Algorithms
StraitGateway supports five load balancing algorithms selectable per service or globally:
1. **Maglev (Default)**: Consistent hashing algorithm developed by Google. Guarantees minimal disruption to existing connections when backends are added or removed. Uses a prime lookup table (default: `128`).
2. **Round-Robin**: Distributes connections evenly across all healthy backends.
3. **Least Connections**: Dynamically routes new connections to the backend with the fewest active flows.
4. **IP Hash**: Deterministic hashing based on source IP address for basic session stickiness.
5. **Random**: Fast, lightweight pseudo-random backend distribution.

### Direct Server Return (DSR)
When DSR is enabled, inbound client requests pass through the gateway/ingress node, but response packets from the backend pod are returned directly to the client. This doubles ingress throughput and eliminates the egress bottleneck on gateway nodes.

### NodePort Acceleration via XDP
Incoming NodePort traffic (ports `30000-32767`) is intercepted at the **XDP (eXpress Data Path)** driver hook before the Linux TCP/IP stack allocates socket buffers. If the target backend resides on another node, the packet is redirected immediately across the fabric at wire speed.

---

## 3. Dynamic BGP & BFD Routing

For bare-metal and hybrid cloud deployments, StraitGateway includes a native BGP/BFD controller:

### BGP Peering (`BGPPeer`)
- Establishes eBGP or iBGP sessions with upstream datacenter switches (ToR / Spine).
- Automatically advertises external IP addresses assigned to Kubernetes `LoadBalancer` services, allowing ECMP (Equal-Cost Multi-Path) hardware routing directly into the cluster.

### BFD Protocol (`BFDSession`)
- **Bidirectional Forwarding Detection** provides sub-second link failure detection.
- Configurable parameters:
  - `peerAddress`: Remote switch or router IP.
  - `detectMultiplier`: Default `3`.
  - `receiveInterval`: Default `300ms` (minimum `10ms`).
  - `transmitInterval`: Default `300ms` (minimum `10ms`).
- Triggers immediate BGP route withdrawal before application timeouts occur.

---

## 4. Operational CLI: `sg-cli`

StraitGateway includes `sg-cli`, an administrative binary featuring **17 modular subcommands**:

```bash
# General status of cluster, nodes, and dataplane
sg-cli status

# Inspect Gateway API resources and routes
sg-cli gateway list
sg-cli gateway routes -g <gateway-name>

# Query node network states and BPF map revisions
sg-cli node list
sg-cli node get <node-name>

# Inspect active BGP neighbors and BFD sessions
sg-cli bgp peers
sg-cli bgp routes

# Query security identities and policy enforcement
sg-cli policy list
sg-cli policy identities

# Multi-cluster transit inspection
sg-cli transit segments
sg-cli transit peers

# WireGuard tunnel status and peer transfer bytes
sg-cli wireguard status

# Open or port-forward the UI dashboard
sg-cli ui
```

---

## 5. UI Dashboard: `straitgateway-ui`

The Angular dashboard provides 15 unified views:
- **Topology Map**: Dynamic D3 network visualization of nodes, pods, services, tunnels, and cross-cluster attachments.
- **Packet Capture**: Live in-browser packet capture with protocol, IP, and port filters directly from the eBPF ring buffer.
- **eBPF Inspector**: Live view of kernel BPF maps (`ct_map`, `lb_map`, `identity_map`, `policy_map`), attachment points, and LSM hooks.
- **Flow Analytics**: Real-time traffic stream with action verdicts (`Allow`, `Deny`, `Drop`) and drop-reason diagnostics.
- **Observability Hub**: Integrated Prometheus charts, Jaeger trace waterfalls, and structured logs with live tailing.
- **Role-Based Access**: Integrated OIDC / OAuth2 token authentication with Angular route guards.

---

## 6. Observability & Telemetry

StraitGateway exposes non-intrusive visibility directly from the kernel:

- **Prometheus Metrics**: Available on port `9090` at `/metrics`:
  - `straitgateway_bpf_packet_count_total`
  - `straitgateway_bpf_drop_count_total`
  - `straitgateway_service_active_backends`
  - `straitgateway_transit_peer_latency_seconds`
  - `straitgateway_policy_verdict_total`
- **eBPF Ring Buffer Flow Logs**: Emits real-time packet flow records (source, destination, protocol, identities, policy verdict, dropped reasons) to stdout, files, or gRPC collectors without impacting dataplane performance.
- **OpenTelemetry Tracing**: Integrates with OpenTelemetry collectors to trace packet latency across multi-cluster transit hops.
