# StraitGateway Architecture Overview

StraitGateway is a Kubernetes-native networking platform providing:

- **Complete standalone CNI** (`strait-cni`) — no Calico, Cilium, or Flannel required
- **eBPF-first datapath** — replaces iptables, IPVS, and kube-proxy
- **Netkit** pod-to-node links — replaces veth pairs
- **Transit Gateway** (`tgwd`) — cross-segment / multi-cluster networking
- **Sidecarless service mesh** — via node-level eBPF datapath
- **Gateway API v1.6.1** — Kubernetes-native routing

## Architecture Principals

| Rule | Constraint |
|------|-----------|
| No CNI assumption | StraitGateway IS the CNI |
| No RFC1918 assumption | Works with any PodCIDR |
| Own what you manage | No shared ownership of datapath objects |
| Divide and conquer | Clear component boundaries |

## Component Ownership

| Component | Owner | Responsibility |
|-----------|-------|---------------|
| Kubernetes desired state | SG Controller | Watches API Server, produces config |
| Node networking | StraitD | Netkit, eBPF, services, routes, policy |
| Pod lifecycle | strait-cni | ADD / DEL / CHECK / IPAM |
| Transit topology | TGWD | Peers, routes, segments, gateways |
| Packet/socket processing | eBPF Datapath | Kernel-level fast path |
| Pod-to-node link | Netkit | Replaces veth |
| Traffic API | Gateway API | Gateways, routes, listeners |
| Metrics | Prometheus | High-cardinality metrics |
| Telemetry | OpenTelemetry | Traces, logs, spans |

## High-Level Architecture Diagram

```
Kubernetes API Server
        │
┌───────▼────────┐
│  SG Controller │ ← watches Gateway API, Services, Nodes, CRDs
└───────┬────────┘
        │ Desired State
┌───────▼────────┐
│    StraitD     │ ← node agent (DaemonSet)
│  Node Agent    │
│ ─────────────  │
│ CNI Manager    │
│ eBPF Manager   │
│ Netkit Manager │
│ Policy Manager │
│ Service Manager│
│ Route Manager  │
│ Gateway Manager│
└──┬─────────┬──┘
   │         │
 CNI      eBPF Datapath
           XDP / TCX / cgroup / LSM / kprobe
           │
        Linux Kernel
           │
        Netkit ←→ Network Interfaces
           │
      Pod A / Pod B / Pod C

TGWD (transit-gateway daemon, optional)
  ↕ internal API
StraitD
```

## Data Flow: Pod Egress

```
Pod network namespace
    │ Netkit peer
    ▼
Netkit parent (host side)
    │ TCX hook
    ▼
eBPF: service LB → policy check → routing → flow tracking
    │
    ▼
Linux network / physical interface
```
