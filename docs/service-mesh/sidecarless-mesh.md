# Sidecarless Service Mesh

StraitGateway eliminates the sidecar proxy tax (high memory footprint, CPU overhead, and 3-4x latency penalty) by introducing a **sidecarless NodeProxy** architecture driven by eBPF socket redirection.

---

## 1. Sidecar vs. Sidecarless Comparison

```
Traditional Sidecar Model (Envoy per Pod):
[App Pod A] ---> [Sidecar Envoy] === TCP ===> [Sidecar Envoy] ---> [App Pod B]
   (2 hops)         (3 hops)                     (2 hops)        Total: 7 hops

StraitGateway Sidecarless Model:
[App Pod A] ----------------------------------------------------> [App Pod B]
      \                                                                 /
       ==== (Only if L7 inspection requested: eBPF socket splice) ======
                      [Shared NodeProxy on Host]
```

| Metric / Characteristic | Traditional Sidecar (Envoy) | StraitGateway Sidecarless |
|---|---|---|
| **Memory per Pod** | 50 MB - 150 MB (Envoy container) | **0 MB** (No pod container injected) |
| **P99 Latency (L4)** | 2.4 ms | **0.25 ms (10x faster)** |
| **L7 Traversal Overhead** | 3 context switches per request | Zero-copy socket redirection via `bpf_msg_redirect_hash` |
| **Upgrade Disruption** | Pod restart required | Zero downtime daemon upgrade |

---

## 2. In-Kernel Socket Splicing (`internal/mesh/proxy.go`)

When a workload requires L7 traffic policy enforcement (such as HTTP header-based routing, retries, or circuit breaking):

1. **Sockmap Registration**:
   * The node agent registers pod sockets into an eBPF `sock_hash` map upon connection establishment.
2. **Zero-Copy Splice**:
   * Packets are spliced directly from the application socket buffer to the shared `NodeProxy` socket buffer using `bpf_msg_redirect_hash`.
3. **Hardware-Efficient Execution**:
   * Data never traverses the physical network device stack, netfilter, or iptables rules.
