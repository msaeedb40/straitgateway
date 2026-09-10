/* SPDX-License-Identifier: Apache-2.0 */
/* Copyright 2026 straitgateway Authors */

/*
 * xdp_edge.c — XDP earliest ingress packet processing.
 *
 * Hook: XDP (native or generic)
 * Attach point: Physical NIC / uplink interface
 *
 * Responsibilities (XDP is the FIRST hook — before sk_buff allocation):
 *   - DDoS/rate-limit filtering at line rate
 *   - NodePort acceleration: translate NodePort → backend without sk_buff
 *   - Drop obviously malformed packets early
 *   - Pass everything else to the kernel stack
 *
 * NOT used for: policy enforcement, NAT64, observability.
 */

#include <linux/bpf.h>
#include <linux/if_ether.h>
#include <linux/ip.h>
#include <linux/ipv6.h>
#include <linux/tcp.h>
#include <linux/udp.h>
#include <linux/in.h>
#include <bpf/bpf_helpers.h>
#include <bpf/bpf_endian.h>

#include "maps.h"

#define ETH_P_IP   0x0800
#define ETH_P_IPV6 0x86DD

/* NodePort range: 30000–32767 (standard Kubernetes) */
#define NODEPORT_MIN 30000
#define NODEPORT_MAX 32767

struct {
    __uint(type, BPF_MAP_TYPE_PERCPU_ARRAY);
    __uint(max_entries, 1);
    __type(key, __u32);
    __type(value, __u64);
} xdp_drop_count SEC(".maps");

static __always_inline void count_drop(void)
{
    __u32 key = 0;
    __u64 *cnt = bpf_map_lookup_elem(&xdp_drop_count, &key);
    if (cnt)
        __sync_fetch_and_add(cnt, 1);
}

SEC("xdp")
int sg_xdp_edge(struct xdp_md *ctx)
{
    void *data_end = (void *)(long)ctx->data_end;
    void *data     = (void *)(long)ctx->data;

    struct ethhdr *eth = data;
    if ((void *)(eth + 1) > data_end)
        return XDP_DROP;

    __u16 proto = bpf_ntohs(eth->h_proto);

    if (proto == ETH_P_IP) {
        struct iphdr *ip = (void *)(eth + 1);
        if ((void *)(ip + 1) > data_end)
            return XDP_DROP;

        /* Drop invalid TTL (except 0 from local) */
        if (ip->ttl == 0) {
            count_drop();
            return XDP_DROP;
        }

        /* Drop fragmented packets (not supported in fast path) */
        if (ip->frag_off & bpf_htons(0x1FFF)) {
            return XDP_PASS; /* let kernel handle fragments */
        }

        /* NodePort acceleration: TCP/UDP on NodePort range */
        if (ip->protocol == IPPROTO_TCP) {
            struct tcphdr *tcp = (void *)ip + (ip->ihl << 2);
            if ((void *)(tcp + 1) > data_end)
                return XDP_DROP;

            __u16 dport = bpf_ntohs(tcp->dest);
            if (dport >= NODEPORT_MIN && dport <= NODEPORT_MAX) {
                /* NodePort: look up service and redirect to backend.
                 * Full DNAT in TCX; XDP just fast-paths the lookup. */
                struct service_key skey = {
                    .vip4     = ip->daddr,
                    .port     = tcp->dest,
                    .protocol = IPPROTO_TCP,
                };
                struct service_value *svc = bpf_map_lookup_elem(&service_map, &skey);
                if (svc && svc->backend_count > 0) {
                    /* Pass to TCX for full DNAT processing */
                    return XDP_PASS;
                }
            }
        } else if (ip->protocol == IPPROTO_UDP) {
            struct udphdr *udp = (void *)ip + (ip->ihl << 2);
            if ((void *)(udp + 1) > data_end)
                return XDP_DROP;

            __u16 dport = bpf_ntohs(udp->dest);
            if (dport >= NODEPORT_MIN && dport <= NODEPORT_MAX) {
                struct service_key skey = {
                    .vip4     = ip->daddr,
                    .port     = udp->dest,
                    .protocol = IPPROTO_UDP,
                };
                struct service_value *svc = bpf_map_lookup_elem(&service_map, &skey);
                if (svc && svc->backend_count > 0)
                    return XDP_PASS;
            }
        }
    }

    return XDP_PASS;
}

char _license[] SEC("license") = "Apache-2.0";
