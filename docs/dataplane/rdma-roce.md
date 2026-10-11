# RDMA / RoCEv2 Dataplane

StraitGateway implements a full-stack RoCEv2 (RDMA over Converged Ethernet v2) dataplane for AI/ML workloads, enabling RDMA traffic to traverse the StraitGateway packet processing pipeline with hardware-equivalent performance.

---

## 1. Background: RoCEv2 and AI/ML Networking

### 1.1 What is RoCEv2?

RoCEv2 encapsulates InfiniBand Transport Protocol (IB-TL) packets inside standard UDP/IP, using the IANA-assigned destination port **4791**. It enables:

- **Zero-copy RDMA** — direct memory-to-memory transfers without CPU involvement
- **Kernel bypass** — user-space applications communicate directly via the NIC's RDMA queue pairs
- **High-bandwidth, ultra-low latency** — critical for distributed AI training (AllReduce, AllGather)

### 1.2 Why RoCEv2 in a CNI?

In Kubernetes GPU clusters, the CNI is responsible for the network namespace and interface lifecycle of GPU pods. StraitGateway's RoCEv2 dataplane provides:

| Capability | Without StraitGateway | With StraitGateway |
|---|---|---|
| Congestion control | NIC-specific DCQCN only | eBPF-accelerated ECN + userspace DCQCN |
| PFC enforcement | Manual NIC configuration | Automated per-QP DSCP marking |
| Topology-aware routing | Static InfiniBand fabric | Dynamic NCCL GPU topology BPF map |
| Observability | Vendor-specific tools | Per-QPN byte/packet/CNP Prometheus metrics |
| Micro-burst protection | None | Phase 14 AIOps burst detector integration |

---

## 2. RoCEv2 Header Parsing

### 2.1 Packet Format

```
Ethernet | IPv4/IPv6 | UDP (dst=4791) | BTH (12B) | [AETH/RETH/DETH] | Payload | ICRC (4B)
```

### 2.2 BTH — Base Transport Header

Every RoCEv2 packet carries a 12-byte Base Transport Header (BTH):

```
 0                   1                   2                   3
 0 1 2 3 4 5 6 7 8 9 0 1 2 3 4 5 6 7 8 9 0 1 2 3 4 5 6 7 8 9 0 1
+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
|    OpCode     |SE|M|PadCnt |TVer|           P_Key             |
+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
|    Reserved   |              Destination QPN (24-bit)          |
+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
|A|  Reserved   |              Packet Sequence Number (24-bit)   |
+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
```

Key fields parsed by StraitGateway's XDP program:

| Field | Bits | Use |
|---|---|---|
| `OpCode` | [7:0] | Classify data vs. ACK vs. CNP |
| `Destination QPN` | [23:0] | Look up `sg_qp_table` |
| `Packet Sequence Number` | [23:0] | Include in congestion events |
| `ACK Req` bit | [31] of word 2 | Trigger CNP generation |

### 2.3 AETH — ACK Extended Transport Header

Follows BTH on RC ACK packets (OpCode `0x11`):

```
 0                   1                   2                   3
 0 1 2 3 4 5 6 7 8 9 0 1 2 3 4 5 6 7 8 9 0 1 2 3 4 5 6 7 8 9 0 1
+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
|   Syndrome    |            Message Sequence Number (24-bit)    |
+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
```

Syndrome values:

| Value | Meaning |
|---|---|
| `0x00` | Pure ACK |
| `0x20` | Receiver-Not-Ready (RNR) |
| `0x60` | NAK (sequence error) |

### 2.4 OpCode Reference

| OpCode | Name | Extended Header |
|---|---|---|
| `0x00` | RC SEND FIRST | — |
| `0x04` | RC SEND ONLY | — |
| `0x06` | RC RDMA WRITE FIRST | RETH |
| `0x0A` | RC RDMA WRITE ONLY | RETH |
| `0x0C` | RC RDMA READ REQUEST | RETH |
| `0x11` | RC ACK | AETH |
| `0x81` | CNP (Congestion Notification Packet) | — |
| `0x64` | UD SEND ONLY | DETH |

---

## 3. BPF Programs

### 3.1 XDP Ingress (`bpf/roce/roce_xdp.c`)

