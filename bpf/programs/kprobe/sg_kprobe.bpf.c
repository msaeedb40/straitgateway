/* SPDX-License-Identifier: GPL-2.0-only */
/*
 * StraitGateway Kernel Probe Program
 *
 * Status: STUB / SCAFFOLDING
 * Milestone: Phase 9 — Kernel Observability, TCP Drop Tracing & Flow Diagnostics
 *
 * Observes TCP state changes and packet drops for flow monitoring.
 */

#include "vmlinux.h"
#include <bpf/bpf_helpers.h>

/*
 * TODO [Milestone Phase 9]:
 *   1. Extract socket context from pt_regs
 *   2. Match flow against sg_flow_map
 *   3. Record drop reason and emit to sg_flow_events ring buffer
 */

SEC("kprobe/tcp_drop")
int sg_kprobe_tcp_drop(struct pt_regs *ctx)
{
	return 0;
}

char LICENSE[] SEC("license") = "GPL";
