/* SPDX-License-Identifier: GPL-2.0-only */
#ifndef __SG_COMMON_H__
#define __SG_COMMON_H__

#include "vmlinux.h"

#ifndef TC_ACT_UNSPEC
#define TC_ACT_UNSPEC     (-1)
#endif
#ifndef TC_ACT_OK
#define TC_ACT_OK        0
#endif
#ifndef TC_ACT_RECLASSIFY
#define TC_ACT_RECLASSIFY 1
#endif
#ifndef TC_ACT_SHOT
#define TC_ACT_SHOT       2
#endif
#ifndef TC_ACT_PIPE
#define TC_ACT_PIPE       3
#endif

#ifndef ETH_P_IP
#define ETH_P_IP   0x0800
#endif
#ifndef ETH_P_IPV6
#define ETH_P_IPV6 0x86DD
#endif

#ifndef IPPROTO_ICMP
#define IPPROTO_ICMP 1
#endif
#ifndef IPPROTO_TCP
#define IPPROTO_TCP  6
#endif
#ifndef IPPROTO_UDP
#define IPPROTO_UDP  17
#endif
#ifndef IPPROTO_IPV6
#define IPPROTO_IPV6 41
#endif
#ifndef IPPROTO_ICMPV6
#define IPPROTO_ICMPV6 58
#endif

#ifndef __always_inline
#define __always_inline inline __attribute__((always_inline))
#endif

/* Packet boundary verification helper */
static __always_inline bool check_pkt_bounds(void *ptr, size_t len, void *data_end) {
    return (ptr + len) <= data_end;
}

/* ========================================================================= */
/* Generic IP Address Structure (Dual-Stack)                                 */
/* ========================================================================= */
#define SG_FAMILY_IPV4 2
#define SG_FAMILY_IPV6 10

struct sg_ip_addr {
	__u8  family;
	__u8  pad[3];
	__u8  raw[16];
};

/* ========================================================================= */
/* Service Map Structures                                                    */
/* ========================================================================= */
#define SG_SVC_FLAG_ACTIVE        (1 << 0)
#define SG_SVC_FLAG_SESSION_AFF   (1 << 1)
#define SG_SVC_FLAG_NODEPORT      (1 << 2)
#define SG_SVC_FLAG_EXTERNAL_IP   (1 << 3)
#define SG_SVC_FLAG_LOADBALANCER  (1 << 4)

/* IPv4 Service Key (8 bytes) */
struct sg_service_key {
	__u32 addr;       /* VIP IPv4 address (network byte order) */
	__u16 port;       /* Service port (network byte order) */
	__u8  proto;      /* IP protocol: IPPROTO_TCP, IPPROTO_UDP */
	__u8  pad;
};

/* Service Value (8 bytes) — shared for IPv4 and IPv6 */
struct sg_service_val {
	__u32 backend_id;    /* Service / Backend group ID */
	__u16 backend_count; /* Total number of backend endpoints */
	__u16 flags;         /* SG_SVC_FLAG_* */
};

/* IPv6 Service Key (20 bytes) */
struct sg_service_v6_key {
	__u8  addr[16];   /* VIP IPv6 address (network byte order) */
	__u16 port;       /* Service port (network byte order) */
	__u8  proto;      /* IP protocol: IPPROTO_TCP, IPPROTO_UDP */
	__u8  pad;
};

/* Backend Key (8 bytes) — shared for IPv4 and IPv6 */
struct sg_backend_key {
	__u32 id;    /* Backend (service) ID */
	__u32 slot;  /* Backend slot index (0-based) */
};

#define SG_BACKEND_FLAG_ACTIVE (1 << 0)
#define SG_BACKEND_FLAG_DRAIN  (1 << 1)

/* IPv4 Backend Value (8 bytes) */
struct sg_backend_val {
	__u32 addr;    /* Backend IPv4 address (network byte order) */
	__u16 port;    /* Backend port (network byte order) */
	__u8  flags;   /* SG_BACKEND_FLAG_* */
	__u8  pad;
};

/* IPv6 Backend Value (20 bytes) */
struct sg_backend_v6_val {
	__u8  addr[16]; /* Backend IPv6 address (network byte order) */
	__u16 port;     /* Backend port (network byte order) */
	__u8  flags;    /* SG_BACKEND_FLAG_* */
	__u8  pad;
};

/* ========================================================================= */
/* Policy Map Structures                                                     */
/* ========================================================================= */
#define SG_POLICY_DENY  0
#define SG_POLICY_ALLOW 1

