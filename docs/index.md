# StraitGateway

<div align="center">
  <img src="https://raw.githubusercontent.com/straitgateway/straitgateway/main/logo/straitgateway.png" alt="StraitGateway Logo" width="160" />
  <h3>Kubernetes-Native CNI, eBPF Datapath, Sidecarless Service Mesh & Transit Gateway</h3>
</div>

---

## Overview

**StraitGateway** is a next-generation Kubernetes networking stack designed to eliminate the overhead of traditional iptables/IPVS and sidecar proxies. Powered by **Linux Netkit** and **eBPF (CO-RE)**, StraitGateway provides high-performance container networking, fine-grained L3/L4/CEL network policies, kernel-level service load balancing, and multi-cluster transit gateway connectivity.

### Key Capabilities

- **Autonomous CNI & IPAM**: Full CNI implementation (`ADD`, `DEL`, `CHECK`) built from scratch without external CNI dependencies (no Calico, Cilium, or Flannel required).
- **Netkit Datapath**: Replaces legacy `veth` pairs with Linux Netkit for ultra-low latency Pod-to-node data transfer.
- **eBPF-First Processing**: TCX, XDP, cgroup socket load balancing, and BPF ring buffers for zero-overhead packet routing.
- **kube-proxy Replacement**: Kernel-native ClusterIP, NodePort, and LoadBalancer implementation with Maglev and Direct Server Return (DSR).
- **Sidecarless Service Mesh**: Mutual TLS, stateful connection tracking, and L4/L7 flow observability without sidecar container injection.
- **Transit Gateway (TGWD)**: Flexible cross-cluster and cross-segment topologies (hub-and-spoke, full mesh, peer-to-peer).
- **Kubernetes Gateway API v1.6.1**: First-class declarative ingress, gateway, and route management.
- **High-Resolution Observability**: Built-in Prometheus metrics and OpenTelemetry tracing.

---

## Quickstart: Install via Helm

StraitGateway is packaged and distributed as an official Helm chart hosted on GitHub Pages:

```bash
# 1. Add StraitGateway Helm repository
helm repo add straitgateway https://straitgateway.github.io/straitgateway/
helm repo update

# 2. Install StraitGateway into kube-system
helm install straitgateway straitgateway/straitgateway \
  --namespace kube-system \
  --set datapath.mode=netkit \
  --set networking.kubeProxyReplacement.enabled=true
```

For custom configurations, see the [Helm Configuration Guide](guides/helm-configuration.md).

---

## Architecture at a Glance

```
                         Kubernetes API Server
                                  │
                                  ▼
                         ┌─────────────────┐
                         │  SG Controller  │ (Desired State)
                         └────────┬────────┘
                                  │ Node API / gRPC
                   ┌──────────────┴──────────────┐
                   ▼                             ▼
              ┌─────────┐                  ┌───────────┐
              │ StraitD │ (Node Agent)     │   TGWD    │ (Transit Gateway)
              └────┬────┘                  └───────────┘
                   │
            ┌──────┼──────────┐
            ▼      ▼          ▼
          IPAM   Netkit     eBPF Manager
            │      │          │
            │      │     ┌────┼───────────┐
            │      │     ▼    ▼           ▼
            │      │    TCX  XDP      Socket LB
            │      │
            └──────┴─────────────┐
                                 ▼
                            Pod Network (Netkit)
                                 │
                                 ▼
                            Linux Kernel
                                 │
                                 ▼
                                NIC
```

---

## Documentation Navigation

- **Architecture Deep-Dives**:
  - [Overview](architecture/overview.md)
  - [Control Plane (`sg-controller`)](architecture/control-plane.md)
  - [Node Agent & Datapath (`straitd`)](architecture/datapath.md)
  - [CNI Plugin & IPAM](architecture/cni.md)
  - [Netkit Pod Interconnect](architecture/netkit.md)
  - [eBPF Architecture & CO-RE](architecture/ebpf.md)
  - [Zero-Trust Policy & CEL](architecture/policy.md)
  - [Sidecarless Service Mesh](architecture/mesh.md)
  - [Transit Gateway (`tgwd`)](architecture/transit-gateway.md)
  - [Kernel Capabilities Matrix](architecture/capabilities.md)
- **Tools & Operations**:
  - `sgctl`: Administrative CLI for status, health, flows, and configuration.
  - `sgpktcap`: Kernel eBPF packet capture and flow inspection.
