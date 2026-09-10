# StraitGateway: Capabilities

StraitGateway provides an end-to-end networking and security substrate for cloud-native infrastructures. This document provides a comprehensive breakdown of its core capabilities and operational features.

---

## Capabilities Matrix

| Capability Area | Feature | Description | Default Status |
| :--- | :--- | :--- | :--- |
| **CNI & IPAM** | NetKit Datapath | Line-rate pod-to-host and pod-to-pod interconnect | Enabled |
| | PerNode IPAM | Automatic per-node CIDR allocation from cluster pod CIDR | Enabled |
| | Dual-Stack | Simultaneous IPv4 and IPv6 routing and addressing | Ready |
| **Service Load Balancer** | Kube-Proxy Replacement | 100% eBPF-driven service routing; zero iptables rules | Enabled |
| | Maglev Consistent Hashing | Stable backend selection minimizing reshuffling during pod scaling | Enabled |
| | Direct Server Return (DSR) | Backends reply directly to clients, bypassing gateway egress bottleneck | Configurable |
| | XDP NodePort Acceleration | Line-rate NodePort traffic handling at the network card driver | Enabled |
| **Security & Policy** | Identity-Based Engine | Numeric security identities derived from label hashing | Enabled |
| | Multi-Dimensional Selectors | Fine-grained rules spanning Namespaces, Pods, Clusters, and Segments | Enabled |
| | BPF LSM Hooks | Kernel-enforced socket creation and connection restrictions | Enabled |
| | Default Deny Posture | Zero-trust baseline policy enforcement | Configurable |
| **Multi-Cluster Transit** | Mesh & Hub-Spoke Topologies | Interconnect clusters across VPCs, regions, and clouds | Enabled |
| | 32-bit Segment Isolation | Virtual network segmentation (Segment 0 backbone) | Enabled |
| | WireGuard Encryption | ChaCha20-Poly1305 encrypted transit tunnels | Enabled |
| **Gateway API v1.6.1** | GatewayClass `skgateway` | Native integration with Kubernetes Gateway API v1.6.1 | Enabled |
| | Multi-Protocol Routing | HTTPRoute, GRPCRoute, TLSRoute, TCPRoute, UDPRoute | Enabled |
| | Canary & Weighting | Fine-grained percentage-based traffic splitting | Enabled |
| **BGP & Dynamic Routing** | BGP Peering | Peering with Top-of-Rack (ToR) switches and cloud routers | Optional |
| | BFD Protocol | Bidirectional Forwarding Detection for sub-second failover | Optional |
| | VIP Advertisement | Announce Service `LoadBalancer` IPs directly via BGP | Optional |
| **Observability** | Prometheus Metrics | High-resolution Prometheus endpoint (`straitgateway_*`) | Port 9090 |
| | Flow Logging | eBPF ring buffer event capture for real-time flow audits | Enabled |
| | OpenTelemetry Tracing | OTLP trace export for distributed network latency analysis | Optional |

---

## 1. CNI & IPAM

### NetKit Datapath
Unlike legacy CNIs that attach containers via virtual ethernet (`veth`) pairs, StraitGateway uses **NetKit** devices. NetKit reduces packet processing overhead, eliminates sk_buff clone operations, and provides direct memory hooks into eBPF programs, achieving significantly lower latency under high packet rates.

### PerNode IPAM Engine
- Automatically partitions the cluster-wide pod CIDR into deterministic per-node subnets.
- Eliminates central IPAM lock contention during high-churn deployments.
- Supports both IPv4 (`/24` default per node) and IPv6 (`/64` default per node).

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

- **BGP Peering**: Establishes eBGP or iBGP sessions with upstream datacenter switches (ToR / Spine) using CRDs (`BGPPeer`).
- **LoadBalancer VIP Advertisement**: Automatically advertises external IP addresses assigned to Kubernetes `LoadBalancer` services, allowing ECMP (Equal-Cost Multi-Path) hardware routing directly into the cluster.
- **BFD (Bidirectional Forwarding Detection)**: Detects link and path failures within tens of milliseconds, triggering immediate BGP route withdrawal before application timeouts occur.

---

## 4. Observability & Telemetry

StraitGateway exposes deep, non-intrusive visibility directly from the kernel:

- **Prometheus Metrics**: Available on port `9090` at `/metrics`. Metrics include:
  - `straitgateway_bpf_packet_count_total`
  - `straitgateway_bpf_drop_count_total`
  - `straitgateway_service_active_backends`
  - `straitgateway_transit_peer_latency_seconds`
  - `straitgateway_policy_verdict_total`
- **eBPF Ring Buffer Flow Logs**: Emits real-time packet flow records (source, destination, protocol, identities, policy verdict, dropped reasons) to stdout, files, or gRPC collectors without impacting dataplane performance.
- **OpenTelemetry Tracing**: Integrates with OpenTelemetry collectors to trace packet latency across multi-cluster transit hops.
- **Angular UI**: Interactive dashboard for visualizing cluster topology, service maps, active routes, and transit peer connectivity.
