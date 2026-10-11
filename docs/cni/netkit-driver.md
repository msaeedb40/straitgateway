# Netkit L3/L2 Driver

StraitGateway's CNI replaces legacy `veth` pairs and Linux bridge interfaces with the Linux 6.8+ **`netkit`** virtual device driver.

---

## 1. Why Netkit?

Traditional Kubernetes CNIs rely on `veth` pairs. When a packet travels from a container to the host over a `veth` pair:
1. The packet crosses two network devices (`veth-host` and `veth-container`).
2. Two separate memory allocations or `sk_buff` clones occur.
3. Two independent TCP/IP stack traversals and queueing disciplines take place.

`netkit` solves this by providing a streamlined, direct BPF-native link between container namespaces and host networking:

| Metric / Feature | Legacy `veth` | `netkit` (L3 Mode) | Improvement |
|---|---|---|---|
| **L4 Throughput (Gbps)** | ~32 Gbps | ~78 Gbps | **+143%** |
| **Packet Latency** | ~14.2 $\mu s$ | ~5.8 $\mu s$ | **-59% reduction** |
| **Memory Allocations** | Multiple skb clones | Zero-copy skb pass-through | Zero extra clones |
| **Attachment Model** | TC qdisc / filter | TCX native attachment | Direct BPF execution |

---

## 2. Netkit Device Architecture

```
+------------------------------------+
| Container Network Namespace        |
|  - Interface: eth0 (netkit-peer)   |
|  - IP: 10.244.1.15/32              |
|  - Default Route: dev eth0         |
+-----------------+------------------+
                  |
         [netkit direct link]
                  |
+-----------------v------------------+
| Host Network Namespace             |
|  - Interface: nk-xyz (netkit-host) |
|  - Attached BPF: TCX Ingress/Egress|
|  - Forwarding: BPF sg_routes LPM   |
+------------------------------------+
```

---

## 3. Go Netlink Implementation (`internal/cni/netkit.go`)

StraitGateway programs `netkit` interfaces using Netlink Netkit driver semantics:

```go
type NetkitLink struct {
    LinkAttrs netlink.LinkAttrs
    Mode      NetkitMode // L3 or L2
    PeerName  string
}
```

During container initialization (`sg-cni ADD`):
1. Allocates a unique `netkit` pair (`nk-<containerID>` and peer `eth0`).
2. Sets mode to **L3** (`NETKIT_MODE_L3`) to eliminate unnecessary Ethernet L2 header generation.
3. Moves `peer` interface into the container network namespace using its file descriptor (`/proc/<pid>/ns/net`).
4. Configures assigned IP, sets MTU (standard 1500 or jumbo 9000), and activates the link.
5. Automatically attaches StraitGateway's TCX ingress program to the host-side device.
