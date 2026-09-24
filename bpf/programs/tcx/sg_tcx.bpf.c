/* SPDX-License-Identifier: GPL-2.0-only */
/*
 * StraitGateway TCX Program
 *
 * Traffic Control eXpress (TCX) hooks attached to Netkit interfaces.
 * Handles pod-egress and pod-ingress packet processing:
 *   - Service load balancing (DNAT to backend)
 *   - Policy enforcement
 *   - Flow tracking and observability
 *
 * Attached via TCX to the Netkit parent (host side) for each pod.
 */

#include "vmlinux.h"
#include <bpf/bpf_helpers.h>
#include <bpf/bpf_endian.h>
#include "sg_common.h"

/* Service and backend maps */
struct {
	__uint(type, BPF_MAP_TYPE_HASH);
	__uint(max_entries, 65536);
	__uint(pinning, LIBBPF_PIN_BY_NAME);
	__type(key, struct sg_service_key);
	__type(value, struct sg_service_val);
} sg_service_map SEC(".maps");

struct {
	__uint(type, BPF_MAP_TYPE_HASH);
	__uint(max_entries, 262144);
	__uint(pinning, LIBBPF_PIN_BY_NAME);
	__type(key, struct sg_backend_key);
	__type(value, struct sg_backend_val);
} sg_backend_map SEC(".maps");

/* Policy map */
struct {
	__uint(type, BPF_MAP_TYPE_HASH);
	__uint(max_entries, 524288);
	__uint(pinning, LIBBPF_PIN_BY_NAME);
	__type(key, struct sg_policy_key);
	__type(value, struct sg_policy_val);
} sg_policy_map SEC(".maps");

/* Identity map */
struct {
	__uint(type, BPF_MAP_TYPE_HASH);
	__uint(max_entries, 65536);
	__uint(pinning, LIBBPF_PIN_BY_NAME);
	__type(key, __u32);   /* ifindex */
	__type(value, __u32); /* security_id */
} sg_identity_map SEC(".maps");

/*
 * sg_tcx_pod_egress — processes packets leaving a pod (pod → node).
 * Attached to the Netkit parent interface, ingress direction.
 */
SEC("tc")
int sg_tcx_pod_egress(struct __sk_buff *skb)
{
	void *data     = (void *)(long)skb->data;
	void *data_end = (void *)(long)skb->data_end;

	struct ethhdr *eth = data;
	if (!check_pkt_bounds(eth, sizeof(*eth), data_end))
		return TC_ACT_OK;

	/* Dual-stack check */
	if (eth->h_proto == bpf_htons(ETH_P_IPV6)) {
		struct ipv6hdr *ip6h = (void *)(eth + 1);
		if (!check_pkt_bounds(ip6h, sizeof(*ip6h), data_end))
			return TC_ACT_OK;
		/* IPv6 packets are passed through to the Linux stack until stateful conntrack is initialized */
		return TC_ACT_OK;
	}

	if (eth->h_proto != bpf_htons(ETH_P_IP))
		return TC_ACT_OK;

	struct iphdr *iph = (void *)(eth + 1);
	if (!check_pkt_bounds(iph, sizeof(*iph), data_end))
		return TC_ACT_OK;

	__u32 ifindex = skb->ifindex;
	__u32 src_id = 0;
	__u32 *id_ptr = bpf_map_lookup_elem(&sg_identity_map, &ifindex);
	if (id_ptr)
		src_id = *id_ptr;

	__u16 dport = 0;
	if (iph->protocol == IPPROTO_TCP) {
		struct tcphdr *tcp = (void *)iph + (iph->ihl * 4);
		if (check_pkt_bounds(tcp, sizeof(*tcp), data_end))
			dport = tcp->dest;
	} else if (iph->protocol == IPPROTO_UDP) {
		struct udphdr *udp = (void *)iph + (iph->ihl * 4);
		if (check_pkt_bounds(udp, sizeof(*udp), data_end))
			dport = udp->dest;
	}

	/* 1. Policy check (egress) */
	if (src_id > 0) {
		struct sg_policy_key pol_key = {
			.src_id = src_id,
			.dst_id = 0, // wild-card/external
			.dst_port = dport,
			.proto = iph->protocol,
			.dir = 1, // egress
		};
		struct sg_policy_val *verdict = bpf_map_lookup_elem(&sg_policy_map, &pol_key);
		if (verdict && verdict->verdict == SG_POLICY_DENY)
			return TC_ACT_SHOT;
	}

	/* 2. Service VIP DNAT */
	struct sg_service_key svc_key = {
		.addr = iph->daddr,
		.port = dport,
		.proto = iph->protocol,
		.pad = 0,
	};
	struct sg_service_val *svc = bpf_map_lookup_elem(&sg_service_map, &svc_key);
	if (svc && svc->backend_count > 0) {
		__u32 slot = bpf_get_prandom_u32() % svc->backend_count;
		struct sg_backend_key bkey = {
			.id = svc->backend_id,
			.slot = slot,
		};
		struct sg_backend_val *be = bpf_map_lookup_elem(&sg_backend_map, &bkey);
		if (be && (be->flags & SG_BACKEND_FLAG_ACTIVE)) {
			__u32 old_daddr = iph->daddr;
			__u32 new_daddr = be->addr;

			/* Rewrite destination IP */
			iph->daddr = new_daddr;
			bpf_l3_csum_replace(skb, sizeof(struct ethhdr) + offsetof(struct iphdr, check), old_daddr, new_daddr, 4);

			/* Update L4 checksum for L3 destination rewrite and port modification */
			if (iph->protocol == IPPROTO_TCP) {
				struct tcphdr *tcp = (void *)iph + (iph->ihl * 4);
				if (check_pkt_bounds(tcp, sizeof(*tcp), data_end)) {
					__u32 l4_csum_off = sizeof(struct ethhdr) + (iph->ihl * 4) + offsetof(struct tcphdr, check);
					bpf_l4_csum_replace(skb, l4_csum_off, old_daddr, new_daddr, BPF_F_PSEUDO_HDR | 4);
					if (dport != be->port) {
						tcp->dest = be->port;
						bpf_l4_csum_replace(skb, l4_csum_off, dport, be->port, 2);
					}
				}
			} else if (iph->protocol == IPPROTO_UDP) {
				struct udphdr *udp = (void *)iph + (iph->ihl * 4);
				if (check_pkt_bounds(udp, sizeof(*udp), data_end)) {
					if (udp->check != 0) {
						__u32 l4_csum_off = sizeof(struct ethhdr) + (iph->ihl * 4) + offsetof(struct udphdr, check);
						bpf_l4_csum_replace(skb, l4_csum_off, old_daddr, new_daddr, BPF_F_PSEUDO_HDR | 4);
						if (dport != be->port) {
							udp->dest = be->port;
							bpf_l4_csum_replace(skb, l4_csum_off, dport, be->port, 2);
						}
					} else {
						udp->dest = be->port;
					}
				}
			}
		}
	}

	return TC_ACT_OK;
}

