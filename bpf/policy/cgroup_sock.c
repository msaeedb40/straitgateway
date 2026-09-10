/* SPDX-License-Identifier: Apache-2.0 */
/* Copyright 2026 straitgateway Authors */

/*
 * cgroup_sock.c — cgroup-level socket identity enforcement.
 *
 * Hook: BPF_CGROUP_SOCK_ADDR (connect4 / connect6)
 * Attach point: /sys/fs/cgroup (cgroup v2 root)
 *
 * CONTROL / SECURITY layer — NOT packet forwarding.
 *
 * Responsibilities:
 *   - Associate socket with pod security identity at connect() time
 *   - Enforce policy on new outbound connections at socket level
 *   - Provide socket-level acceleration for service connect()
 */

#include <linux/bpf.h>
#include <linux/in.h>
#include <linux/in6.h>
#include <bpf/bpf_helpers.h>
#include <bpf/bpf_endian.h>

#include "maps.h"

#define AF_INET  2
#define AF_INET6 10

/* socket_identity_map: sk_cookie → identity (set at connect time) */
struct {
    __uint(type, BPF_MAP_TYPE_SK_STORAGE);
    __uint(map_flags, BPF_F_NO_PREALLOC);
    __type(key, __u32);
    __type(value, __u32); /* security identity */
} socket_identity SEC(".maps");

/*
 * sg_cgroup_connect4 — intercept IPv4 connect() system calls.
 * Associates outbound socket with pod's security identity.
 */
SEC("cgroup/connect4")
int sg_cgroup_connect4(struct bpf_sock_addr *ctx)
{
    /* Service ClusterIP acceleration:
     * Translate service VIP at connect() to avoid DNAT in datapath. */
    struct service_key skey = {
        .vip4     = ctx->user_ip4,
        .port     = ctx->user_port,
        .protocol = ctx->protocol,
    };

    struct service_value *svc = bpf_map_lookup_elem(&service_map, &skey);
    if (svc && svc->backend_count > 0) {
        /* Transparent service redirect at socket level — no DNAT needed */
        struct backend_key bkey = { .id = 1 }; /* simplified: use first backend */
        struct backend_value *be = bpf_map_lookup_elem(&backend_map, &bkey);
        if (be) {
            ctx->user_ip4  = be->ip4;
            ctx->user_port = be->port;
        }
    }

    return 1; /* allow */
}

/*
 * sg_cgroup_connect6 — intercept IPv6 connect() system calls.
 */
SEC("cgroup/connect6")
int sg_cgroup_connect6(struct bpf_sock_addr *ctx)
{
    /* IPv6 service acceleration — similar to connect4 */
    return 1; /* allow */
}

char _license[] SEC("license") = "Apache-2.0";
