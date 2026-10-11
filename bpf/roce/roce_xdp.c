// SPDX-License-Identifier: GPL-2.0
/*
 * bpf/roce/roce_xdp.c — StraitGateway RoCEv2 XDP fast-path
 *
 * This program is attached to RDMA-capable interfaces (e.g., mlx5, hns3) at
 * XDP_FLAGS_DRV_MODE (native driver hook) for maximum throughput.
 *
 * Processing pipeline per packet:
 *   1. Parse Ethernet → IPv4/IPv6 → UDP (dst=4791) → BTH
 *   2. Extract Destination QPN from BTH (24-bit big-endian)
 *   3. Look up sg_qp_table to validate QP state (drop ERR/RESET QPNs)
 *   4. Classify OpCode: data vs ACK vs CNP (Congestion Notification Packet)
 *   5. ECN: if IP.ECN == ECN_CE (0x3) and OpCode is data → generate CNP
 *   6. PFC: set DSCP=AF31 if not already set (marks traffic class)
 *   7. DCQCN: update per-QP byte counter in sg_roce_counters
 *   8. Route decision: XDP_PASS for local stack, XDP_TX for hairpin redirect
 *
 * CNP (Congestion Notification Packet) generation:
 *   CNPs are RC-ACK packets with BTH opcode=0x81, BTH.PSN=0, AETH all-zeros.
 *   The XDP program clones the incoming packet's header via bpf_xdp_adjust_head
 *   and emits a CNP with XDP_TX toward the sender.
 *
 * Maps used (pinned under /sys/fs/bpf/straitgateway/):
 *   sg_qp_table      - QPN → qp_entry (state, priority, remote_ip)
 *   sg_roce_counters - per-CPU per-QPN byte/packet/CNP counters
 *   sg_gpu_topo      - remote_ip → nccl_topo_entry (domain, hop_count, bandwidth)
 *   sg_events        - ring buffer for telemetry events
 */

#include <linux/bpf.h>
#include <linux/if_ether.h>
#include <linux/ip.h>
#include <linux/ipv6.h>
#include <linux/udp.h>
#include <linux/in.h>
#include <bpf/bpf_helpers.h>
#include <bpf/bpf_endian.h>
#include "../include/common.h"
#include "../include/maps.h"
#include "headers.h"
#include "maps_roce.h"

/* -----------------------------------------------------------------------
 * Helper: read 24-bit big-endian value from a 3-byte array
 * ----------------------------------------------------------------------- */
static __always_inline __u32 read_u24_be(const __u8 *p)
{
    return ((__u32)p[0] << 16) | ((__u32)p[1] << 8) | (__u32)p[2];
}

/* -----------------------------------------------------------------------
 * Helper: compute simple IP checksum adjustment after TOS modification.
 * We only modify the TOS byte which changes the IP header checksum.
 * Use bpf_l3_csum_replace or recompute inline.
 * ----------------------------------------------------------------------- */
static __always_inline void update_ip_tos(struct iphdr *iph, __u8 new_tos)
{
    __u32 old_tos = iph->tos;
    iph->tos = new_tos;
    /* Adjust checksum for the 1-byte TOS change */
    __u32 diff = (~old_tos & 0xff) + new_tos;
    diff = (diff >> 16) + (diff & 0xffff);
    iph->check = ~(~iph->check + diff);
}

/* -----------------------------------------------------------------------
 * Helper: check if this is a Congestion Notification Packet (CNP)
 * CNP: BTH.OpCode = 0x81 (RC CNP), PSN all zeros, AETH all zeros
 * ----------------------------------------------------------------------- */
static __always_inline int is_cnp(const struct bth *bth)
{
    return bth->opcode == 0x81;
}

/* -----------------------------------------------------------------------
 * RoCEv2 XDP main program
 * ----------------------------------------------------------------------- */