/*
 * sg_tcx_pod_ingress — processes packets arriving at a pod (node → pod).
 * Attached to the Netkit parent interface, egress direction.
 */
SEC("tc")
int sg_tcx_pod_ingress(struct __sk_buff *skb)
{
	void *data     = (void *)(long)skb->data;
	void *data_end = (void *)(long)skb->data_end;

	struct ethhdr *eth = data;
	if (!check_pkt_bounds(eth, sizeof(*eth), data_end))
		return TC_ACT_OK;

	/* Dual-stack check */
	if (eth->h_proto == bpf_htons(ETH_P_IPV6)) {
		struct ipv6hdr *ip6h = (void *)(eth + 1);
		if (!check_pkt_bounds(ip6h, sizeof(*ip6h), data_end))
			return TC_ACT_OK;
		return TC_ACT_OK;
	}

	if (eth->h_proto != bpf_htons(ETH_P_IP))
		return TC_ACT_OK;

	struct iphdr *iph = (void *)(eth + 1);
	if (!check_pkt_bounds(iph, sizeof(*iph), data_end))
		return TC_ACT_OK;

	__u32 ifindex = skb->ifindex;
	__u32 dst_id = 0;
	__u32 *id_ptr = bpf_map_lookup_elem(&sg_identity_map, &ifindex);
	if (id_ptr)
		dst_id = *id_ptr;

	__u16 dport = 0;
	if (iph->protocol == IPPROTO_TCP) {
		struct tcphdr *tcp = (void *)iph + (iph->ihl * 4);
		if (check_pkt_bounds(tcp, sizeof(*tcp), data_end))
			dport = tcp->dest;
	} else if (iph->protocol == IPPROTO_UDP) {
		struct udphdr *udp = (void *)iph + (iph->ihl * 4);
		if (check_pkt_bounds(udp, sizeof(*udp), data_end))
			dport = udp->dest;
	}

	/* Policy check (ingress) */
	if (dst_id > 0) {
		struct sg_policy_key pol_key = {
			.src_id = 0,
			.dst_id = dst_id,
			.dst_port = dport,
			.proto = iph->protocol,
			.dir = 0, // ingress
		};
		struct sg_policy_val *verdict = bpf_map_lookup_elem(&sg_policy_map, &pol_key);
		if (verdict && verdict->verdict == SG_POLICY_DENY)
			return TC_ACT_SHOT;
	}

	return TC_ACT_OK;
}

char LICENSE[] SEC("license") = "GPL";
