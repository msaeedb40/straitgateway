# Multi-Plane Architecture

StraitGateway is organized into four distinct planes, each with well-defined responsibilities and communication boundaries. This document describes the complete plane architecture, inter-plane communication protocols, and failure isolation model.

---

## 1. Plane Overview

```
┌──────────────────────────────────────────────────────────────────┐
│                     CONTROL PLANE                                │
│   sg-controller (in-cluster)                                     │
│   Watches: Gateway API, TransitGateway, ClusterPeering CRDs      │
│   Reconciles: route policy, peer mesh topology, identity maps    │
└────────────────────────────┬─────────────────────────────────────┘
                             │ gRPC ControlSync / REST API
┌────────────────────────────▼─────────────────────────────────────┐
│                      AGENT PLANE                                 │
│   straitd (per-node DaemonSet)                                   │
│   Manages: BPF map updates, link attach/detach, routing daemons  │
│   Runs: BGP speaker, BFD monitor, WireGuard manager, IPAM        │
└──────────┬──────────────────────────────────┬────────────────────┘
           │ BPF map writes                   │ Telemetry reads
┌──────────▼──────────────┐   ┌───────────────▼────────────────────┐
│     DATA PLANE          │   │     OBSERVABILITY PLANE            │
│   eBPF CO-RE programs   │   │   Prometheus /metrics              │
│   pinned in bpffs       │   │   Flow ring buffer consumer        │
│   XDP / TCX / cgroup    │   │   AIOps Engine + BurstDetector     │
│   hooks on all ifaces   │   │   W3C Trace Propagator             │
└─────────────────────────┘   └────────────────────────────────────┘
```

---

## 2. Control Plane (`sg-controller`)

### 2.1 Responsibilities

| Function | Implementation |
|---|---|
| Gateway API reconciliation | Watches `Gateway`, `HTTPRoute`, `GRPCRoute`, `TCPRoute` objects |
| TransitGateway lifecycle | Manages `TransitGateway` CRD → programs `sg_routes` BPF maps via Agent |
| ClusterPeering topology | Establishes WireGuard mesh topology; distributes peer public keys |
| Identity assignment | Derives 32-bit cryptographic workload identity tokens |
| SecurityPolicy enforcement | Compiles `NetworkPolicy` + `SecurityPolicy` CRDs → `sg_policies` map entries |
| Certificate management | Issues SPIFFE SVIDs via SPIRE integration |

### 2.2 Reconciliation Loop

```
Watch event (K8s API)
     │
     ▼
EnqueueObject(key)
     │
     ▼
WorkQueue → Reconcile(ctx, req)
     ├─ Fetch current desired state from K8s API
     ├─ Fetch current actual state from BPF maps (via Agent gRPC)
     ├─ Compute delta (desired ⊖ actual)
     ├─ Apply delta → Agent.SyncState(delta)
     └─ Update CRD status.conditions
```

### 2.3 Leader Election

The controller uses standard Kubernetes leader election (`leases.coordination.k8s.io`) with a 30-second lease duration and 10-second renewal deadline. Only the leader pod reconciles; followers watch but do not write.

### 2.4 CRD Types

| CRD | API Group | Version |
|---|---|---|
| `TransitGateway` | `networking.straitgateway.io` | `v1alpha2` |
| `GatewayPeer` | `networking.straitgateway.io` | `v1alpha2` |
| `ClusterPeering` | `networking.straitgateway.io` | `v1alpha2` |
| `SecurityPolicy` | `security.straitgateway.io` | `v1alpha1` |

---

## 3. Agent Plane (`straitd` / `sg-agent`)

### 3.1 Subsystem Map

