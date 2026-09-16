# StraitGateway: Overview

**StraitGateway** is an eBPF-native Kubernetes Container Network Interface (CNI), Service Load Balancer, Gateway API implementation, and Multi-Cluster Transit Gateway built for modern Linux kernels (`>= 6.7.0`).

Designed from the ground up to bypass the legacy baggage of iptables, IPVS, and veth pairs, StraitGateway delivers line-rate network performance, sub-millisecond latencies, and granular identity-based security policies for mission-critical cloud-native workloads.

---

## The Problems StraitGateway Solves

### 1. The Legacy iptables & IPVS Bottleneck
Standard Kubernetes networking relies on `kube-proxy`, which programs thousands of sequential `iptables` rules or IPVS virtual servers. As clusters scale past thousands of services and endpoints:
- Rule evaluation causes linear packet latency degradation.
- Concurrent updates trigger kernel lock contention and CPU thrashing.
- Connection tracking (`conntrack`) table exhaustion leads to silent packet drops.

**StraitGateway replaces kube-proxy entirely** with eBPF programs attached to Linux TCX and XDP hooks, using constant-time `O(1)` hash table lookups and consistent hashing algorithms (Maglev).

### 2. High-Overhead veth Pair Encapsulation
Traditional CNIs connect container network namespaces using virtual ethernet (`veth`) pairs. Every packet traverses the Linux TCP/IP stack twice, causing context switches, socket allocations, and sk_buff copies.

**StraitGateway leverages NetKit and eBPF redirect engines** (available in Linux 6.7+), streaming packets between namespaces with minimal kernel overhead and direct memory access.

### 3. Fragmented Networking Tooling
In typical enterprise setups, operators stitch together multiple disparate tools:
- A basic CNI plugin for pod IP assignment.
- An ingress controller for external HTTP traffic.
- A multi-cluster overlay/mesh tool for cross-cluster pod communication.
- A separate security agent for layer 4/7 network policies.
- Host routing daemons for BGP/BFD top-of-rack peering.

**StraitGateway unifies these capabilities into a single integrated platform** driven by a single control plane (`sg-controller`), a node agent (`straitgatewayd`), an administrative CLI (`sg-cli`), and an interactive Angular topology console (`straitgateway-ui`).

---

## Core Pillars & Innovations

### eBPF-Native Dataplane
- **Kernel-Level Execution**: Programs run directly in the Linux kernel via eBPF (Extended Berkeley Packet Filter), inspecting and redirecting packets before they reach the main network stack.
- **TCX & XDP Acceleration**: Fast eBPF hooks on ingress/egress interfaces and high-speed XDP drivers for edge packet filtering and NodePort routing.
- **NetKit Driver Integration**: High-throughput container interconnect replacing traditional veth pairs (requires Linux 6.7+).

### Strict Intermediate Representation (IR) Pipeline
- **Decoupled Architecture**: Kubernetes controllers never manipulate raw BPF maps or netlink sockets directly. Instead, 10 specialized controllers compile Kubernetes specifications into strongly-typed Intermediate Representation (`DataplaneState`).
- **Atomic & Generation-Tracked**: The Dataplane Compiler computes revision deltas and atomically commits changes to BPF maps, preventing partial states or race conditions during rapid pod churn.
- **Node-to-Cluster State Reconciliation**: Node state and dataplane readiness are synchronized via `ClusterNetworkConfig` and `NodeNetworkConfig` custom resources.

### Gateway API v1.6.1 First-Class Support
- Complete implementation of the Kubernetes Gateway API standard.
- Controller identifier: `straitgateway.io/skgateway`.
- Managed GatewayClass: `skgateway`.
- Comprehensive protocol support: `HTTPRoute`, `GRPCRoute`, `TLSRoute`, `TCPRoute`, and `UDPRoute`.

### Multi-Cluster Transit Gateway
- Enterprise-grade transit networking enabling Kubernetes clusters to communicate seamlessly across clouds, on-premises datacenters, and VPCs.
- Supports flexible topologies: `Mesh`, `HubAndSpoke`, `PeerToPeer`, and `GatewayToGateway`.
- 32-bit segment isolation (`TransitSegment`) with automated WireGuard encryption tunnels.

---

## Capabilities At a Glance

| Feature Area | Implementation |
| :--- | :--- |
| **CNI Networking** | Native routing, dual-stack IPv4/IPv6, NetKit interfaces, PerNode IPAM |
| **Service Proxy** | Full kube-proxy replacement, Maglev consistent hash (prime size default: 128), DSR, XDP NodePort |
| **Security** | Identity-based `StraitNetworkPolicy`, BPF LSM socket connect/bind controls, default-deny posture |
| **Ingress / Gateway** | Gateway API v1.6.1 (`GatewayClass`, `Gateway`, `HTTPRoute`, `GRPCRoute`, `TLSRoute`, `TCPRoute`, `UDPRoute`) |
| **Multi-Cluster** | Transit Gateway, 32-bit segment isolation, WireGuard ChaCha20-Poly1305 / IPsec encryption |
| **Routing Protocols** | Dynamic BGP route advertisements (`BGPPeer`), sub-second BFD failure detection (`BFDSession`) |
| **Observability** | Prometheus metrics (`straitgateway_*`), OpenTelemetry tracing, ring-buffer flow logs |
| **Management** | Lightweight Go daemonset (`straitgatewayd`), 17-subcommand `sg-cli` tool, and 15-module Angular UI |

---

## Next Steps

- Explore the detailed internal design in [Architecture](architecture.md).
- Learn about the feature set in [Capabilities](capabilities.md).
- Read about the security and isolation model in [Security](security.md).
- Follow the [Installation Guide](guide.md) to deploy StraitGateway on your cluster.
