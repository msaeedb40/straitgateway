# eBPF Map Architecture & Memory Layouts

StraitGateway maintains 13 core pinned BPF maps located under `/sys/fs/bpf/straitgateway/`. These maps form the shared state boundary between userspace daemons (`straitd`, `sg-agent`, `sgctl`) and in-kernel programs (TCX, XDP, cgroup socket hooks).

---

## 1. Pinned Maps Catalog

| Map Name | BPF Type | Key Type | Value Type | Max Entries | Purpose |
|---|---|---|---|---|---|
| `sg_routes` | `BPF_MAP_TYPE_LPM_TRIE` | `struct lpm_key_v4` / `v6` | `struct route_value` | 65,536 | Longest Prefix Match L3 forwarding table |
| `sg_nexthops` | `BPF_MAP_TYPE_HASH` | `__u32` (nexthop_id) | `struct nexthop_value` | 16,384 | Gateway next-hops, egress ifindex & MACs |
| `sg_neighbors` | `BPF_MAP_TYPE_LRU_HASH` | `__u32` (ipv4_addr) | `struct neighbor_value` | 32,768 | L3 IP to L2 MAC resolution cache |
| `sg_interfaces` | `BPF_MAP_TYPE_ARRAY` | `__u32` (ifindex) | `struct iface_value` | 1,024 | Interface properties (MTU, VRF, mode) |
| `sg_services` | `BPF_MAP_TYPE_HASH` | `struct service_key` | `struct service_value` | 32,768 | Kubernetes Service VIP to backend group mapping |
| `sg_endpoints` | `BPF_MAP_TYPE_HASH` | `__u32` (backend_id) | `struct endpoint_value` | 131,072 | Physical pod IP & port backend pool |
| `sg_conntrack` | `BPF_MAP_TYPE_LRU_HASH`| `struct ct_key` | `struct ct_value` | 262,144 | 5-tuple bidirectional connection tracking |
| `sg_nat` | `BPF_MAP_TYPE_HASH` | `struct nat_key` | `struct nat_value` | 65,536 | SNAT / masquerade reverse mapping |
| `sg_policies` | `BPF_MAP_TYPE_HASH` | `struct policy_key` | `struct policy_value` | 32,768 | Identity-to-Identity L3/L4 ACL rules |
| `sg_identities` | `BPF_MAP_TYPE_HASH` | `__u32` (ip_addr) | `__u32` (identity_id) | 65,536 | Pod IP to 32-bit cryptographic identity |
| `sg_maglev` | `BPF_MAP_TYPE_ARRAY` | `__u32` (slot) | `__u32` (backend_id) | 65,537 | Maglev consistent hashing permutation lookup |
| `sg_counters` | `BPF_MAP_TYPE_PERCPU_ARRAY`| `__u32` (index) | `__u64` (count) | 64 | Per-CPU telemetry, packet & byte counters |
| `sg_events` | `BPF_MAP_TYPE_RINGBUF` | N/A | `struct flow_event` | 16 MB | Lockless kernel-to-userspace flow streaming |

---

## 2. Key Data Structures

### 2.1. Routing & Next-Hop Structures (`bpf/include/maps.h`)

```c
struct lpm_key_v4 {
    __u32 prefixlen;     // Prefix length (0 - 32)
    __u32 vrf_id;        // Routing domain (VRF)
    __u32 ip;            // IPv4 address in network byte order
};

struct route_value {
    __u32 nexthop_id;    // Index into sg_nexthops
    __u32 precedence;    // Route administrative distance
    __u32 metric;        // Metric cost
    __u32 flags;         // ROUTE_LOCAL, ROUTE_TUNNEL, etc.
};

struct nexthop_value {
    __u32 ifindex;       // Target network interface index
    __u8  dst_mac[6];    // Next-hop L2 MAC address
    __u8  src_mac[6];    // Egress interface MAC address
    __u32 gateway_ip;    // Next-hop IP (if non-local)
};
```

### 2.2. Service & Maglev Structures

```c
struct service_key {
    __u32 vip;           // Service virtual IP (IPv4)
    __u16 port;          // Service port
    __u8  proto;         // IPPROTO_TCP or IPPROTO_UDP
    __u8  pad;
};

struct service_value {
    __u32 maglev_offset; // Starting index in sg_maglev table
    __u32 backend_count; // Number of healthy backends
    __u16 session_affinity_timeout;
    __u16 flags;
};
```

### 2.3. Security & Identity Structures

```c
struct policy_key {
    __u32 src_identity;  // Source 32-bit identity ID
    __u32 dst_identity;  // Destination 32-bit identity ID
    __u16 dst_port;      // Destination port (0 = wildcard)
    __u8  protocol;      // Protocol (6 = TCP, 17 = UDP, 0 = ANY)
    __u8  pad;
};

struct policy_value {
    __u8  action;        // 1 = ALLOW, 2 = DENY, 3 = LOG_AND_ALLOW
    __u8  pad[3];
    __u32 rule_id;       // SecurityPolicy rule correlation ID
};
```

---

## 3. Concurrency & Memory Model

1. **Per-CPU Array Scaling**: Counters (`sg_counters`) use `BPF_MAP_TYPE_PERCPU_ARRAY` to eliminate cache-line bouncing and CPU lock contention.
2. **Lockless Ring Buffer**: Telemetry flows use `BPF_MAP_TYPE_RINGBUF` with memory-mapped circular queues, ensuring zero drop rates up to millions of events per second.
3. **Atomic Map Updates**: Control plane programs updates with `bpf_map_update_elem` using `BPF_ANY` for non-blocking in-place swaps.