```
straitd
├── BPF Loader / Map Manager    (internal/dataplane/bpf/)
│   ├── CO-RE ELF loader        (loader.go)
│   ├── Link attachment manager (attach.go)
│   └── Counter aggregator      (maps.go)
│
├── Routing Subsystems          (internal/routing/)
│   ├── FIB / RouteManager      (fib.go)
│   ├── BGP Speaker             (bgp/bgp.go)
│   └── BFD Monitor             (bfd/bfd.go)
│
├── Network Subsystems
│   ├── WireGuard Manager       (internal/wireguard/)
│   ├── IPAM Allocator          (internal/cni/ipam/)
│   └── Netkit Driver           (internal/cni/netkit.go)
│
├── Security
│   ├── SPIFFE/SPIRE SVID       (internal/security/)
│   ├── kTLS + sockops          (internal/security/)
│   └── Policy Compiler         (internal/dataplane/policy/)
│
├── Service Mesh
│   ├── NodeProxy (L7)          (internal/mesh/)
│   ├── Identity Engine         (internal/mesh/)
│   └── DNS Intercept           (internal/mesh/)
│
├── RoCEv2 / AI-ML              (internal/dataplane/roce/)
│   ├── QP Manager              (manager.go)
│   ├── DCQCN Controller        (dcqcn.go)
│   └── NCCL Topology           (nccl_topology.go)
│
└── Observability               (internal/observability/)
    ├── Metrics Collector       (metrics.go)
    ├── Flow Exporter           (flows.go)
    ├── AIOps Engine            (aiops/engine.go)
    ├── BurstDetector           (aiops/burst_detector.go)
    └── Trace Propagator        (aiops/trace_context.go)
```

### 3.2 Startup Sequence

```
1. Parse config (ConfigMap + CLI flags)
2. Mount bpffs at /sys/fs/bpf (if not already mounted)
3. Load + pin BPF map objects (reuse existing pins if present)
4. Load + attach BPF programs to all managed interfaces
5. Restore IPAM state from disk cache
6. Establish BGP sessions to configured neighbors
7. Start BFD monitors for all BGP peers with BFD enabled
8. Start WireGuard mesh (load peers from ConfigMap)
9. Start RoCEv2 Manager (if roce.enabled=true)
10. Start AIOps Engine + BurstDetectors
11. Start observability HTTP server (:9090)
12. Start control plane sync loop (gRPC to sg-controller)
```

### 3.3 Graceful Shutdown

On `SIGTERM`, straitd:
1. Stops accepting new connections via NodeProxy
2. Drains in-flight flows (30s timeout)
3. Detaches XDP programs (preserving TC programs for traffic continuity)
4. Withdraws all BGP routes (`BGP_UPDATE` with empty NLRI)
5. Stops BFD keepalives
6. Closes BPF map file descriptors (maps remain pinned — no dataplane disruption)

---

## 4. Data Plane (eBPF)

### 4.1 Program Inventory

| Program | Hook | Interface | File |
|---|---|---|---|
| XDP ingress filter | `XDP_FLAGS_DRV_MODE` | `eth0`, `mlx5_0` | `bpf/xdp/ingress.c` |
| XDP redirect | `XDP_FLAGS_DRV_MODE` | transit ifaces | `bpf/xdp/redirect.c` |
| TC ingress classifier | TCX INGRESS | all managed ifaces | `bpf/tc/ingress.c` |
| TC egress classifier | TCX EGRESS | all managed ifaces | `bpf/tc/egress.c` |
| Netkit peer ingress | TCX INGRESS | `pod0` (netkit) | `bpf/netkit/peer_ingress.c` |
| Netkit peer egress | TCX EGRESS | `pod0` (netkit) | `bpf/netkit/peer_egress.c` |
| cgroup connect4 | CGROUP SOCK_ADDR | root cgroup v2 | `bpf/cgroup/connect4.c` |
| cgroup connect6 | CGROUP SOCK_ADDR | root cgroup v2 | `bpf/cgroup/connect6.c` |
| cgroup sendmsg4 | CGROUP SOCK_MSG | root cgroup v2 | `bpf/cgroup/sendmsg4.c` |
| cgroup sendmsg6 | CGROUP SOCK_MSG | root cgroup v2 | `bpf/cgroup/sendmsg6.c` |
| Service LB | TCX (ingress) | all | `bpf/service/lb.c` |
| NAT engine | TCX (egress) | all | `bpf/service/nat.c` |
| Conntrack | TCX (ingress + egress) | all | `bpf/service/conntrack.c` |
| RoCEv2 XDP | `XDP_FLAGS_DRV_MODE` | `mlx5_0` | `bpf/roce/roce_xdp.c` |
| RoCEv2 TC egress | TCX EGRESS | `mlx5_0` | `bpf/roce/roce_tc.c` |