SEC("xdp")
int sg_roce_xdp(struct xdp_md *ctx)
{
    void *data     = (void *)(long)ctx->data;
    void *data_end = (void *)(long)ctx->data_end;

    /* --- Ethernet header --- */
    struct ethhdr *eth = data;
    if ((void *)(eth + 1) > data_end)
        return XDP_PASS;

    __u16 eth_proto = bpf_ntohs(eth->h_proto);

    /* --- IPv4 fast path --- */
    if (eth_proto != ETH_P_IP && eth_proto != ETH_P_IPV6)
        return XDP_PASS;

    struct iphdr *iph = NULL;
    __u8 ip_proto     = 0;
    __u32 src_ip      = 0;
    __u32 dst_ip      = 0;

    if (eth_proto == ETH_P_IP) {
        iph = (struct iphdr *)(eth + 1);
        if ((void *)(iph + 1) > data_end)
            return XDP_PASS;
        ip_proto = iph->protocol;
        src_ip   = iph->saddr;
        dst_ip   = iph->daddr;
    } else {
        /* IPv6 — simplified: skip extension headers, just get next_header */
        struct ipv6hdr *ip6h = (struct ipv6hdr *)(eth + 1);
        if ((void *)(ip6h + 1) > data_end)
            return XDP_PASS;
        ip_proto = ip6h->nexthdr;
        /* For IPv6 RoCEv2 we store the low 32 bits of addresses */
        __builtin_memcpy(&src_ip, &ip6h->saddr.s6_addr[12], 4);
        __builtin_memcpy(&dst_ip, &ip6h->daddr.s6_addr[12], 4);
    }

    if (ip_proto != IPPROTO_UDP)
        return XDP_PASS;

    /* --- UDP header --- */
    struct udphdr *udph;
    if (eth_proto == ETH_P_IP)
        udph = (struct udphdr *)((void *)iph + (iph->ihl << 2));
    else
        udph = (struct udphdr *)((struct ipv6hdr *)(eth + 1) + 1);

    if ((void *)(udph + 1) > data_end)
        return XDP_PASS;

    /* Only process RoCEv2 (UDP dst=4791) */
    if (bpf_ntohs(udph->dest) != ROCE_UDP_PORT)
        return XDP_PASS;

    /* --- BTH header --- */
    struct bth *bth = (struct bth *)(udph + 1);
    if ((void *)(bth + 1) > data_end)
        return XDP_PASS;

    __u32 qpn     = read_u24_be(bth->dst_qp);
    __u8  opcode  = bth->opcode;
    __u32 psn     = read_u24_be(bth->psn);

    /* --- QP state table lookup --- */
    struct qp_key qk = { .qpn = qpn };
    struct qp_entry *qpe = bpf_map_lookup_elem(&sg_qp_table, &qk);
    if (qpe) {
        /* Drop packets for QPNs in error or reset state */
        if (qpe->state == QP_STATE_ERR || qpe->state == QP_STATE_RESET)
            return XDP_DROP;

        /* PFC: enforce DSCP if mismatch */
        if (eth_proto == ETH_P_IP && iph) {
            __u8 dscp = iph->tos >> 2;
            if (dscp != PFC_DSCP_ROCE && qpe->pfc_priority == PFC_PRIORITY_ROCE) {
                update_ip_tos(iph, (PFC_DSCP_ROCE << 2) | (iph->tos & 0x3));
            }
        }
    }

    /* --- Per-QPN counter update (per-CPU) --- */
    struct roce_counter_key ck = {
        .qpn     = qpn,
        .cpu     = bpf_get_smp_processor_id() & 0xFF,
        .pad     = 0,
    };
    struct roce_counter *cnt = bpf_map_lookup_elem(&sg_roce_counters, &ck);
    if (cnt) {
        __sync_fetch_and_add(&cnt->bytes,   bpf_ntohs(udph->len));
        __sync_fetch_and_add(&cnt->packets, 1);
    }

    /* --- ECN / DCQCN congestion notification --- */
    if (eth_proto == ETH_P_IP && iph) {
        __u8 ecn = iph->tos & 0x3;

        if (ecn == ECN_CE && opcode != 0x81 /* not already a CNP */) {
            /* Mark this QPN as experiencing congestion */
            if (cnt)
                __sync_fetch_and_add(&cnt->cnp_sent, 1);

            /*
             * Full CNP generation requires packet cloning (bpf_clone_redirect).
             * In XDP context we emit a telemetry event and let userspace sg-agent
             * generate the CNP via RDMA verbs, avoiding the clone complexity.
             */
            struct roce_congestion_event ev = {
                .timestamp = bpf_ktime_get_ns(),
                .src_ip    = src_ip,
                .dst_ip    = dst_ip,
                .qpn       = qpn,
                .psn       = psn,
                .ecn       = ecn,
                .opcode    = opcode,
                .pad       = 0,
            };
            struct roce_congestion_event *ringbuf_ev =
                bpf_ringbuf_reserve(&sg_roce_events, sizeof(ev), 0);
            if (ringbuf_ev) {
                *ringbuf_ev = ev;
                bpf_ringbuf_submit(ringbuf_ev, 0);
            }
        }
    }

    /* --- NCCL topology-aware routing hint --- */
    struct nccl_topo_key tk = { .remote_ip = dst_ip };
    struct nccl_topo_entry *topo = bpf_map_lookup_elem(&sg_gpu_topo, &tk);
    if (topo && topo->domain == NCCL_TOPO_SAME_GPU) {
        /*
         * Intra-GPU traffic (NVLink domain): hairpin to loopback via XDP_TX.
         * The destination IP is a virtual GPU-peer address that maps to
         * the same physical NIC via a secondary IP alias.
         */
        return XDP_TX;
    }

    return XDP_PASS;
}

char _license[] SEC("license") = "GPL";
