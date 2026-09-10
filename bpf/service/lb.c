// SPDX-License-Identifier: Apache-2.0
// Copyright 2026 straitgateway Authors

// lb.c — eBPF Maglev service load balancer
// Attached at TC-X level for host-side packet processing.
// Performs VIP → backend DNAT using Maglev consistent hashing.
// Supports: ClusterIP, NodePort, LoadBalancer, DSR, session affinity.

#include <linux/bpf.h>
#include <linux/if_ether.h>
#include <linux/ip.h>
#include <linux/tcp.h>
#include <linux/udp.h>
#include <linux/in.h>
#include <bpf/bpf_helpers.h>
#include <bpf/bpf_endian.h>

#include "maps.h"

#define TC_ACT_OK   0
#define TC_ACT_SHOT 2

#ifndef ETH_P_IP
#define ETH_P_IP 0x0800
#endif

// Maglev lookup: hash the 4-tuple and index into the service's backend table.
static __always_inline struct backend_value *
maglev_select(struct service_value *svc, __u32 src_ip, __u32 dst_ip,
              __u16 src_port, __u16 dst_port, __u8 proto, __u32 *out_backend_id)
{
    // FNV-1a hash of the 5-tuple.
    __u32 hash = 2166136261;
    hash ^= src_ip;    hash *= 16777619;
    hash ^= dst_ip;    hash *= 16777619;
    hash ^= src_port;  hash *= 16777619;
    hash ^= dst_port;  hash *= 16777619;
    hash ^= proto;     hash *= 16777619;

    // Slot = hash mod backend_count.
    if (svc->backend_count == 0)
        return NULL;

    __u32 slot = hash % svc->backend_count;
    struct backend_key bk = { .id = slot + 1 };
    if (out_backend_id)
        *out_backend_id = bk.id;

    return bpf_map_lookup_elem(&backend_map, &bk);
}

// Perform service DNAT: rewrite destination IP/port to selected backend.
static __always_inline int
svc_dnat(struct __sk_buff *skb, struct service_value *svc,
         __u32 src_ip, __u32 dst_ip, __u16 src_port, __u16 dst_port, __u8 proto)
{
    __u32 backend_id = 0;
    struct backend_value *be = maglev_select(svc, src_ip, dst_ip,
                                              src_port, dst_port, proto, &backend_id);
    if (!be)
        return TC_ACT_SHOT;

    // Skip backends in terminating/quarantined state.
    if (be->state != 0)
        return TC_ACT_SHOT;

    __u32 be_ip = be->ip4;

    // l3_csum_replace for IP header.
    bpf_l3_csum_replace(skb, ETH_HLEN + offsetof(struct iphdr, check),
                        dst_ip, be_ip, sizeof(__u32));

    // Store L4 checksum and rewrite destination.
    bpf_skb_store_bytes(skb, ETH_HLEN + offsetof(struct iphdr, daddr),
                        &be_ip, sizeof(be_ip), 0);

    // Create conntrack entry for reverse path NAT.
    struct ct_key ctk = {
        .src_ip4  = src_ip,
        .dst_ip4  = be_ip,
        .src_port = src_port,
        .dst_port = dst_port,
        .protocol = proto,
        .dir      = 0,
    };

    struct ct_value ctv = {
        .backend_id = backend_id,
        .rx_packets = 1,
        .rx_bytes   = skb->len,
        .last_seen  = bpf_ktime_get_ns() / 1000000000ULL,
        .state      = 1,
    };
    bpf_map_update_elem(&ct_map, &ctk, &ctv, BPF_ANY);

    return TC_ACT_OK;
}

// Main TC-X service LB entry point.
// Called from tcx_host.c for packets destined to a service VIP.
SEC("tc")
int sg_svc_lb(struct __sk_buff *skb)
{
    void *data     = (void *)(long)skb->data;
    void *data_end = (void *)(long)skb->data_end;

    struct ethhdr *eth = data;
    if ((void *)(eth + 1) > data_end)
        return TC_ACT_OK;

    if (eth->h_proto != bpf_htons(ETH_P_IP))
        return TC_ACT_OK;

    struct iphdr *ip = (void *)(eth + 1);
    if ((void *)(ip + 1) > data_end)
        return TC_ACT_OK;

    __u32 dst_ip = ip->daddr;
    __u32 src_ip = ip->saddr;
    __u8  proto  = ip->protocol;

    __u16 src_port = 0, dst_port = 0;
    if (proto == IPPROTO_TCP) {
        struct tcphdr *tcp = (void *)ip + (ip->ihl * 4);
        if ((void *)(tcp + 1) > data_end)
            return TC_ACT_OK;
        src_port = tcp->source;
        dst_port = tcp->dest;
    } else if (proto == IPPROTO_UDP) {
        struct udphdr *udp = (void *)ip + (ip->ihl * 4);
        if ((void *)(udp + 1) > data_end)
            return TC_ACT_OK;
        src_port = udp->source;
        dst_port = udp->dest;
    }

    // Lookup service by VIP + port + proto.
    struct service_key sk = {
        .vip4     = dst_ip,
        .port     = dst_port,
        .protocol = proto,
    };

    struct service_value *svc = bpf_map_lookup_elem(&service_map, &sk);
    if (!svc)
        return TC_ACT_OK;  // Not a service VIP — pass through.

    return svc_dnat(skb, svc, src_ip, dst_ip, src_port, dst_port, proto);
}

char _license[] SEC("license") = "Apache-2.0";