### 4.2 BPF Map Architecture

All maps are pinned under `/sys/fs/bpf/straitgateway/` and survive daemon restarts without dropping packets:

```
/sys/fs/bpf/straitgateway/
├── sg_routes          LPM_TRIE   — IPv4 routing table
├── sg_nexthops        HASH       — NextHop IP/MAC/ifindex
├── sg_neighbors       HASH       — L2 neighbor cache
├── sg_services        HASH       — Service VIP → endpoint set
├── sg_endpoints       HASH       — Service backend list
├── sg_conntrack       LRU_HASH   — 5-tuple flow state
├── sg_nat             HASH       — NAT translation entries
├── sg_policies        HASH       — Security policy ACLs
├── sg_identities      HASH       — Workload identity tokens
├── sg_counters        PERCPU_ARRAY — Throughput/drop counters
├── sg_events          RINGBUF    — Flow telemetry ring buffer
└── roce/
    ├── sg_qp_table         HASH        — 128K QP state entries
    ├── sg_roce_counters    PERCPU_HASH — Per-QPN byte/packet counters
    ├── sg_gpu_topo         HASH        — 4K GPU peer topology
    ├── sg_roce_events      RINGBUF     — 1MiB ECN/CNP event ring
    └── sg_pfc_config       ARRAY       — PFC per-priority config
```

### 4.3 Packet Processing Order

For a packet arriving on `eth0`:

```
NIC DMA → XDP hook (ingress.c)
    │  DDoS check → XDP_DROP if policy hit
    │  RoCEv2 check (UDP dst=4791) → xdp_pass to TCX
    ▼
TCX INGRESS hook (tc/ingress.c)
    │  L3 route lookup (sg_routes LPM)
    │  Policy check (sg_policies)
    │  Conntrack lookup / new entry creation
    │  Service VIP rewrite (sg_services → sg_endpoints)
    │  Identity check (sg_identities)
    │  Telemetry → sg_events ring buffer
    ▼
Linux netstack (for local delivery)
    OR
XDP_TX / bpf_redirect_map() for forwarding
```

### 4.4 CO-RE and BTF

All BPF programs are compiled with Clang using `-g -O2 -target bpf -D__TARGET_ARCH_x86` and the `vmlinux.h` generated from the target kernel's BTF. CO-RE relocations allow the same ELF object to run on any kernel that supports the accessed fields, with automatic struct-member offset adjustment.

---

## 5. Observability Plane

### 5.1 Components

| Component | Port | Protocol | Description |
|---|---|---|---|
| Prometheus exporter | `:9090/metrics` | HTTP | All dataplane + routing + AIOps metrics |
| Flow log API | `:9090/api/v1/flows` | HTTP/JSON | Last N flow log entries |
| eBPF program status | `:9090/api/v1/ebpf` | HTTP/JSON | Verifier stats, run counts |
| Health endpoint | `:9090/health` | HTTP | Liveness probe |
| Ready endpoint | `:9090/ready` | HTTP | Readiness probe (maps loaded) |

### 5.2 Ring Buffer Consumer

The `flows.go` FlowExporter continuously reads from the `sg_events` BPF ring buffer:

```
sg_events ring buffer
     │  bpf_ringbuf_reserve / bpf_ringbuf_submit (in-kernel)
     ▼
ringbuf.Reader (cilium/ebpf)
     │  struct telemetry_event
     ▼
FlowLogEntry → in-memory circular buffer (last 10,000 flows)
     │
     ├─► Prometheus histogram (bytes per flow)
     └─► /api/v1/flows JSON API
```

