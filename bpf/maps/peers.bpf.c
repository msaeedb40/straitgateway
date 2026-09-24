/* SPDX-License-Identifier: GPL-2.0-only */
/*
 * StraitGateway eBPF Peers Map — transit gateway and mesh peers
 */

#include "vmlinux.h"
#include <bpf/bpf_helpers.h>
#include "sg_common.h"

/* IPv4 Peers map */
struct {
	__uint(type, BPF_MAP_TYPE_HASH);
	__uint(max_entries, 4096);
	__uint(pinning, LIBBPF_PIN_BY_NAME);
	__type(key, struct sg_peer_key);
	__type(value, struct sg_peer_val);
} sg_peers_map SEC(".maps");

/* IPv6 Peers map */
struct {
	__uint(type, BPF_MAP_TYPE_HASH);
	__uint(max_entries, 4096);
	__uint(pinning, LIBBPF_PIN_BY_NAME);
	__type(key, struct sg_peer_key);
	__type(value, struct sg_peer_v6_val);
} sg_peers_v6_map SEC(".maps");

char LICENSE[] SEC("license") = "GPL";
