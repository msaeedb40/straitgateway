/* SPDX-License-Identifier: GPL-2.0-only */
/*
 * StraitGateway eBPF Policy Map
 *
 * Enforces network policy by allowing or dropping packets
 * based on source/destination identity and L4 attributes.
 */

#include "vmlinux.h"
#include <bpf/bpf_helpers.h>
#include <bpf/bpf_endian.h>
#include "sg_common.h"

/*
 * Policy map: identity pair → verdict.
 */
struct {
	__uint(type, BPF_MAP_TYPE_HASH);
	__uint(max_entries, 524288);
	__uint(pinning, LIBBPF_PIN_BY_NAME);
	__type(key, struct sg_policy_key);
	__type(value, struct sg_policy_val);
} sg_policy_map SEC(".maps");

/*
 * Identity map: netkit ifindex → identity ID.
 */
struct {
	__uint(type, BPF_MAP_TYPE_HASH);
	__uint(max_entries, 65536);
	__uint(pinning, LIBBPF_PIN_BY_NAME);
	__type(key, __u32);   /* netkit ifindex */
	__type(value, __u32); /* identity ID */
} sg_identity_map SEC(".maps");

char LICENSE[] SEC("license") = "GPL";
