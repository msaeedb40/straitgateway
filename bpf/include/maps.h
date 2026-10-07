#ifndef __STRAITGATEWAY_MAPS_H__
#define __STRAITGATEWAY_MAPS_H__

#include "common.h"

#define BPF_MAP_DEF(name, map_type, key_type, val_type, max_ents) \
struct { \
    __uint(type, map_type); \
    __type(key, key_type); \
    __type(value, val_type); \
    __uint(max_entries, max_ents); \
} name SEC(".maps")

// Protocol definitions
#define SG_PROTO_TCP 6
#define SG_PROTO_UDP 17

// Endpoint state
#define SG_EP_ACTIVE 1
#define SG_EP_DRAINING 2
#define SG_EP_OFFLINE 3

// L3 Route Key (for LPM trie)
struct lpm_v4_key {
    __u32 prefixlen;
    __u32 addr;
};

// NextHop descriptor
struct nexthop_info {
    __u32 ifindex;
    __u32 ip;
    __u8 mac[6];
    __u16 flags;
    __u32 vrf_id;
};

// Service VIP descriptor
struct service_key {
    __u32 ip;
    __u16 port;
    __u8 proto;
    __u8 pad;
};

// Service Backend Endpoint
struct service_endpoint {
    __u32 ip;
    __u16 port;
    __u8 weight;
    __u8 state;
};

// Flow / Conntrack Key
struct flow_key {
    __u32 src_ip;
    __u32 dst_ip;
    __u16 src_port;
    __u16 dst_port;
    __u8 proto;
    __u8 pad[3];
};

// Conntrack State
struct flow_entry {
    __u64 bytes;
    __u64 packets;
    __u64 last_seen;
    __u32 flags;
    __u32 rev_src_ip;
    __u32 rev_dst_ip;
    __u16 rev_src_port;
    __u16 rev_dst_port;
};

// Security Policy Key & Action
struct policy_key {
    __u32 src_id;
    __u32 dst_id;
    __u16 dst_port;
    __u8 proto;
    __u8 pad;
};

enum policy_action {
    POLICY_DROP = 0,
    POLICY_ALLOW = 1,
    POLICY_AUDIT = 2,
};

// VRF Domain Descriptor
struct vrf_config {
    __u32 id;
    __u32 table_id;
    __u32 flags;
};

// NAT Translation Entry
struct nat_entry {
    __u32 orig_ip;
    __u16 orig_port;
    __u32 trans_ip;
    __u16 trans_port;
    __u8 proto;
    __u8 pad[3];
};

// WireGuard / Transit Tunnel Descriptor
struct tunnel_info {
    __u32 tunnel_id;
    __u32 ifindex;
    __u32 endpoint_ip;
    __u16 endpoint_port;
    __u16 flags;
};

// L2 Neighbor Cache Entry (IP to MAC)
struct neighbor_entry {
    __u32 ip;
    __u8 mac[6];
    __u16 state;
};

// Telemetry Event (delivered via BPF ring buffer)
struct telemetry_event {
    __u64 timestamp;
    __u32 src_ip;
    __u32 dst_ip;
    __u16 src_port;
    __u16 dst_port;
    __u8 proto;
    __u32 src_id;
    __u32 dst_id;
    __u8 action;
    __u8 pad[2];
    __u64 bytes;
    __u64 packets;
};

#endif /* __STRAITGATEWAY_MAPS_H__ */
