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
| **Language** | **Go 1.27.1** | Controllers, agents, networking services, and CLI |
| **Kernel** | **Linux 6.8+ (6.12+ LTS Recommended)** | Host networking, TCX, netkit, and eBPF dataplane |
| **Dataplane** | **eBPF CO-RE** (`cilium/ebpf`) | Programmable, high-performance packet processing across kernels |
| **Pod Networking** | **netkit** | Next-generation direct pod-to-kernel packet handoff |
| **Security & Tunnels** | **WireGuard** | High-speed encrypted inter-node and inter-cluster overlay |
| **Dynamic Routing** | **BGP** | Dynamic route exchange with external networks & top-of-rack switches |
| **Failure Detection**| **BFD** | Sub-second detection of routing peer failures |
| **Orchestration** | **Kubernetes Gateway API v1.6.1** | Native declarative traffic management and transit gateway routing |
| **Observability** | **Prometheus + OTel + Loki** | eBPF ring buffer telemetry, metrics, traces, and flow logs |
| **Web UI** | **React 19 + TypeScript + D3.js** | Live cluster topology & observability without third-party query bloat |

---

## Building StraitGateway

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
```

---

## License

StraitGateway is licensed under the Apache 2.0 License. See [LICENSE](LICENSE) for details.