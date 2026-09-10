/* SPDX-License-Identifier: Apache-2.0 */
/* Copyright 2026 straitgateway Authors */

/*
 * trace.c — OBSERVABILITY only. tracepoints and kprobes for diagnostics.
 *
 * Architectural invariant: kprobes/tracepoints/perf/ringbuf are
 * OBSERVABILITY mechanisms — NOT normal packet-forwarding mechanisms.
 *
 * Hooks: tracepoint, kprobe — NEVER used in the forwarding fast path.
 *
 * Emits flow_event records to the flow_events ring buffer.
 */

#include <linux/bpf.h>
#include <linux/ptrace.h>
#include <bpf/bpf_helpers.h>
#include <bpf/bpf_tracing.h>
#include <bpf/bpf_core_read.h>

#include "maps.h"

struct trace_event_raw_net_dev_xmit {
    __u64 pad;
    void *skbaddr;
    unsigned int len;
    int rc;
    char name[16];
};

struct trace_event_raw_kfree_skb {
    __u64 pad;
    void *skbaddr;
    void *location;
    unsigned short protocol;
    int reason;
};

/*
 * sg_trace_net_dev_xmit — trace packet transmit events.
 * Used for flow visibility, NOT for forwarding decisions.
 */
SEC("tracepoint/net/net_dev_xmit")
int sg_trace_net_dev_xmit(struct trace_event_raw_net_dev_xmit *ctx)
{
    struct flow_event *ev;

    ev = bpf_ringbuf_reserve(&flow_events, sizeof(*ev), 0);
    if (!ev)
        return 0;

    ev->timestamp_ns = bpf_ktime_get_ns();
    ev->direction    = 1; /* egress */
    ev->action       = 0; /* allow */
    ev->drop_reason  = 0;

    bpf_ringbuf_submit(ev, 0);
    return 0;
}

/*
 * sg_trace_skb_drop — trace packet drops for diagnostics.
 */
SEC("tracepoint/skb/kfree_skb")
int sg_trace_skb_drop(struct trace_event_raw_kfree_skb *ctx)
{
    struct flow_event *ev;

    ev = bpf_ringbuf_reserve(&flow_events, sizeof(*ev), 0);
    if (!ev)
        return 0;

    ev->timestamp_ns = bpf_ktime_get_ns();
    ev->drop_reason  = ((__u8)ctx->reason) & 0xFF;
    ev->action       = 1; /* drop */

    bpf_ringbuf_submit(ev, 0);
    return 0;
}

char _license[] SEC("license") = "Apache-2.0";
