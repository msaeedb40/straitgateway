<div align="center">
  <img src="logo/straitgateway.png" alt="StraitGateway Logo" width="480">

  <br/>

  **eBPF-Native Kubernetes CNI, Service Load Balancer, Gateway API, & Multi-Cluster Transit Gateway**

  [![Go Version](https://img.shields.io/badge/Go-1.24%2B-00ADD8?style=flat&logo=go)](https://go.dev)
  [![Kubernetes](https://img.shields.io/badge/Kubernetes-%3E%3D%201.32.0-326CE5?style=flat&logo=kubernetes)](https://kubernetes.io)
  [![Gateway API](https://img.shields.io/badge/Gateway%20API-v1.6.1-blue?style=flat)](https://gateway-api.sigs.k8s.io)
  [![Datapath](https://img.shields.io/badge/Datapath-eBPF%20%7C%20NetKit%20%7C%20TCX-orange?style=flat)](https://ebpf.io)
  [![License](https://img.shields.io/badge/License-Apache%202.0-green.svg)](LICENSE)

</div>

---

## Overview

**StraitGateway** is a next-generation cloud-native networking platform built from the ground up for modern Linux kernels (`>= 5.15`). Bypassing legacy `iptables`, `IPVS`, and overhead-heavy `veth` pairs, StraitGateway executes line-rate packet routing, load balancing, security policies, and transit encapsulation directly inside the kernel via **eBPF**, **NetKit**, and **TCX**.

It replaces fragmented networking stacks with a unified, high-performance architecture driven by strict separation of concerns:

- 🚀 **Standalone Zero-Overhead CNI**: Direct NetKit-powered container interconnect (`strait-cni`) with dynamic Per-Node IPAM—no external CNI (Calico, Cilium, Flannel) required.
- ⚡ **Complete Kube-Proxy Replacement**: Constant-time $O(1)$ L4 Service Load Balancing using Google Maglev consistent hashing, Direct Server Return (DSR), and XDP NodePort acceleration.
- 🌐 **Gateway API v1.6.1 Standard**: First-class Kubernetes Gateway controller (`straitgateway.io/skgateway`) supporting `HTTPRoute`, `GRPCRoute`, `TLSRoute`, `TCPRoute`, and `UDPRoute`.
- 🌉 **Multi-Cluster Transit Gateway**: Autonomous transit daemon (`tgwd`) supporting full-mesh and hub-and-spoke topologies with 32-bit segment isolation (`TransitSegment`) and automated WireGuard mesh encryption.
- 🛡️ **Identity-Aware Zero Trust**: Numeric identity enforcement (`StraitNetworkPolicy`), BPF Linux Security Module (LSM) socket hooks, and strict default-deny policies.
- 📊 **Kernel-Level Observability**: Native Prometheus metrics, OpenTelemetry distributed tracing, and high-throughput eBPF ring buffer event streaming.

---

## Architectural Principles

1. **No Existing CNI Assumption**: StraitGateway is a complete standalone networking implementation. It directly owns and provisions the network interfaces, addresses, routes, and kernel programs.
2. **No RFC1918 / PodCIDR Assumption**: Operates cleanly over any IP space (IPv4/IPv6, non-RFC1918, overlapping VPC CIDRs) with segment translation.
3. **Decoupled Control & Datapath**: Controllers compile Kubernetes CRD state into node-level desired state. The kernel datapath never makes synchronous requests to the Kubernetes API server.
4. **Divide and Conquer**: Distinct, focused binaries for node orchestration, cluster reconciliation, transit gateway, CLI diagnostics, and packet capture.

---

## Architecture at a Glance

```
                Kubernetes API Server
     (Gateway API v1.6.1, Services, Nodes, CRDs)
                         │
                         ▼
               ┌───────────────────┐
               │   sg-controller   │ ── Control-plane manager
               └─────────┬─────────┘
                         │ Desired State Stream (gRPC)
                         ▼
               ┌───────────────────┐
               │      straitd      │ ── Privileged node daemon (DaemonSet)
               └────┬─────────┬────┘
                    │         │
       ┌────────────┘         └────────────┐
       ▼                                   ▼
┌──────────────┐                 ┌───────────────────┐
│  strait-cni  │                 │   eBPF Datapath   │
├──────────────┤                 ├───────────────────┤
│ NetKit Links │                 │ XDP Ingress Hook  │
│ Host IPAM    │                 │ TCX Egress/Ingress│
│ Linux Routing│                 │ BPF LSM & Cgroups │
└──────────────┘                 └─────────┬─────────┘
                                           │
                           ┌───────────────┴───────────────┐
                           ▼                               ▼
                 [ Pod Workloads ]               [ Linux Kernel NetKit ]
                                                           │
                                                           ▼
                                                 ┌───────────────────┐
                                                 │       tgwd        │ ── WireGuard Mesh
                                                 │  Transit Gateway  │    & Multi-Cluster
                                                 └───────────────────┘
```

---

## Component Matrix

| Binary | Source Path | Description |
| :--- | :--- | :--- |
| **`straitd`** | [`cmd/straitd`](cmd/straitd) | Node-level daemon managing NetKit links, eBPF maps, services, and local reconciliation |
| **`sg-controller`** | [`cmd/sg-controller`](cmd/sg-controller) | Cluster-wide controller reconciling Gateway API, Services, Endpoints, and Network Policies |
| **`tgwd`** | [`cmd/tgwd`](cmd/tgwd) | Multi-cluster Transit Gateway daemon managing peers, routing domains, and WireGuard tunnels |
| **`strait-cni`** | [`cmd/strait-cni`](cmd/strait-cni) | Host CNI plugin binary invoked by container runtimes (`containerd`, `CRI-O`) |
| **`sgctl`** | [`cmd/sgctl`](cmd/sgctl) | Administrative CLI for inspecting kernel BPF maps, endpoints, routes, and transit peers |
| **`sgpktcap`** | [`cmd/sgpktcap`](cmd/sgpktcap) | Datapath diagnostic tool capturing packet flows and ring buffer events in real time |

---

## Documentation

Comprehensive architectural guides and specifications are located in the [`docs/`](docs) directory:

- 📖 **[Architecture Overview](docs/architecture/overview.md)**: Core design principles, component ownership, and high-level workflows.
- ⚙️ **[Control Plane](docs/architecture/control-plane.md)**: `sg-controller` architecture, CRD reconcilers, and gRPC sync streams.
- ⚡ **[eBPF Datapath](docs/architecture/datapath.md)**: XDP ingress filtering, TCX hooks, connection tracking, and DSR routing.
- 🔌 **[CNI & IPAM](docs/architecture/cni.md)**: CNI plugin lifecycle, IPAM allocation algorithms, and runtime interactions.
- 🚀 **[NetKit Subsystem](docs/architecture/netkit.md)**: NetKit device creation, peer anchoring, and kernel performance advantages over veth.
- 🛡️ **[Security & Policy](docs/architecture/policy.md)**: Numeric identity generation, `StraitNetworkPolicy`, and BPF LSM socket controls.
- 🕸️ **[Sidecarless Service Mesh](docs/architecture/mesh.md)**: Transparent mTLS, kernel traffic encryption, and per-pod proxy-less telemetry.
- 🌉 **[Transit Gateway](docs/architecture/transit-gateway.md)**: Cross-cluster topologies, segment isolation, and automated WireGuard routing.
- 📋 **[Full Capabilities Reference](docs/architecture/capabilities.md)**: Complete feature matrix and functional boundaries.
- 🔍 **[Datapath Troubleshooting Guide](docs/troubleshooting/datapath.md)**: Diagnostic recipes, BPF map inspection, and flow debugging.

---

## Quick Start

### 1. Install Gateway API CRDs
```bash
kubectl apply --server-side -f https://github.com/kubernetes-sigs/gateway-api/releases/download/v1.6.1/standard-install.yaml
```

### 2. Deploy StraitGateway via Helm
```bash
helm upgrade --install straitgateway ./straitgateway-helm \
  --namespace straitgateway-system \
  --create-namespace \
  --set global.clusterName="production-cluster" \
  --set kubeProxyReplacement.enabled=true \
  --set dataplane.overlay=Native \
  --wait --timeout=15m
```

### 3. Verify Deployment
```bash
# Verify daemonset and controller rollouts
kubectl rollout status daemonset/straitgatewayd -n straitgateway-system
kubectl rollout status deployment/straitgateway-controller -n straitgateway-system

# Verify GatewayClass registration
kubectl get gatewayclass skgateway
```

---

## Local Development & Testing

StraitGateway provides a comprehensive [`Makefile`](Makefile) for local builds, testing, and cluster orchestration:

```bash
# Build all Go binaries (amd64 and arm64)
make build

# Run unit tests
make test

# Run unit tests with race detection
make test-race

# Run code linters and vet checks
make lint
make vet

# Spin up a local Kind test cluster
make kind-create

# Deploy StraitGateway onto Kind
make kind-deploy

# Teardown the Kind test environment
make kind-delete
```

---

## Command Line Interface (`sgctl`)

`sgctl` provides direct visibility into kernel datapath state:

```bash
# View active node datapath status
sgctl status

# Dump eBPF service and load balancing maps
sgctl bpf maps dump lb

# List active transit gateway peers and segments
sgctl transit peers list
sgctl transit routes list

# Inspect pod network identities
sgctl identity list
```

---

## License

StraitGateway is licensed under the [Apache 2.0 License](LICENSE).