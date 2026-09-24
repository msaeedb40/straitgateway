/* SPDX-License-Identifier: GPL-2.0-only */
/*
 * StraitGateway XDP Program
 *
 * Status: STUB / SCAFFOLDING
 * Milestone: Phase 6 — XDP Acceleration, DDoS Mitigation & NodePort Acceleration
 *
 * Earliest ingress processing point. Handles:
 *   - Fast-path packet filtering and policy enforcement
 *   - DDoS mitigation / rate limiting
 *   - NodePort external packet acceleration
 *
 * Attached via XDP to the node's physical/uplink interface.
 */

#include "vmlinux.h"
#include <bpf/bpf_helpers.h>
#include <bpf/bpf_endian.h>
#include "sg_common.h"

/*
 * sg_xdp_ingress — main XDP hook for ingress processing.
 *
 * Returns XDP_PASS to continue to the kernel stack,
 * XDP_DROP to silently discard, XDP_TX for hairpin redirect.
 */
SEC("xdp")
int sg_xdp_ingress(struct xdp_md *ctx)
{
	void *data     = (void *)(long)ctx->data;
	void *data_end = (void *)(long)ctx->data_end;

	struct ethhdr *eth = data;
	if ((void *)(eth + 1) > data_end)
		return XDP_PASS;

	__u16 proto = bpf_ntohs(eth->h_proto);

	/* Only process IPv4 and IPv6 */
	if (proto != ETH_P_IP && proto != ETH_P_IPV6)
		return XDP_PASS;

	/*
	 * TODO [Milestone Phase 6]:
	 *   1. Lookup destination in sg_service_map for NodePort/LB VIPs
	 *   2. Check sg_policy_map for policy verdict
	 *   3. Update sg_flow_map counters
	 *   4. Emit flow events to sg_flow_events ring buffer
	 */

	return XDP_PASS;
}

char LICENSE[] SEC("license") = "GPL";
