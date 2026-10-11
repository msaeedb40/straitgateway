# Telemetry & Flow Monitoring

StraitGateway provides low-overhead, in-kernel telemetry that exports Prometheus metrics and real-time 5-tuple flow logs.

---

## 1. Prometheus Metrics Exporter

Exposed by default on TCP `:9090` at `/metrics` and `/observability/metrics`.

### 1.1. Core Metric Dictionary

| Metric Name | Type | Labels | Description |
|---|---|---|---|
| `straitgateway_packets_total` | Counter | `direction` (ingress/egress), `interface` | Total packets processed by eBPF datapath |
| `straitgateway_bytes_total` | Counter | `direction`, `interface` | Total bytes processed by eBPF datapath |
| `straitgateway_policy_drops_total` | Counter | `src_identity`, `dst_identity`, `rule_id` | Packets dropped due to security policy violations |
| `straitgateway_fib_lookups_total` | Counter | `vrf`, `result` (hit/miss) | Forwarding Information Base routing lookups |
| `straitgateway_conntrack_entries_active` | Gauge | `protocol` | Active 5-tuple connection tracking sessions |
| `straitgateway_bgp_session_state` | Gauge | `peer_ip`, `asn` | BGP session status (1 = Established, 0 = Down) |
| `straitgateway_bfd_session_state` | Gauge | `peer_ip` | BFD session status (1 = Up, 0 = Down) |
| `straitgateway_wireguard_peers_active` | Gauge | `interface` | Total active WireGuard tunnels in mesh |
| `straitgateway_clustermesh_peers_total` | Gauge | `status` | Remote cluster peering connections |
| `straitgateway_clustermesh_active_routes`| Gauge | `cluster_id` | Synchronized cross-cluster transit routes |

---

## 2. Real-Time Flow Logging (`sg_events` Ring Buffer)

The eBPF ring buffer stream `sg_events` outputs 5-tuple flow events directly to `straitd`:

```json
{
  "timestamp": "2026-10-08T10:15:30.124Z",
  "source_ip": "10.244.1.15",
  "destination_ip": "10.96.0.1",
  "source_port": 42150,
  "destination_port": 443,
  "protocol": "TCP",
  "source_identity": 1001,
  "destination_identity": 1,
  "verdict": "ALLOW",
  "packets": 48,
  "bytes": 5210
}
```

* **Zero Memory Overhead**: Uses single-copy ringbuffer mechanics; no packet duplication in user memory.
* **Stream Endpoints**: Available via REST `/observability/flows?limit=100` or CLI: `sgctl flows`.