/* Policy Key (12 bytes) */
struct sg_policy_key {
	__u32 src_id;   /* Source identity (eBPF-assigned) */
	__u32 dst_id;   /* Destination identity */
	__u16 dst_port; /* Destination L4 port (network byte order) */
	__u8  proto;    /* L4 Protocol */
	__u8  dir;      /* 0 = ingress, 1 = egress */
};

/* Policy Value (4 bytes) */
struct sg_policy_val {
	__u8 verdict; /* SG_POLICY_ALLOW or SG_POLICY_DENY */
	__u8 pad[3];
};

/* ========================================================================= */
/* Route Map Structures (Linux Kernel LPM Trie Layout)                       */
/* ========================================================================= */
#define SG_ROUTE_FLAG_BLACKHOLE (1 << 0)
#define SG_ROUTE_FLAG_LOCAL     (1 << 1)

/* IPv4 LPM Route Key (8 bytes) — prefix_len MUST be first for BPF_MAP_TYPE_LPM_TRIE */
struct sg_route_key {
	__u32 prefix_len; /* Prefix length in bits (0–32) */
	__u32 prefix;     /* IPv4 network address (network byte order) */
};

/* IPv4 Route Value (12 bytes) */
struct sg_route_val {
	__u32 nexthop;    /* Next-hop IPv4 address (network byte order) */
	__u32 ifindex;    /* Output interface index */
	__u8  flags;      /* SG_ROUTE_FLAG_* */
	__u8  pad[3];
};

/* IPv6 LPM Route Key (20 bytes) */
struct sg_route_v6_key {
	__u32 prefix_len; /* Prefix length in bits (0–128) */
	__u8  prefix[16]; /* IPv6 network prefix (network byte order) */
};

/* IPv6 Route Value (24 bytes) */
struct sg_route_v6_val {
	__u8  nexthop[16]; /* Next-hop IPv6 address (network byte order) */
	__u32 ifindex;     /* Output interface index */
	__u8  flags;       /* SG_ROUTE_FLAG_* */
	__u8  pad[3];
};

/* ========================================================================= */
/* Peer Map Structures                                                       */
/* ========================================================================= */
#define SG_PEER_TYPE_TRANSIT 1
#define SG_PEER_TYPE_MESH    2

/* Peer Key (4 bytes) */
struct sg_peer_key {
	__u32 peer_id;    /* Peer identity identifier */
};

/* IPv4 Peer Value (8 bytes) */
struct sg_peer_val {
	__u32 addr;       /* Peer IPv4 endpoint address (network byte order) */
	__u16 port;       /* Peer port (e.g., for VXLAN or WireGuard) */
	__u8  type;       /* SG_PEER_TYPE_* */
	__u8  flags;
};

/* IPv6 Peer Value (20 bytes) */
struct sg_peer_v6_val {
	__u8  addr[16];   /* Peer IPv6 endpoint address (network byte order) */
	__u16 port;       /* Peer port (network byte order) */
	__u8  type;       /* SG_PEER_TYPE_* */
	__u8  flags;
};

/* ========================================================================= */
/* Flow Map Structures                                                       */
/* ========================================================================= */
/* IPv4 Flow Key: 5-tuple (16 bytes) */
struct sg_flow_key {
	__u32 src_addr;   /* Source IPv4 (network byte order) */
	__u32 dst_addr;   /* Destination IPv4 (network byte order) */
	__u16 src_port;   /* Source port (network byte order) */
	__u16 dst_port;   /* Destination port (network byte order) */
	__u8  proto;      /* L4 Protocol */
	__u8  pad[3];
};

/* IPv6 Flow Key: 5-tuple (40 bytes) */
struct sg_flow_v6_key {
	__u8  src_addr[16]; /* Source IPv6 (network byte order) */
	__u8  dst_addr[16]; /* Destination IPv6 (network byte order) */
	__u16 src_port;     /* Source port (network byte order) */
	__u16 dst_port;     /* Destination port (network byte order) */
	__u8  proto;        /* L4 Protocol */
	__u8  pad[3];
};

/* Flow Value: packet/byte counters and metadata (40 bytes, 8-byte aligned) */
struct sg_flow_val {
	__u64 packets;      /* Cumulative packets */
	__u64 bytes;        /* Cumulative bytes */
	__u64 last_seen_ns; /* ktime_get_ns() timestamp */
	__u32 src_id;       /* Source identity */
	__u32 dst_id;       /* Destination identity */
	__u8  state;        /* Connection state */
	__u8  pad[7];       /* 8-byte struct alignment padding */
};

#endif /* __SG_COMMON_H__ */
