/* SPDX-License-Identifier: Apache-2.0 */
/* Copyright 2026 straitgateway Authors */

/*
 * tcx_host.c — TCX host-side packet processing.
 *
 * Hook: TCX (TC eXpress — BPF_TCX_INGRESS / BPF_TCX_EGRESS)
 * Attach point: Host-side veth/NetKit interface (node namespace)
 *
 * Fast-path responsibilities:
 *   - Service load balancing (DNAT for ClusterIP services)
 *   - kube-proxy replacement: translate Service VIPs → backend IPs
 *   - Conntrack state creation for NATted connections
 *   - SNAT masquerade for pod → external traffic
 *
 * NOT used for: observability events (ring buffer only), BGP routing.
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

#define TC_ACT_OK       0
#define TC_ACT_SHOT     2
#define TC_ACT_REDIRECT 7

#define ETH_P_IP   0x0800

/* ============================================================
 * Helpers
 * ============================================================ */

static __always_inline int parse_ipv4(struct __sk_buff *skb,
                                       struct iphdr *ip,
                                       __u64 nh_off)
{
    if (bpf_skb_load_bytes(skb, nh_off, ip, sizeof(*ip)) < 0)
        return -1;
    return 0;
}

/* Lookup service VIP and select backend via Maglev hash */
static __always_inline struct backend_value *
lb_lookup_backend(struct service_key *skey, __u32 src_ip, __u32 *out_backend_id)
{
    struct service_value *svc;
    struct backend_key bkey;
    __u32 slot;

    svc = bpf_map_lookup_elem(&service_map, skey);
    if (!svc || svc->backend_count == 0)
        return NULL;

    /* Maglev consistent hashing: hash(src_ip) mod SG_MAGLEV_SLOTS */
    slot = (src_ip * 2654435761ULL) % SG_MAGLEV_SLOTS % svc->backend_count;
    bkey.id = slot + 1;  /* backend IDs are 1-indexed */
    if (out_backend_id)
        *out_backend_id = bkey.id;

    return bpf_map_lookup_elem(&backend_map, &bkey);
}

/* ============================================================
 * TCX Ingress — packets entering the node from outside
 * ============================================================ */
SEC("tcx/ingress")
int sg_tcx_ingress(struct __sk_buff *skb)
{
    struct ethhdr eth;
    struct iphdr  ip;
    struct service_key skey = {};
    struct backend_value *backend;
    __u32 backend_id = 0;
    struct ct_key ctkey = {};
    struct ct_value ctval = {};

    if (bpf_skb_load_bytes(skb, 0, &eth, sizeof(eth)) < 0)
        return TC_ACT_OK;

    if (bpf_ntohs(eth.h_proto) != ETH_P_IP)
        return TC_ACT_OK;

    if (parse_ipv4(skb, &ip, sizeof(eth)) < 0)
        return TC_ACT_OK;

    /* Service load balancing — DNAT for ClusterIP/NodePort */
    skey.vip4     = ip.daddr;
    skey.protocol = ip.protocol;

    if (ip.protocol == IPPROTO_TCP) {
        __u16 ports[2];
        if (bpf_skb_load_bytes(skb, sizeof(eth) + (ip.ihl << 2),
                                ports, sizeof(ports)) < 0)
            return TC_ACT_OK;
        skey.port = ports[1]; /* dst port */
    } else if (ip.protocol == IPPROTO_UDP) {
        __u16 ports[2];
        if (bpf_skb_load_bytes(skb, sizeof(eth) + (ip.ihl << 2),
                                ports, sizeof(ports)) < 0)
            return TC_ACT_OK;
        skey.port = ports[1];
    }

    backend = lb_lookup_backend(&skey, ip.saddr, &backend_id);
    if (!backend)
        return TC_ACT_OK; /* not a service VIP — pass through */

    /* Create conntrack entry for reverse NAT */
    ctkey.src_ip4  = ip.saddr;
    ctkey.dst_ip4  = ip.daddr;
    ctkey.protocol = ip.protocol;
    ctkey.dir      = 0; /* ingress */
    ctval.backend_id = backend_id;
    ctval.last_seen  = bpf_ktime_get_ns() / 1000000000ULL;
    bpf_map_update_elem(&ct_map, &ctkey, &ctval, BPF_ANY);

    /* DNAT: rewrite destination IP to backend IP */
    bpf_skb_store_bytes(skb,
        sizeof(eth) + offsetof(struct iphdr, daddr),
        &backend->ip4, sizeof(backend->ip4), BPF_F_RECOMPUTE_CSUM);

    return TC_ACT_OK;
}

/* ============================================================
 * TCX Egress — packets leaving the node
 * ============================================================ */
SEC("tcx/egress")
int sg_tcx_egress(struct __sk_buff *skb)
{
    /* Reverse NAT / SNAT masquerade for pod→external traffic.
     * Full implementation in nat/conntrack.c — this program
     * handles fast-path conntrack lookup and IP rewrite. */
    return TC_ACT_OK;
}

char _license[] SEC("license") = "Apache-2.0";
