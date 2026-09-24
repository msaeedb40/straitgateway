/* SPDX-License-Identifier: GPL-2.0-only */
/*
 * StraitGateway cgroup BPF Program
 *
 * Status: STUB / SCAFFOLDING
 * Milestone: Phase 7 — Cgroup v2 Socket-Level Traffic Policing & Identity Propagation
 *
 * Implements socket-level enforcement and egress/ingress policing
 * attached to the Kubernetes pod cgroup v2 hierarchy.
 */

#include "vmlinux.h"
#include <bpf/bpf_helpers.h>

/*
 * TODO [Milestone Phase 7]:
 *   1. Associate pod socket with security identity
 *   2. Enforce bandwidth rate limits / token buckets
 *   3. Enforce namespace/pod isolation at cgroup boundary
 */

SEC("cgroup_skb/ingress")
int sg_cgroup_skb_ingress(struct __sk_buff *skb)
{
	return 1; // 1 = allow
}

SEC("cgroup_skb/egress")
int sg_cgroup_skb_egress(struct __sk_buff *skb)
{
	return 1; // 1 = allow
}

char LICENSE[] SEC("license") = "GPL";
