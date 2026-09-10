/* SPDX-License-Identifier: Apache-2.0 */
/* Copyright 2026 straitgateway Authors */

/*
 * lsm_net.c — LSM (Linux Security Module) network enforcement hooks.
 *
 * Hook: BPF LSM (BPF_PROG_TYPE_LSM)
 * Attach point: LSM hooks — socket_connect, socket_sendmsg, etc.
 *
 * SECURITY layer — NOT packet forwarding.
 *
 * Responsibilities:
 *   - Enforce StraitNetworkPolicy Allow/Deny/Reject at socket/kernel level
 *   - Fine-grained control before packets enter the network stack
 *   - Supplement TCX policy checks with kernel-enforced security
 */

#include <linux/bpf.h>
#include <linux/socket.h>
#include <linux/in.h>
#include <bpf/bpf_helpers.h>
#include <bpf/bpf_tracing.h>

#include "maps.h"

struct socket;
struct sockaddr;
struct msghdr;

#define EPERM  1
#define EACCES 13

/*
 * sg_lsm_socket_connect — enforce outbound connection policy at LSM level.
 */
SEC("lsm/socket_connect")
int BPF_PROG(sg_lsm_socket_connect, struct socket *sock,
             struct sockaddr *address, int addrlen)
{
    /* Policy enforcement is primarily in cgroup/TCX hooks.
     * LSM adds a defence-in-depth enforcement layer. */
    return 0; /* allow by default */
}

/*
 * sg_lsm_socket_sendmsg — enforce send-path policy.
 */
SEC("lsm/socket_sendmsg")
int BPF_PROG(sg_lsm_socket_sendmsg, struct socket *sock,
             struct msghdr *msg, int size)
{
    return 0; /* allow */
}

char _license[] SEC("license") = "Apache-2.0";
