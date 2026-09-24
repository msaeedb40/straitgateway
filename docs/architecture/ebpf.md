# StraitGateway eBPF Subsystem Architecture

## Overview
StraitGateway operates an eBPF datapath that powers container networking, Service load balancing, security policies, and transit routing in the Linux kernel without requiring `iptables`, `nftables`, or `kube-proxy`.

## Core Invariants
1. **Single Authoritative Lifecycle**: All BPF programs, links, and maps are loaded and managed exclusively by `straitd` through `internal/straitd/ebpf`.
2. **CO-RE Portability**: All BPF programs use Compile Once - Run Everywhere (CO-RE) via `vmlinux.h` and Clang/LLVM 22 toolchains.
3. **Explicit ABI Structs**: Go userspace structs and BPF C structs share explicit memory layouts in `internal/bpf/abi/` to avoid marshalling size mismatches.
4. **Resilience & Restart**: BPF maps are pinned to `/sys/fs/bpf/straitgateway/` so in-flight packet processing continues uninterrupted if `straitd` restarts.

## BPF Program Hierarchy
| Hook Point | Program Type | Target Attachment | Responsibility |
| :--- | :--- | :--- | :--- |
| **Ingress Uplink** | `XDP` (`sg_xdp.bpf.c`) | Node physical interface | Fast-path drop, early VIP filtering, DDoS mitigation |
| **Pod Ingress/Egress** | `TCX` (`sg_tcx.bpf.c`) | Netkit parent device | Policy enforcement, packet routing, flow telemetry |
| **Socket Connection** | `cgroup/sock_addr` (`sg_sock_lb.bpf.c`) | Root / Pod cgroup v2 | Transparent Service VIP translation on `connect4`/`sendmsg4` |
| **Security LSM** | `LSM` (`sg_lsm.bpf.c`) | Socket hooks | Kernel-level identity security and socket isolation |
| **Diagnostics** | `kprobe` / `tracepoint` | Kernel network stack | Latency tracing, packet drop analysis, ring buffer events |

## Shared BPF Maps
- **`sg_services`** (`BPF_MAP_TYPE_HASH`): Maps `(VIP, Port, Proto)` to active backend endpoints.
- **`sg_routes`** (`BPF_MAP_TYPE_LPM_TRIE`): Longest-prefix match for pod CIDRs and transit routes.
- **`sg_policy`** (`BPF_MAP_TYPE_HASH`): Evaluated ingress/egress policy allow/deny decisions.
- **`sg_peers`** (`BPF_MAP_TYPE_HASH`): Transit gateway peer definitions and next-hop encapsulation.
- **`sg_flow`** (`BPF_MAP_TYPE_LRU_HASH` & `BPF_MAP_TYPE_RINGBUF`): Connection tracking and flow telemetry events.