### 5.3 AIOps Integration

The AIOps subsystem hooks into the observability plane by:

1. Subscribing `BurstDetector.OnBurst` to the AIOps `Engine.HandleBurst` callback
2. Receiving `RoCECongestionEvent` from `sg_roce_events` ring buffer for CNP telemetry
3. Annotating `FlowLogEntry` with W3C `TraceID` + `SpanID` via `Propagator.AnnotateFlow`

---

## 6. Inter-Plane Communication

### 6.1 Control Plane → Agent Plane

- **Protocol**: gRPC (TLS + mTLS with SPIFFE SVIDs)
- **Service**: `ControlSyncService` (bi-directional stream)
- **Direction**: Controller pushes diffs; Agent streams telemetry back
- **Reconnect**: Exponential backoff with jitter (max 30s)

### 6.2 Agent Plane → Data Plane

- **Write path**: Direct BPF map operations via `cilium/ebpf` Go library
- **Read path**: Map iteration for counter aggregation; ring buffer reads for events
- **Atomicity**: Batch updates use `BPF_MAP_UPDATE_BATCH` syscall for consistency
- **No packet disruption**: Map updates are instantaneous from the kernel's perspective; programs see a consistent view without quiescing

### 6.3 Data Plane → Observability Plane

- **Counters**: Per-CPU arrays summed on each 10-second scrape interval
- **Events**: Lock-free ring buffer — kernel writes, userspace reads; no shared mutex
- **Ring buffer sizing**: `sg_events` = 16 MiB; `sg_roce_events` = 1 MiB

---

## 7. Failure Isolation Model

### 7.1 Control Plane Failure

- **Effect**: No new policy or route changes can be applied
- **Dataplane impact**: Zero — existing BPF maps continue to operate
- **Recovery**: sg-controller pod restart; agent reconciles full state on reconnect

### 7.2 Agent Plane Failure (`straitd` crash)

- **Effect**: No new BPF map updates; existing state frozen
- **Dataplane impact**: Zero — kernel BPF programs run autonomously
- **Recovery**: straitd restart; maps reloaded from bpffs pins (no state loss)

### 7.3 BPF Program Crash (BPF_EXIT_CODE != 0)

- **Effect**: Program exits, returning `XDP_PASS` or `TC_ACT_OK` by default
- **Dataplane impact**: Traffic passes without policy enforcement (fail-open)
- **Detection**: `sg_counters[PROG_ERROR]` increments; Prometheus alert fires
- **Recovery**: Agent detects via `/proc/bpf` poll; reloads and reattaches program

### 7.4 BFD Session Down

- **Effect**: BFD fires `handlePeerFailure()` within one detection interval (50ms × 3 = 150ms)
- **Dataplane impact**: BGP peer routes withdrawn from `sg_routes` within ~200ms total
- **Recovery**: BFD re-establish → BGP `UPDATE` → FIB sync → map update

### 7.5 RoCEv2 QP Error

- **Effect**: XDP program drops all packets for that QPN
- **Dataplane impact**: RDMA application receives `IBV_WC_REM_ABORT_ERR`
- **Recovery**: Application re-creates the QP; straitd receives QP state update and re-inserts into `sg_qp_table` in `RESET` state

---

## 8. Scalability Design Points

| Dimension | Limit | Mechanism |
|---|---|---|
| Pods per node | 250 | Netkit devices + IPAM bitmap |
| BPF map capacity | 131,072 entries (most maps) | LRU eviction on conntrack |
| BGP peers | 128 | Per-VRF peer table |
| GPU Queue Pairs | 131,072 | `sg_qp_table` max_entries |
| GPU peer topology | 4,096 | `sg_gpu_topo` max_entries |
| Metrics scrape interval | 10 seconds | Prometheus pull |
| Flow ring buffer throughput | ~1M events/s | Lock-free bpf_ringbuf |
| Trace flow cache | 65,536 flows | Bounded map + TTL eviction |
| Cluster peers | 256 | WireGuard mesh topology |