Attached at `XDP_FLAGS_DRV_MODE` (native driver hook) on the RDMA-capable NIC:

```
ingress packet
   │
   ├─ ETH parse ──► not IP → XDP_PASS
   │
   ├─ IP parse ───► not UDP → XDP_PASS
   │
   ├─ UDP parse ──► dst≠4791 → XDP_PASS
   │
   ├─ BTH parse
   │      │
   │      ├─ QPN lookup (sg_qp_table)
   │      │    └─ state ERR/RESET → XDP_DROP
   │      │
   │      ├─ PFC DSCP enforcement
   │      │    └─ rewrite IP.TOS if DSCP≠AF31 for RoCE priority
   │      │
   │      ├─ Per-QPN counter update (sg_roce_counters)
   │      │
   │      ├─ ECN-CE detection → emit sg_roce_events ring buffer
   │      │
   │      └─ NCCL topo lookup → NVLink domain → XDP_TX hairpin
   │
   └─ XDP_PASS (to host stack / RDMA driver)
```

### 3.2 TC Egress (`bpf/roce/roce_tc.c`)

Attached via TCX on the RDMA interface egress path:

- **PFC DSCP rewrite**: Enforces DSCP=AF31 (26) on outbound RoCEv2 data packets for switch-level PFC queuing. Uses `bpf_l3_csum_replace` for correct IP header checksum.
- **ECN ECT(0) marking**: Marks `ECN=ECT(0)` on outbound packets when the QP's traffic class has `ecn_capable=1`, enabling end-to-end DCQCN.
- **Egress accounting**: Updates `sg_roce_counters` for outbound bytes/packets.

---

## 4. BPF Maps

All RoCEv2 maps are pinned under `/sys/fs/bpf/straitgateway/roce/`:

### `sg_qp_table` — Queue Pair State Table

```
Key:   { qpn: u32 }
Value: { state: u8, pfc_priority: u8, mtu: u16, remote_ip: u32,
         rq_psn: u32, sq_psn: u32, last_seen_ns: u64 }
Capacity: 128,000 QPNs
Type: BPF_MAP_TYPE_HASH
```

QP states:

| Value | State | XDP Action |
|---|---|---|
| 0 | RESET | DROP |
| 1 | INIT | PASS |
| 2 | RTR | PASS |
| 3 | RTS | PASS |
| 4 | SQD | PASS |
| 5 | SQE | PASS |
| 6 | ERR | DROP |
| 7 | DRAIN | PASS |

### `sg_roce_counters` — Per-CPU Per-QPN Counters

```
Key:   { qpn: u32, cpu: u8 }
Value: { bytes: u64, packets: u64, cnp_sent: u64,
         ecn_ce_seen: u64, drops: u64 }
Type: BPF_MAP_TYPE_PERCPU_HASH
```

### `sg_gpu_topo` — NCCL GPU Topology

```
Key:   { remote_ip: u32 }
Value: { domain: u8, hop_count: u8, bandwidth_gbps: u16,
         local_gpu_id: u32, remote_gpu_id: u32,
         nvlink_mask: u32, last_updated_ns: u64 }
Capacity: 4,096 peers
Type: BPF_MAP_TYPE_HASH
```

Topology domains:

| Domain | Value | Routing Action |
|---|---|---|
| SAME_GPU | 0 | XDP_TX hairpin |
| NVLINK | 1 | XDP_TX hairpin |
| PCIE | 2 | Local switch path |
| SWITCH | 3 | ToR path |
| REMOTE | 4 | Standard RoCEv2 over IP |

### `sg_roce_events` — ECN/CNP Ring Buffer

1 MiB ring buffer delivering `roce_congestion_event` structs to userspace for CNP generation:

```c
struct roce_congestion_event {
    u64 timestamp;
    u32 src_ip, dst_ip;
    u32 qpn, psn;
    u8  ecn, opcode;
};
```

### `sg_pfc_config` — PFC Per-Priority Configuration

8-entry array (one per IEEE 802.1p PCP value):

```
Key:   u32 (priority 0–7)
Value: { enabled: u8, dscp: u8, ecn_capable: u8,
         pause_time_us: u32, xon_threshold: u32, xoff_threshold: u32 }
```

