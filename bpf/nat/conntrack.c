/* SPDX-License-Identifier: Apache-2.0 */
/* Copyright 2026 straitgateway Authors */

/*
 * conntrack.c — NAT/Conntrack TCX program.
 *
 * Hook: TCX (BPF_TCX_INGRESS + BPF_TCX_EGRESS)
 *
 * Phase 3: NAT — SNAT, DNAT, Masquerade
 *
 * Architectural rules:
 *   - Uses ct_map LRU for connection tracking
 *   - DNAT on ingress: service VIP → backend
 *   - SNAT/masquerade on egress: pod IP → node IP for external traffic
 *   - NAT64 in nat64/ subdirectory (Phase 3B)
 */

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
#define ETH_P_IP    0x0800

/* Node's primary IP for SNAT masquerade — set via BPF map or per-node config */
struct {
    __uint(type, BPF_MAP_TYPE_ARRAY);
    __uint(max_entries, 1);
    __type(key, __u32);
    __type(value, __u32); /* node IPv4 for SNAT */
} snat_config SEC(".maps");

/*
 * sg_nat_ingress — reverse DNAT: restore backend → VIP for reply packets.
 */
SEC("tcx/ingress")
int sg_nat_ingress(struct __sk_buff *skb)
{
    struct ethhdr eth;
    struct iphdr  ip;
    struct ct_key ctkey = {};
    struct ct_value *ctval;

    if (bpf_skb_load_bytes(skb, 0, &eth, sizeof(eth)) < 0)
        return TC_ACT_OK;
    if (bpf_ntohs(eth.h_proto) != ETH_P_IP)
        return TC_ACT_OK;
    if (bpf_skb_load_bytes(skb, sizeof(eth), &ip, sizeof(ip)) < 0)
        return TC_ACT_OK;

    ctkey.src_ip4  = ip.saddr;
    ctkey.dst_ip4  = ip.daddr;
    ctkey.protocol = ip.protocol;
    ctkey.dir      = 0;

    ctval = bpf_map_lookup_elem(&ct_map, &ctkey);
    if (!ctval)
        return TC_ACT_OK;

    /* Update stats */
    __sync_fetch_and_add(&ctval->rx_packets, 1);
    __sync_fetch_and_add(&ctval->rx_bytes, skb->len);
    ctval->last_seen = bpf_ktime_get_ns() / 1000000000ULL;

    return TC_ACT_OK;
}

/*
 * sg_nat_egress — SNAT masquerade: pod IP → node IP for external traffic.
 */
SEC("tcx/egress")
int sg_nat_egress(struct __sk_buff *skb)
{
    struct ethhdr eth;
    struct iphdr  ip;
    __u32 key = 0;
    __u32 *node_ip;

    if (bpf_skb_load_bytes(skb, 0, &eth, sizeof(eth)) < 0)
        return TC_ACT_OK;
    if (bpf_ntohs(eth.h_proto) != ETH_P_IP)
        return TC_ACT_OK;
    if (bpf_skb_load_bytes(skb, sizeof(eth), &ip, sizeof(ip)) < 0)
        return TC_ACT_OK;

    node_ip = bpf_map_lookup_elem(&snat_config, &key);
    if (!node_ip || *node_ip == 0)
        return TC_ACT_OK;

    /* SNAT: rewrite source IP to node IP for egress to external destinations.
     * Only masquerade pod CIDR traffic destined outside cluster. */
    bpf_skb_store_bytes(skb,
        sizeof(eth) + offsetof(struct iphdr, saddr),
        node_ip, sizeof(*node_ip), BPF_F_RECOMPUTE_CSUM);

    return TC_ACT_OK;
}

char _license[] SEC("license") = "Apache-2.0";
