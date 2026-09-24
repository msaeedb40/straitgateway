/* SPDX-License-Identifier: GPL-2.0-only */
/*
 * StraitGateway eBPF Route Map
 *
 * LPM trie routing tables for IPv4 and IPv6 egress/transit routing.
 */

#include "vmlinux.h"
#include <bpf/bpf_helpers.h>
#include "sg_common.h"

/* IPv4 LPM route map */
struct {
	__uint(type, BPF_MAP_TYPE_LPM_TRIE);
	__uint(max_entries, 65536);
	__uint(map_flags, BPF_F_NO_PREALLOC);
	__uint(pinning, LIBBPF_PIN_BY_NAME);
	__type(key, struct sg_route_key);
	__type(value, struct sg_route_val);
} sg_route_map SEC(".maps");

/* IPv6 LPM route map */
struct {
	__uint(type, BPF_MAP_TYPE_LPM_TRIE);
	__uint(max_entries, 65536);
	__uint(map_flags, BPF_F_NO_PREALLOC);
	__uint(pinning, LIBBPF_PIN_BY_NAME);
	__type(key, struct sg_route_v6_key);
	__type(value, struct sg_route_v6_val);
} sg_route_v6_map SEC(".maps");

char LICENSE[] SEC("license") = "GPL";
