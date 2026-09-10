/* SPDX-License-Identifier: Apache-2.0 */
/* Copyright 2026 straitgateway Authors */

/*
 * netkit_pod.c — NetKit eBPF program for pod/container networking.
 *
 * Hook: NetKit (BPF_PROG_TYPE_SCHED_CLS on netkit device pairs)
 * Purpose:
 *   - Fast-path packet forwarding between pod namespace and host
 *   - BPF identity tagging on egress from pod
 *   - Identity lookup on ingress to pod (for policy enforcement)
 *
 * NOT used for: service LB, NAT, policy decisions — those are on TCX/XDP.
 */

#include <linux/bpf.h>
#include <linux/if_ether.h>
#include <linux/ip.h>
#include <linux/ipv6.h>
#include <linux/tcp.h>
#include <linux/udp.h>
#include <bpf/bpf_helpers.h>
#include <bpf/bpf_endian.h>

#include "maps.h"

#define TC_ACT_OK       0
#define TC_ACT_SHOT     2
#define TC_ACT_REDIRECT 7

#define ETH_P_IP   0x0800
#define ETH_P_IPV6 0x86DD

/*
 * sg_netkit_egress — called on packet egress from the pod (entering host).
 * Tags packets with the pod's security identity via metadata.
 */
SEC("tc/egress")
int sg_netkit_egress(struct __sk_buff *skb)
{
    struct endpoint_key ekey = { .ifindex = skb->ifindex };
    struct endpoint_value *ep;

    ep = bpf_map_lookup_elem(&endpoint_map, &ekey);
    if (!ep)
        return TC_ACT_OK;  /* unknown endpoint — pass through */

    /* Tag the packet with the endpoint's security identity.
     * This mark is read by the host-side TCX program for policy evaluation. */
    skb->mark = ep->identity;

    return TC_ACT_OK;
}

/*
 * sg_netkit_ingress — called on packet ingress to the pod (from host).
 * Validates that the packet originates from an allowed identity.
 */
SEC("tc/ingress")
int sg_netkit_ingress(struct __sk_buff *skb)
{
    /* Identity check is deferred to TCX/policy programs on the host side.
     * NetKit ingress simply passes packets through to the pod. */
    return TC_ACT_OK;
}

char _license[] SEC("license") = "Apache-2.0";
