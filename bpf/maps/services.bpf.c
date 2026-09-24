/* SPDX-License-Identifier: GPL-2.0-only */
/*
 * StraitGateway eBPF Service Map
 *
 * Maps service VIPs to backend endpoints for the kube-proxy replacement.
 * Used by the Socket LB, TCX, and XDP programs.
 */

#include "vmlinux.h"
#include <bpf/bpf_helpers.h>
#include <bpf/bpf_endian.h>
#include "sg_common.h"

/*
 * IPv4 Service map: VIP → service metadata.
 */
struct {
	__uint(type, BPF_MAP_TYPE_HASH);
	__uint(max_entries, 65536);
	__uint(pinning, LIBBPF_PIN_BY_NAME);
	__type(key, struct sg_service_key);
	__type(value, struct sg_service_val);
} sg_service_map SEC(".maps");

/*
 * IPv6 Service map: VIP → service metadata.
 */
struct {
	__uint(type, BPF_MAP_TYPE_HASH);
	__uint(max_entries, 65536);
	__uint(pinning, LIBBPF_PIN_BY_NAME);
	__type(key, struct sg_service_v6_key);
	__type(value, struct sg_service_val);
} sg_service_v6_map SEC(".maps");

/*
 * IPv4 Backend map: backend_id + slot → endpoint.
 */
struct {
	__uint(type, BPF_MAP_TYPE_HASH);
	__uint(max_entries, 262144);
	__uint(pinning, LIBBPF_PIN_BY_NAME);
	__type(key, struct sg_backend_key);
	__type(value, struct sg_backend_val);
} sg_backend_map SEC(".maps");

/*
 * IPv6 Backend map: backend_id + slot → endpoint.
 */
struct {
	__uint(type, BPF_MAP_TYPE_HASH);
	__uint(max_entries, 262144);
	__uint(pinning, LIBBPF_PIN_BY_NAME);
	__type(key, struct sg_backend_key);
	__type(value, struct sg_backend_v6_val);
} sg_backend_v6_map SEC(".maps");

char LICENSE[] SEC("license") = "GPL";
