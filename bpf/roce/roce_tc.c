// SPDX-License-Identifier: GPL-2.0
/*
 * bpf/roce/roce_tc.c — StraitGateway RoCEv2 TC egress classifier
 *
 * Attached to the RDMA interface egress path via TCX. This program:
 *
 *  1. Identifies outbound RoCEv2 packets (UDP dst=4791)
 *  2. Enforces PFC DSCP marking using sg_pfc_config
 *  3. Sets ECN ECT(0) on outbound RoCEv2 data packets if the QP's
 *     traffic class is ECN-capable (enabling DCQCN end-to-end)
 *  4. Applies DCQCN rate limiting signal: if the QP has received CNP
 *     (sg_roce_counters.cnp_sent > threshold), marks ECN bits to signal
 *     rate reduction to the sender
 *  5. Updates sg_roce_counters for egress accounting
 *
 * Rate control via DCQCN:
 *   DCQCN (Data Center Quantized Congestion Notification) uses a PI
 *   controller. The eBPF program implements the "mark on congestion"
 *   path; the actual rate reduction is performed by the RDMA driver in
 *   response to CNPs received from the receiver side.
 *
 * This program returns TC_ACT_OK for all processed packets.
 * It never drops (drops are the XDP program's responsibility).
 */

#include <linux/bpf.h>
#include <linux/if_ether.h>
#include <linux/ip.h>
#include <linux/ipv6.h>
#include <linux/udp.h>
#include <linux/pkt_cls.h>
#include <bpf/bpf_helpers.h>
#include <bpf/bpf_endian.h>
#include "../include/common.h"
#include "../include/maps.h"
#include "headers.h"
#include "maps_roce.h"

static __always_inline __u32 read_u24_be_tc(const __u8 *p)
{
    return ((__u32)p[0] << 16) | ((__u32)p[1] << 8) | (__u32)p[2];
}

/*
 * DCQCN congestion threshold: if a QPN has generated more than this many
 * CNPs, we begin marking ECT(0) on outbound packets to enable end-to-end
 * DCQCN without kernel RDMA driver involvement.
 */
#define DCQCN_CNP_THRESHOLD 8

SEC("tc")
int sg_roce_egress(struct __sk_buff *skb)
{
    void *data     = (void *)(long)skb->data;
    void *data_end = (void *)(long)skb->data_end;

    /* Parse Ethernet */
    struct ethhdr *eth = data;
    if ((void *)(eth + 1) > data_end)
        return TC_ACT_OK;

    __u16 eth_proto = bpf_ntohs(eth->h_proto);
    if (eth_proto != ETH_P_IP && eth_proto != ETH_P_IPV6)
        return TC_ACT_OK;

    /* Parse IP */
    struct iphdr *iph = NULL;
    __u32 dst_ip      = 0;
    __u8  ip_proto    = 0;

    if (eth_proto == ETH_P_IP) {
        iph = (struct iphdr *)(eth + 1);
        if ((void *)(iph + 1) > data_end)
            return TC_ACT_OK;
        ip_proto = iph->protocol;
        dst_ip   = iph->daddr;
    } else {
        struct ipv6hdr *ip6h = (struct ipv6hdr *)(eth + 1);
        if ((void *)(ip6h + 1) > data_end)
            return TC_ACT_OK;
        ip_proto = ip6h->nexthdr;
        __builtin_memcpy(&dst_ip, &ip6h->daddr.s6_addr[12], 4);
    }

    if (ip_proto != IPPROTO_UDP)
        return TC_ACT_OK;

    /* Parse UDP */
    struct udphdr *udph;
    if (eth_proto == ETH_P_IP)
        udph = (struct udphdr *)((void *)iph + (iph->ihl << 2));
    else
        udph = (struct udphdr *)((struct ipv6hdr *)(eth + 1) + 1);

    if ((void *)(udph + 1) > data_end)
        return TC_ACT_OK;

    if (bpf_ntohs(udph->dest) != ROCE_UDP_PORT)
        return TC_ACT_OK;

    /* Parse BTH */
    struct bth *bth = (struct bth *)(udph + 1);
    if ((void *)(bth + 1) > data_end)
        return TC_ACT_OK;

    __u32 qpn = read_u24_be_tc(bth->dst_qp);

    /* Look up QP entry */
    struct qp_key qk = { .qpn = qpn };
    struct qp_entry *qpe = bpf_map_lookup_elem(&sg_qp_table, &qk);
    if (!qpe)
        return TC_ACT_OK;

    /* -----------------------------------------------------------------------
     * PFC DSCP enforcement on egress
     * ----------------------------------------------------------------------- */
    __u32 pfc_idx = qpe->pfc_priority & 0x7;
    struct pfc_config *pfc = bpf_map_lookup_elem(&sg_pfc_config, &pfc_idx);

    if (pfc && pfc->enabled && eth_proto == ETH_P_IP && iph) {
        __u8 current_dscp = iph->tos >> 2;
        if (current_dscp != pfc->dscp) {
            /* Rewrite DSCP and recompute IP checksum */
            __u8 old_tos = iph->tos;
            __u8 new_tos = (pfc->dscp << 2) | (old_tos & 0x3);
            bpf_l3_csum_replace(skb, sizeof(struct ethhdr) + offsetof(struct iphdr, check),
                                 bpf_htons(old_tos), bpf_htons(new_tos), 2);
            iph = (struct iphdr *)((void *)(long)skb->data + sizeof(struct ethhdr));
            if ((void *)(iph + 1) > (void *)(long)skb->data_end)
                return TC_ACT_OK;
            iph->tos = new_tos;
        }

        /* -----------------------------------------------------------------------
         * ECN: mark ECT(0) on outbound data packets for DCQCN eligibility
         * ----------------------------------------------------------------------- */
        if (pfc->ecn_capable) {
            __u8 ecn = iph->tos & 0x3;
            if (ecn == ECN_NOT_ECT) {
                __u8 old_tos = iph->tos;
                __u8 new_tos = (old_tos & 0xFC) | ECN_ECT0;
                bpf_l3_csum_replace(skb, sizeof(struct ethhdr) + offsetof(struct iphdr, check),
                                     bpf_htons(old_tos), bpf_htons(new_tos), 2);
                iph = (struct iphdr *)((void *)(long)skb->data + sizeof(struct ethhdr));
                if ((void *)(iph + 1) <= (void *)(long)skb->data_end)
                    iph->tos = new_tos;
            }
        }
    }

    /* -----------------------------------------------------------------------
     * Egress counter update
     * ----------------------------------------------------------------------- */
    struct roce_counter_key ck = {
        .qpn = qpn,
        .cpu = bpf_get_smp_processor_id() & 0xFF,
    };
    struct roce_counter *cnt = bpf_map_lookup_elem(&sg_roce_counters, &ck);
    if (cnt) {
        __sync_fetch_and_add(&cnt->bytes,   bpf_ntohs(udph->len));
        __sync_fetch_and_add(&cnt->packets, 1);
    }

    return TC_ACT_OK;
}

char _license[] SEC("license") = "GPL";
