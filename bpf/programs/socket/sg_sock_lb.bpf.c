/* SPDX-License-Identifier: GPL-2.0-only */
/*
 * StraitGateway Socket-Level Load Balancer
 *
 * Intercepts connect() and sendmsg() syscalls at cgroup level to
 * transparently redirect service VIPs to backends without going
 * through the network stack.
 *
 * This provides the lowest-latency service load balancing path
 * (socket LB) for pod-to-service traffic.
 */

#include "vmlinux.h"
#include <bpf/bpf_helpers.h>
#include <bpf/bpf_endian.h>
#include "sg_common.h"

/* Service map */
struct {
	__uint(type, BPF_MAP_TYPE_HASH);
	__uint(max_entries, 65536);
	__uint(pinning, LIBBPF_PIN_BY_NAME);
	__type(key, struct sg_service_key);
	__type(value, struct sg_service_val);
} sg_service_map SEC(".maps");

/* Backend map */
struct {
	__uint(type, BPF_MAP_TYPE_HASH);
	__uint(max_entries, 262144);
	__uint(pinning, LIBBPF_PIN_BY_NAME);
	__type(key, struct sg_backend_key);
	__type(value, struct sg_backend_val);
} sg_backend_map SEC(".maps");

/*
 * sg_sock_connect4 — intercepts connect() to translate VIP → backend.
 * Attached to cgroup/connect4.
 */
SEC("cgroup/connect4")
int sg_sock_connect4(struct bpf_sock_addr *ctx)
{
	if (ctx->protocol != IPPROTO_TCP && ctx->protocol != IPPROTO_UDP)
		return 1;

	struct sg_service_key key = {
		.addr = ctx->user_ip4,
		.port = (__u16)ctx->user_port,
		.proto = (__u8)ctx->protocol,
		.pad = 0,
	};

	struct sg_service_val *svc = bpf_map_lookup_elem(&sg_service_map, &key);
	if (!svc || svc->backend_count == 0)
		return 1; /* Not a service VIP, pass transparently */

	/* Select backend: pseudo-random round-robin slot */
	__u32 slot = bpf_get_prandom_u32() % svc->backend_count;

	struct sg_backend_key bkey = {
		.id = svc->backend_id,
		.slot = slot,
	};

	struct sg_backend_val *backend = bpf_map_lookup_elem(&sg_backend_map, &bkey);
	if (!backend || !(backend->flags & SG_BACKEND_FLAG_ACTIVE))
		return 1;

	/* Transparently rewrite destination socket address to chosen pod backend */
	ctx->user_ip4 = backend->addr;
	ctx->user_port = (__u32)backend->port;

	return 1;
}

/*
 * sg_sock_sendmsg4 — intercepts sendmsg() for connectionless UDP.
 * Attached to cgroup/sendmsg4.
 */
SEC("cgroup/sendmsg4")
int sg_sock_sendmsg4(struct bpf_sock_addr *ctx)
{
	if (ctx->protocol != IPPROTO_UDP)
		return 1;

	struct sg_service_key key = {
		.addr = ctx->user_ip4,
		.port = (__u16)ctx->user_port,
		.proto = IPPROTO_UDP,
		.pad = 0,
	};

	struct sg_service_val *svc = bpf_map_lookup_elem(&sg_service_map, &key);
	if (!svc || svc->backend_count == 0)
		return 1;

	__u32 slot = bpf_get_prandom_u32() % svc->backend_count;
	struct sg_backend_key bkey = {
		.id = svc->backend_id,
		.slot = slot,
	};

	struct sg_backend_val *backend = bpf_map_lookup_elem(&sg_backend_map, &bkey);
	if (!backend || !(backend->flags & SG_BACKEND_FLAG_ACTIVE))
		return 1;

	ctx->user_ip4 = backend->addr;
	ctx->user_port = (__u32)backend->port;

	return 1;
}

char LICENSE[] SEC("license") = "GPL";
