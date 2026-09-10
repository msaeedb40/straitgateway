/* SPDX-License-Identifier: Apache-2.0 */
/* Copyright 2026 straitgateway Authors */

#ifndef __STRAITGATEWAY_MAPS_H__
#define __STRAITGATEWAY_MAPS_H__

#include <linux/bpf.h>
#include <bpf/bpf_helpers.h>

/* ============================================================
 * BPF Map Layout Version
 * All map schemas are versioned. Increment on any ABI change.
 * ============================================================ */
#define SG_MAP_VERSION 1

/* ============================================================
 * Constants
 * ============================================================ */
#define SG_MAX_ENDPOINTS    65536
#define SG_MAX_SERVICES     16384
#define SG_MAX_BACKENDS     65536
#define SG_MAX_CT_ENTRIES   524288  /* conntrack LRU */
#define SG_MAX_POLICIES     16384
#define SG_MAX_IDENTITIES   65536
#define SG_MAX_NODES        1024
#define SG_MAGLEV_SLOTS     128     /* Maglev hash table size per service */

#define SG_BPFFS_PIN_DIR    "/sys/fs/bpf/straitgateway"

/* ============================================================
 * Key / Value Structures
 * ============================================================ */

/* Endpoint key: interface index */
struct endpoint_key {
    __u32 ifindex;      /* host-side NetKit ifindex */
    __u8  pad[4];
} __attribute__((packed));

/* Endpoint value: full endpoint state */
struct endpoint_value {
    __u32 identity;     /* 32-bit security identity */
    __u32 node_ip4;     /* node IPv4 address (big-endian) */
    __u32 pod_ip4;      /* pod IPv4 address (big-endian) */
    __u8  pod_ip6[16];  /* pod IPv6 address */
    __u8  mac[6];       /* host-side MAC */
    __u16 lxc_id;       /* local endpoint id */
    __u8  pad[4];
} __attribute__((packed));

/* Service key: VIP + port + protocol */
struct service_key {
    __u32 vip4;         /* service VIP IPv4 (big-endian) */
    __u16 port;         /* service port */
    __u8  protocol;     /* IPPROTO_TCP / IPPROTO_UDP */
    __u8  pad;
} __attribute__((packed));

/* Service value: backend count + Maglev table index */
struct service_value {
    __u32 backend_count;    /* number of active backends */
    __u32 maglev_idx;       /* index into maglev_table map */
    __u8  algorithm;        /* LB algorithm: 0=maglev, 1=rr, 2=lc */
    __u8  flags;            /* DSR=0x1, session_affinity=0x2 */
    __u8  pad[2];
} __attribute__((packed));

/* Backend key: backend ID */
struct backend_key {
    __u32 id;
} __attribute__((packed));

/* Backend value: real server address */
struct backend_value {
    __u32 ip4;          /* backend IPv4 */
    __u16 port;         /* backend port */
    __u8  protocol;
    __u8  state;        /* 0=active, 1=draining, 2=down */
    __u32 weight;
    __u8  pad[4];
} __attribute__((packed));

/* Conntrack key: 5-tuple */
struct ct_key {
    __u32 src_ip4;
    __u32 dst_ip4;
    __u16 src_port;
    __u16 dst_port;
    __u8  protocol;
    __u8  dir;          /* 0=ingress, 1=egress */
    __u8  pad[2];
} __attribute__((packed));

/* Conntrack value: connection state */
struct ct_value {
    __u32 backend_id;   /* selected backend (for DNAT) */
    __u32 rx_packets;
    __u64 rx_bytes;
    __u32 tx_packets;
    __u64 tx_bytes;
    __u32 last_seen;    /* timestamp seconds */
    __u8  state;        /* TCP state */
    __u8  pad[3];
} __attribute__((packed));

/* Identity key: numeric identity */
struct identity_key {
    __u32 identity;
} __attribute__((packed));

/* Identity value: labels hash + segment */
struct identity_value {
    __u64 labels_hash;
    __u32 segment_id;
    __u8  pad[4];
} __attribute__((packed));

/* Policy key: src identity + dst identity + port + protocol */
struct policy_key {
    __u32 src_identity;
    __u32 dst_identity;
    __u16 dst_port;
    __u8  protocol;
    __u8  pad;
} __attribute__((packed));

/* Policy value: action (0=allow, 1=deny, 2=reject) + priority */
struct policy_value {
    __u8  action;       /* 0=allow, 1=deny, 2=reject */
    __u8  pad[3];
    __u32 priority;     /* lower = higher priority */
} __attribute__((packed));