Default for RoCEv2 (priority 3): `DSCP=AF31`, `ecn_capable=1`, `pause=65535µs`.

---

## 5. DCQCN Rate Controller

StraitGateway implements the **DCQCN** (Data Center Quantized Congestion Notification) algorithm in Go ([`internal/dataplane/roce/dcqcn.go`](../internal/dataplane/roce/dcqcn.go)):

### 5.1 Algorithm Summary

**Receiver side (eBPF XDP)**:
1. Detect `ECN=CE` on incoming RoCEv2 packet
2. Emit `roce_congestion_event` to `sg_roce_events` ring buffer
3. Userspace generates CNP (Congestion Notification Packet) with BTH OpCode `0x81`

**Sender side (Go `RateController`)**:

On CNP received:
```
α_new = (1 - g) × α + g × 1           // g = AlphaDecay (default 0.0625)
Rc_new = Rc × (1 - α/2)               // Rate halving weighted by α
Rc_new = max(Rc_new, MinRateMbps)
```

On timer expiry (no CNP):
```
α_new = (1 - g) × α                   // α decays toward 0
Rc_new = Rc + RAI                      // Additive increase (5 Mbps/tick)
// After K consecutive non-CNP timers → HAI mode:
Rc_new = Rc + RHAI                     // Hyperactive increase (50 Mbps/tick)
```

### 5.2 Default Parameters

| Parameter | Default | Description |
|---|---|---|
| `Alpha` | 1.0 | Initial ECN fraction |
| `AlphaDecay` (g) | 0.0625 | Alpha convergence speed |
| `RateDecreaseInterval` | 4ms | Minimum time between reductions |
| `RateAI` | 5 Mbps | Additive increase step |
| `RateHAI` | 50 Mbps | Hyperactive increase step |
| `MinRateMbps` | 100 Mbps | Rate floor |
| `MaxRateMbps` | 100,000 Mbps | Rate ceiling (100 Gbps) |
| `ByteCounterThreshold` | 150,000 bytes | Fast recovery trigger |
| `TimerPeriod` | 55ms | DCQCN timer granularity (IBTA default) |
| `HAITimerCount` | 5 | Non-CNP ticks before HAI mode |

---

## 6. NCCL Topology-Aware Routing

### 6.1 Topology Discovery

The [`TopologyGraph`](../internal/dataplane/roce/nccl_topology.go) is built from:

1. **`/sys/class/infiniband/*/ports/*/gids/*`** — enumerates RDMA GIDs to discover GPU interfaces and map them to virtual IPv4 addresses
2. **`nvidia-smi topo -m`** — parses the NVLink adjacency matrix (NVLink lane counts)
3. **PCIe bus IDs** — classifies GPU pairs on the same root complex as PCIe-proximity peers

### 6.2 Domain Classification

```
GPU A → GPU B topology:
  Same GID prefix     →  SAME_GPU (shared memory)
  NVLink adjacency    →  NVLINK   (600 GB/s aggregate)
  Same PCIe root      →  PCIE     (64 GB/s PCIe Gen4 x16)
  Same rack/ToR       →  SWITCH   (100 Gbps Ethernet)
  Different rack      →  REMOTE   (standard RoCEv2 over IP)
```

### 6.3 BPF Routing Decisions

| Domain | XDP Action | Rationale |
|---|---|---|
| SAME_GPU / NVLINK | `XDP_TX` | Hairpin to loopback (nanosecond latency) |
| PCIE | `XDP_PASS` | Let kernel route to local interface |
| SWITCH / REMOTE | `XDP_PASS` | Standard network path with PFC/ECN |

---

## 7. QP Lifecycle Management

The Go-side [`Manager`](../internal/dataplane/roce/manager.go) enforces the IB Spec Vol 1 Table 69 state machine:

```
RESET ──► INIT ──► RTR ──► RTS ──► SQD ──► ERR ──► RESET
   ▲                │       │       │
   └────────────────┴───────┴───────┘  (any → RESET on error)
```

Valid transitions enforced in `validTransition()` — invalid transitions return an error without modifying the BPF map. QPs in `ERR` or `RESET` state receive `XDP_DROP` in the dataplane.

