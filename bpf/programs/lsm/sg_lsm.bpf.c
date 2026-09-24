/* SPDX-License-Identifier: GPL-2.0-only */
/*
 * StraitGateway LSM Hook
 *
 * Status: STUB / SCAFFOLDING
 * Milestone: Phase 8 — LSM Mandatory Security Enforcement
 *
 * Linux Security Module hook for mandatory policy enforcement
 * at the kernel security layer. Enforces StraitGateway policy
 * for socket operations and network access.
 */

#include "vmlinux.h"
#include <bpf/bpf_helpers.h>
#include <bpf/bpf_tracing.h>

/*
 * sg_lsm_socket_connect — enforces policy on socket connect().
 * Attached to lsm/socket_connect.
 */
SEC("lsm/socket_connect")
int BPF_PROG(sg_lsm_socket_connect, struct socket *sock,
             struct sockaddr *address, int addrlen)
{
	/*
	 * TODO [Milestone Phase 8]:
	 *   1. Identify source identity from cgroup/socket context
	 *   2. Look up destination in sg_policy_map
	 *   3. Return 0 (allow) or -EPERM (deny)
	 */
	return 0; /* Allow by default until policy is enforced */
}

char LICENSE[] SEC("license") = "GPL";