/* Node key: node IP */
struct node_key {
    __u32 ip4;
} __attribute__((packed));

/* Node value: wireguard public key + tunnel endpoint */
struct node_value {
    __u8  wg_pubkey[32];    /* WireGuard public key */
    __u32 tunnel_ip4;       /* tunnel endpoint IP */
    __u32 segment_id;
    __u8  pad[4];
} __attribute__((packed));

/* Flow event: per-packet observability ring buffer record */
struct flow_event {
    __u32 src_ip4;
    __u32 dst_ip4;
    __u16 src_port;
    __u16 dst_port;
    __u32 src_identity;
    __u32 dst_identity;
    __u32 policy_id;
    __u64 bytes;
    __u64 timestamp_ns;
    __u8  protocol;
    __u8  direction;    /* 0=ingress, 1=egress */
    __u8  action;       /* 0=allow, 1=deny, 2=reject */
    __u8  drop_reason;
} __attribute__((packed));

/* ============================================================
 * Map Declarations (pinned to /sys/fs/bpf/straitgateway/)
 * ============================================================ */

/* endpoint_map: ifindex → endpoint state */
struct {
    __uint(type, BPF_MAP_TYPE_HASH);
    __uint(max_entries, SG_MAX_ENDPOINTS);
    __type(key, struct endpoint_key);
    __type(value, struct endpoint_value);
    __uint(pinning, LIBBPF_PIN_BY_NAME);
} endpoint_map SEC(".maps");

/* service_map: VIP:port:proto → service state */
struct {
    __uint(type, BPF_MAP_TYPE_HASH);
    __uint(max_entries, SG_MAX_SERVICES);
    __type(key, struct service_key);
    __type(value, struct service_value);
    __uint(pinning, LIBBPF_PIN_BY_NAME);
} service_map SEC(".maps");

/* backend_map: backend_id → backend address */
struct {
    __uint(type, BPF_MAP_TYPE_HASH);
    __uint(max_entries, SG_MAX_BACKENDS);
    __type(key, struct backend_key);
    __type(value, struct backend_value);
    __uint(pinning, LIBBPF_PIN_BY_NAME);
} backend_map SEC(".maps");

/* maglev_outer: service maglev_idx → inner Maglev LRU map */
struct {
    __uint(type, BPF_MAP_TYPE_ARRAY_OF_MAPS);
    __uint(max_entries, SG_MAX_SERVICES);
    __type(key, __u32);
    __uint(pinning, LIBBPF_PIN_BY_NAME);
} maglev_outer SEC(".maps");

/* ct_map: conntrack LRU for NAT */
struct {
    __uint(type, BPF_MAP_TYPE_LRU_HASH);
    __uint(max_entries, SG_MAX_CT_ENTRIES);
    __type(key, struct ct_key);
    __type(value, struct ct_value);
    __uint(pinning, LIBBPF_PIN_BY_NAME);
} ct_map SEC(".maps");

/* policy_map: policy rules → action */
struct {
    __uint(type, BPF_MAP_TYPE_HASH);
    __uint(max_entries, SG_MAX_POLICIES);
    __type(key, struct policy_key);
    __type(value, struct policy_value);
    __uint(pinning, LIBBPF_PIN_BY_NAME);
} policy_map SEC(".maps");

/* identity_map: identity → labels + segment */
struct {
    __uint(type, BPF_MAP_TYPE_HASH);
    __uint(max_entries, SG_MAX_IDENTITIES);
    __type(key, struct identity_key);
    __type(value, struct identity_value);
    __uint(pinning, LIBBPF_PIN_BY_NAME);
} identity_map SEC(".maps");

/* node_map: node IP → WireGuard + tunnel info */
struct {
    __uint(type, BPF_MAP_TYPE_HASH);
    __uint(max_entries, SG_MAX_NODES);
    __type(key, struct node_key);
    __type(value, struct node_value);
    __uint(pinning, LIBBPF_PIN_BY_NAME);
} node_map SEC(".maps");

/* flow_events: ring buffer for per-packet observability */
struct {
    __uint(type, BPF_MAP_TYPE_RINGBUF);
    __uint(max_entries, 1 << 24); /* 16 MB */
    __uint(pinning, LIBBPF_PIN_BY_NAME);
} flow_events SEC(".maps");

#endif /* __STRAITGATEWAY_MAPS_H__ */
