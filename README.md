<p align="center">
  <img src="straitgateway.jpeg" alt="StraitGateway Logo" width="500" />
</p>

# StraitGateway

[![CI](https://github.com/straitgateway/straitgateway/actions/workflows/ci.yml/badge.svg)](https://github.com/straitgateway/straitgateway/actions/workflows/ci.yml)
[![License](https://img.shields.io/badge/License-Apache_2.0-blue.svg)](LICENSE)
[![Go Report Card](https://goreportcard.com/badge/github.com/straitgateway/straitgateway)](https://goreportcard.com/report/github.com/straitgateway/straitgateway)

StraitGateway is a Kubernetes-native transit gateway, CNI, sidecarless service mesh, and networking platform designed to provide high-performance connectivity between pods, nodes, services, Kubernetes clusters, and external networks.

It leverages **eBPF**, **netkit**, **WireGuard**, **BGP**, and **BFD** to provide programmable packet processing, secure connectivity, dynamic routing, service load balancing, and network observability.

---

## Core Architecture

```
                            ┌──────────────────────────────┐
                            │          Kubernetes          │
                            │   API Server / Gateway API   │
                            └──────────────┬───────────────┘
                                           │
                                           ▼
                  ┌─────────────────────────────────────────────┐
                  │         StraitGateway Control Plane         │
                  │  ┌────────────────┐   ┌──────────────────┐  │
                  │  │ SG Controller  │   │ Gateway / CNI    │  │
                  │  │                │   │ Controller       │  │
                  │  └───────┬────────┘   └────────┬─────────┘  │
                  └──────────┼─────────────────────┼────────────┘
                             │                     │
                             ▼                     ▼
      ┌──────────────────────────────────────────────────────────────┐
      │                       Kubernetes Nodes                       │
      │  ┌────────────────────────────────────────────────────────┐  │
      │  │                  StraitGateway Agent                   │  │
      │  │                       (straitd)                        │  │
      │  │  ┌──────────────┐     ┌────────────────────────────┐   │  │
      │  │  │ CNI Manager  │────▶│     eBPF Dataplane         │   │  │
      │  │  └──────────────┘     │  eBPF + CO-RE              │   │  │
      │  │                       │  netkit                    │   │  │
      │  │                       │  Service / LB / Routing    │   │  │
      │  │                       └─────────────┬──────────────┘   │  │
      │  │       ┌─────────────────────────────┴─────────────┐    │  │
      │  │  ┌────▼───────┐                            ┌──────▼──┐ │  │
      │  │  │ WireGuard  │                            │ BGP/BFD │ │  │
      │  │  │ Tunnel     │                            │ Routing │ │  │
      │  │  └────┬───────┘                            └──────┬──┘ │  │
      │  └───────┼───────────────────────────────────────────┼────┘  │
      └──────────┼───────────────────────────────────────────┼───────┘
                 │                                           │
                 ▼                                           ▼
           ┌───────────┐                              ┌─────────────┐
           │ Other SG  │                              │ External    │
           │ Nodes     │                              │ Routers /   │
           │           │                              │ Networks    │
           └───────────┘                              └─────────────┘
```

---

## Core Technology Stack

| Layer | Technology | Purpose |
|---|---|---|
| **Language** | **Go 1.27.2** | Controllers, agents, networking services, and CLI |
| **Kernel** | **Linux 6.8+ (6.12+ LTS Recommended)** | Host networking, TCX, netkit, and eBPF dataplane |
| **Dataplane** | **eBPF CO-RE** (`cilium/ebpf`) | Programmable, high-performance packet processing across kernels |
| **Pod Networking** | **netkit** | Next-generation direct pod-to-kernel packet handoff |
| **Security & Tunnels** | **WireGuard** | High-speed encrypted inter-node and inter-cluster overlay |
| **Dynamic Routing** | **BGP** | Dynamic route exchange with external networks & top-of-rack switches |
| **Failure Detection**| **BFD** | Sub-second detection of routing peer failures |
| **Orchestration** | **Kubernetes Gateway API v1.6.3** | Native declarative traffic management and transit gateway routing |
| **Observability** | **Prometheus + OTel + Loki** | eBPF ring buffer telemetry, metrics, traces, and flow logs |
| **Web UI** | **React 19 + TypeScript + D3.js** | Live cluster topology & observability without third-party query bloat |

---

## Implementation Roadmap

| Phase | Milestone | Focus Areas | Status |
|---|---|---|:---:|
| **Phase 1** | **Dataplane Foundation & BPF Maps** | 13 maps, loader, pins, TCX/XDP/cgroups, per-CPU counters | **Done** |
| **Phase 2** | **Native CNI & IPAM Subsystem** | `netkit` L3/L2 driver, IPAM bitmap allocator, `sg-cni` binary | **Done** |
| **Phase 3** | **Service Networking & Maglev** | Maglev lookup table, connect4/6 in-socket LB, TC NAT | **Done** |
| **Phase 4** | **Transit Gateway & Routing** | FIB LPM engine, VRF routing, WireGuard mesh, BGP, sub-50ms BFD | **Done** |
| **Phase 5** | **Security Policies & Mesh** | Workload identity engine, policy compiler, NodeProxy L7 engine | **Done** |
| **Phase 6** | **Dataplane Observability** | Prometheus exporter, flow log ring buffer, HTTP status server | **Done** |
| **Phase 7** | **CLI, Helm & Manifests** | `sgctl` CLI (11 subcommands), Helm chart v1.0.0, `all-in-one.yaml` | **Done** |
| **Phase 8** | **UI Plane & Integration** | React 19, Tailwind CSS 4, D3.js topology, REST API bridge | **Done** |
| **Phase 9** | **Multi-Cluster Transit (ClusterMesh)** | Global WireGuard mesh, cross-cluster Maglev, Split-DNS | **Done** |
| **Phase 10** | **Advanced L7 & Hardware Acceleration** | SmartNIC `XDP_OFFLOAD`, Wasm extension engine, Gateway API v1.6.3 | **Done** |
| **Phase 11** | **Enterprise Zero-Trust & In-Kernel mTLS** | SPIFFE/SPIRE attestation, kTLS + `sockops` zero-copy encryption | **Done** |
| **Phase 12** | **Chaos Engineering & eBPF Fuzzing** | Sonobuoy CNI conformance, BPF map fault injection, multi-kernel CI | **Done** |
| **Phase 13** | **AI/ML & RDMA/RoCEv2 Acceleration** | GPU/TPU RoCEv2 bypass, NCCL optimization, PFC/ECN telemetry | **Done** |
| **Phase 14** | **Autonomous Dataplane AIOps** | Micro-burst detection, self-healing routing, zero-code tracing | **Done** |
| **Phase 15** | **Documentation System (`docs/`)** | Markdown technical specs, guides, CRD & dataplane manuals (`*.md`) | **Done** |

---

## Key Features

* **Ultra-Fast Pod Networking (`netkit`)**: Employs Linux 6.8+ `netkit` virtual interfaces in direct L3 mode, achieving 78+ Gbps throughput and ~5.8 µs latency with zero-copy packet transfer.
* **In-Socket Service Load Balancing**: Transparently rewrites `connect4/6` syscalls using eBPF cgroups and Maglev consistent hashing, completely bypassing iptables, kube-proxy, and node-port NAT loops.
* **Dynamic Transit Gateway & Routing**: Multi-VRF LPM routing trie supporting RFC 4271 BGP-4 dynamic peering and sub-50ms RFC 5880 BFD hardware/software failure detection.
* **Global ClusterMesh Overlay**: WireGuard-backed point-to-multipoint kernel overlay mesh for zero-trust, encrypted inter-cluster transit and cross-cluster service routing.
* **Kubernetes Gateway API v1.6.3 Conformance**: Declarative ingress, egress, canary releases, header manipulation, and gRPC routing with SmartNIC `XDP_OFFLOAD` support.
* **AI/ML & RoCEv2 Acceleration**: eBPF-driven RoCEv2 packet processing, BTH/AETH parsing, DCQCN hardware rate limiting, and NCCL topology routing for high-performance GPU/TPU training clusters.
* **Autonomous Dataplane AIOps**: Dual EWMA micro-burst anomaly detection with automated route withdrawal and W3C Trace Context propagation.

---

## Quick Start & Installation

### Option 1: Standalone All-In-One Manifest

Deploy CRDs, controller, node agent (`straitd`), CNI plugin, and RBAC with a single command:

```bash
kubectl apply -f deploy/all-in-one.yaml
```

Verify deployment status and health probes:

```bash
# Check running pods in straitgateway-system namespace
kubectl get pods -n straitgateway-system -o wide

# Check CRDs
kubectl get crds | grep straitgateway

# Query health and readiness endpoints
curl -s http://localhost:9090/readyz
curl -s http://localhost:9090/healthz
```

### Option 2: Helm Chart

```bash
helm repo add straitgateway https://charts.straitgateway.io
helm repo update

helm install straitgateway charts/straitgateway \
  --namespace straitgateway-system \
  --create-namespace \
  --set bpf.mountPath="/sys/fs/bpf" \
  --set cni.mode="netkit" \
  --set wireguard.enabled=true \
  --set routing.bgp.localAsn=65001
```

---

## Documentation System

Explore in-depth technical documentation, API specifications, and operational playbooks in the [`docs/`](docs/) directory:

| Section | Topic | Documentation Link |
|---|---|---|
| **Architecture** | System Topology & Design | [Architecture Overview](docs/architecture/overview.md) |
| | Four-Plane Multi-Plane Deep Dive | [Multi-Plane Architecture](docs/architecture/multi-plane.md) |
| | Kernel Prerequisites & BTF | [Kernel Compatibility](docs/architecture/kernel-compatibility.md) |
| **Dataplane** | 18 Pinned BPF Maps | [eBPF Map Architecture](docs/dataplane/ebpf-maps.md) |
| | Packet Lifecycle (XDP → TCX) | [Packet Processing Lifecycle](docs/dataplane/packet-processing.md) |
| | AI/ML & RoCEv2 Dataplane | [RDMA & RoCEv2 Acceleration](docs/dataplane/rdma-roce.md) |
| **CNI & IPAM** | Linux 6.8+ Netkit Driver | [Netkit L3/L2 Driver](docs/cni/netkit-driver.md) |
| | Lockless Bitmap IPAM Allocator | [Dual-Stack IPAM](docs/cni/ipam.md) |
| **Dynamic Routing** | Multi-VRF Transit Gateway | [Transit Gateway](docs/routing/transit-gateway.md) |
| | BGP-4 & Sub-50ms BFD | [BGP & BFD Guide](docs/routing/bgp-bfd.md) |
| | WireGuard ClusterMesh Mesh | [WireGuard Overlay Mesh](docs/routing/wireguard.md) |
| **Service Mesh** | Sidecarless Shared NodeProxy | [Sidecarless Service Mesh](docs/service-mesh/sidecarless-mesh.md) |
| | Workload Identity & Policies | [Zero-Trust Security Policies](docs/service-mesh/security-policy.md) |
| | In-Kernel mTLS (kTLS + sockops)| [In-Kernel mTLS](docs/service-mesh/in-kernel-mtls.md) |
| | Gateway API v1.6.3 | [Gateway API Implementation](docs/gateway-api/gateway-api.md) |
| **Observability** | Prometheus & Flow Logs | [Telemetry & Monitoring](docs/observability/telemetry.md) |
| | Micro-Burst AIOps & Tracing | [Autonomous AIOps](docs/observability/aiops.md) |
| **Operations** | Installation & Deployment | [Installation Guide](docs/operations/installation.md) |
| | Diagnostic CLI Reference | [sgctl Reference](docs/operations/sgctl-reference.md) |
| | Troubleshooting & Playbooks | [Troubleshooting Guide](docs/operations/troubleshooting.md) |
| | Chaos & Conformance Testing | [Chaos & Conformance](docs/operations/chaos-conformance.md) |

---

## Diagnostic CLI (`sgctl`)

StraitGateway includes `sgctl`, a unified diagnostic CLI for operators:

```bash
# General health & BPF filesystem check
sgctl health

# List pinned BPF maps and dump contents
sgctl bpf maps list
sgctl bpf maps dump sg_routes

# Inspect dynamic routes, BGP sessions, and WireGuard peers
sgctl routes list
sgctl peers status

# Live flow logs and policy drops
sgctl flows --limit 20
sgctl policy list

# Run full cluster diagnostics
sgctl diagnose --all
```

---

## Building & Testing

```bash
# Clone the repository
git clone https://github.com/straitgateway/straitgateway.git
cd straitgateway

# Build all binaries
make build

# Output binaries will be in ./bin/:
#   ./bin/straitd         (Node Daemon)
#   ./bin/sg-controller   (Cluster Controller)
#   ./bin/sg-cni          (CNI Plugin)
#   ./bin/sg-agent        (Node Agent)
#   ./bin/sgctl           (Diagnostic CLI)

# Run tests and linter
make test
golangci-lint run ./...

# Run Gateway API and CNI conformance test suites
make conformance-cni
make conformance-gateway
```

---

## License

StraitGateway is licensed under the Apache 2.0 License. See [LICENSE](LICENSE) for details.