---

## 8. Prometheus Metrics

RoCEv2 metrics exposed at `/metrics` on the `straitd` observability server:

| Metric | Type | Labels | Description |
|---|---|---|---|
| `sg_roce_qp_bytes_total` | Counter | `qpn`, `node` | Total bytes per QPN |
| `sg_roce_qp_packets_total` | Counter | `qpn`, `node` | Total packets per QPN |
| `sg_roce_cnp_sent_total` | Counter | `qpn`, `node` | CNPs generated |
| `sg_roce_ecn_ce_total` | Counter | `qpn`, `node` | ECN-CE packets received |
| `sg_roce_drops_total` | Counter | `qpn`, `node` | Packets dropped (bad QP state) |
| `sg_roce_active_qps` | Gauge | `node` | Currently active Queue Pairs |
| `sg_roce_dcqcn_rate_mbps` | Gauge | `qpn`, `node` | Current DCQCN rate limit |

---

## 9. Configuration Example

```yaml
# straitd ConfigMap (RoCEv2 section)
apiVersion: v1
kind: ConfigMap
metadata:
  name: straitgateway-config
  namespace: kube-system
data:
  config.yaml: |
    roce:
      enabled: true
      interface: "mlx5_0"       # RDMA-capable NIC
      xdpMode: "driver"         # driver | generic
      pfc:
        priority: 3             # IEEE 802.1p PCP for RoCEv2
        dscp: 26                # AF31
        ecnCapable: true
        pauseTimeUs: 65535
        xonThreshold: 262144    # 256 KiB
        xoffThreshold: 524288   # 512 KiB
      dcqcn:
        alpha: 1.0
        alphaDecay: 0.0625
        rateAIMbps: 5
        rateHAIMbps: 50
        minRateMbps: 100
        maxRateMbps: 100000
        timerPeriodMs: 55
        haiTimerCount: 5
      topology:
        refreshIntervalSeconds: 60
```

---

## 10. Operational Reference

### Inspect QP State Table

```bash
# List all active Queue Pairs
sgctl roce qp list

# Show detail for a specific QPN
sgctl roce qp get --qpn 1234

# Force a QP into error state (for testing drain)
sgctl roce qp set-state --qpn 1234 --state ERR
```

### Inspect Per-QPN Counters

```bash
# Aggregate counters across all CPUs
sgctl roce counters --qpn 1234

# Watch live (1s refresh)
sgctl roce counters --qpn 1234 --watch
```

### Dump GPU Topology

```bash
# Show NCCL topology table
sgctl roce topology

# Force refresh from nvidia-smi
sgctl roce topology --refresh
```

### Check DCQCN Rate Controller

```bash
# Show all QPN rate states
sgctl roce dcqcn stats

# Simulate a CNP for testing
sgctl roce dcqcn cnp --qpn 1234
```

### Verify PFC Configuration

```bash
# Dump sg_pfc_config array (all 8 priorities)
sgctl bpf dump-map --map sg_pfc_config --pin /sys/fs/bpf/straitgateway/roce/sg_pfc_config
```

---

## 11. Troubleshooting

### XDP Program Not Attaching

```
Error: attach XDP (driver + generic fallback): operation not permitted
```
**Cause**: Insufficient capability. RoCEv2 XDP requires `CAP_BPF` + `CAP_NET_ADMIN`.
**Fix**: Ensure `straitd` runs with `securityContext.capabilities.add: [NET_ADMIN, BPF]`.

### High Drop Rate

```bash
sgctl roce counters --all | grep drops
```
If `drops_total` is increasing, check `sg_qp_table` for QPNs in `ERR`/`RESET` state:
```bash
sgctl roce qp list --state ERR
```
Stale QP entries accumulate when the RDMA application crashes without cleaning up. The Manager's TTL-based eviction (`last_seen_ns` > 5 minutes) will remove them automatically.

### ECN Not Working

Verify the NIC supports ECN in hardware:
```bash
ethtool --show-features mlx5_0 | grep -i ecn
```
Verify ECT(0) is being set on egress:
```bash
tcpdump -i mlx5_0 'udp port 4791' -n | head -20
# Look for DSCP=26 (0x68) in IP ToS byte
```
