/* SPDX-License-Identifier: GPL-2.0-only */
/*
 * StraitGateway eBPF Flow Map
 *
 * Tracks active flows for observability, latency, and telemetry.
 * Used by sgpktcap and the flow inspection subsystem.
 */

#include "vmlinux.h"
#include <bpf/bpf_helpers.h>
#include "sg_common.h"

/*
 * IPv4 Flow tracking map (LRU for automatic eviction).
 */
struct {
	__uint(type, BPF_MAP_TYPE_LRU_HASH);
	__uint(max_entries, 1048576);
	__uint(pinning, LIBBPF_PIN_BY_NAME);
	__type(key, struct sg_flow_key);
	__type(value, struct sg_flow_val);
} sg_flow_map SEC(".maps");

/*
 * IPv6 Flow tracking map (LRU for automatic eviction).
 */
struct {
	__uint(type, BPF_MAP_TYPE_LRU_HASH);
	__uint(max_entries, 1048576);
	__uint(pinning, LIBBPF_PIN_BY_NAME);
	__type(key, struct sg_flow_v6_key);
	__type(value, struct sg_flow_val);
} sg_flow_v6_map SEC(".maps");

/*
 * Per-CPU flow event ring buffer for userspace consumption.
 */
struct {
	__uint(type, BPF_MAP_TYPE_RINGBUF);
	__uint(max_entries, 16 * 1024 * 1024); /* 16 MiB */
	__uint(pinning, LIBBPF_PIN_BY_NAME);
} sg_flow_events SEC(".maps");

char LICENSE[] SEC("license") = "GPL";
