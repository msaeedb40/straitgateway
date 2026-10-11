// SPDX-License-Identifier: GPL-2.0
/*
 * bpf/roce/maps_roce.h — RoCEv2-specific BPF map declarations
 *
 * These maps are pinned under /sys/fs/bpf/straitgateway/roce/ and are managed
 * by the Go-side internal/dataplane/roce.Manager.
 *
 * Map summary:
 *   sg_qp_table      Hash[qp_key → qp_entry]        QP state + priority
 *   sg_roce_counters PercpuHash[roce_counter_key → roce_counter]
 *   sg_gpu_topo      Hash[nccl_topo_key → nccl_topo_entry]  GPU topology
 *   sg_roce_events   Ringbuf                          ECN/CNP events
 *   sg_pfc_config    Array[__u32 priority → pfc_config]     PFC per-prio cfg
 */

#ifndef __STRAITGATEWAY_MAPS_ROCE_H__
#define __STRAITGATEWAY_MAPS_ROCE_H__

#include "../include/common.h"
#include "headers.h"

/* -----------------------------------------------------------------------
 * sg_qp_table — Queue Pair state table
 * Key: destination QPN (24-bit, zero-extended to 32-bit)
 * Value: qp_entry struct
 * ----------------------------------------------------------------------- */
struct qp_key {
    __u32 qpn;
};

struct qp_entry {
    __u8  state;        /* QP_STATE_* constants from headers.h */
    __u8  pfc_priority; /* PFC traffic class (0-7) for this QP */
    __u16 mtu;          /* Path MTU in bytes */
    __u32 remote_ip;    /* Peer's IPv4 address */
    __u32 rq_psn;       /* Expected receive PSN */
    __u32 sq_psn;       /* Next send PSN */
    __u64 last_seen_ns; /* ktime_get_ns of last packet on this QP */
};

struct {
    __uint(type, BPF_MAP_TYPE_HASH);
    __type(key,  struct qp_key);
    __type(value, struct qp_entry);
    __uint(max_entries, 131072);  /* 128K QPs per node */
    __uint(pinning, LIBBPF_PIN_BY_NAME);
} sg_qp_table SEC(".maps");

/* -----------------------------------------------------------------------
 * sg_roce_counters — Per-CPU per-QPN byte/packet/CNP counters
 * Key: {qpn, cpu}
 * ----------------------------------------------------------------------- */
struct roce_counter_key {
    __u32 qpn;
    __u8  cpu;
    __u8  pad[3];
};

struct roce_counter {
    __u64 bytes;
    __u64 packets;
    __u64 cnp_sent;     /* Congestion Notification Packets generated */
    __u64 ecn_ce_seen;  /* ECN-CE marked packets received */
    __u64 drops;        /* Packets dropped (bad QP state, etc.) */
};

struct {
    __uint(type, BPF_MAP_TYPE_PERCPU_HASH);
    __type(key,  struct roce_counter_key);
    __type(value, struct roce_counter);
    __uint(max_entries, 131072);
    __uint(pinning, LIBBPF_PIN_BY_NAME);
} sg_roce_counters SEC(".maps");

/* -----------------------------------------------------------------------
 * sg_gpu_topo — NCCL GPU topology table
 * Key: remote_ip (IPv4 of peer GPU host)
 * Value: nccl_topo_entry (domain, hop_count, bandwidth_gbps)
 * ----------------------------------------------------------------------- */
struct nccl_topo_key {
    __u32 remote_ip;
};

struct nccl_topo_entry {
    __u8  domain;           /* NCCL_TOPO_* constants */
    __u8  hop_count;        /* Number of network hops */
    __u16 bandwidth_gbps;   /* Effective bandwidth in Gbps */
    __u32 local_gpu_id;     /* Local GPU ordinal (0-based) */
    __u32 remote_gpu_id;    /* Remote GPU ordinal */
    __u32 nvlink_mask;      /* Bitmask of NVLink lanes available */
    __u64 last_updated_ns;  /* ktime_get_ns of last topology refresh */
};

struct {
    __uint(type, BPF_MAP_TYPE_HASH);
    __type(key,  struct nccl_topo_key);
    __type(value, struct nccl_topo_entry);
    __uint(max_entries, 4096);    /* Up to 4096 GPU peer hosts */
    __uint(pinning, LIBBPF_PIN_BY_NAME);
} sg_gpu_topo SEC(".maps");

/* -----------------------------------------------------------------------
 * sg_roce_events — Ring buffer for ECN / CNP telemetry events
 * ----------------------------------------------------------------------- */
struct roce_congestion_event {
    __u64 timestamp;    /* ktime_get_ns */
    __u32 src_ip;
    __u32 dst_ip;
    __u32 qpn;
    __u32 psn;
    __u8  ecn;
    __u8  opcode;
    __u16 pad;
};

struct {
    __uint(type, BPF_MAP_TYPE_RINGBUF);
    __uint(max_entries, 1 << 20);  /* 1 MiB ring buffer */
    __uint(pinning, LIBBPF_PIN_BY_NAME);
} sg_roce_events SEC(".maps");

/* -----------------------------------------------------------------------
 * sg_pfc_config — PFC per-priority configuration (indexed by 802.1p PCP)
 * ----------------------------------------------------------------------- */
struct pfc_config {
    __u8  enabled;          /* 1 = PFC pause enabled for this priority */
    __u8  dscp;             /* DSCP value to enforce */
    __u8  ecn_capable;      /* 1 = mark ECT(0) on egress */
    __u8  pad;
    __u32 pause_time_us;    /* PFC PAUSE quanta in microseconds */
    __u32 xon_threshold;    /* XON watermark in bytes */
    __u32 xoff_threshold;   /* XOFF watermark in bytes */
};

struct {
    __uint(type, BPF_MAP_TYPE_ARRAY);
    __type(key,  __u32);
    __type(value, struct pfc_config);
    __uint(max_entries, 8);   /* 8 PFC priority levels (802.1p) */
    __uint(pinning, LIBBPF_PIN_BY_NAME);
} sg_pfc_config SEC(".maps");

#endif /* __STRAITGATEWAY_MAPS_ROCE_H__ */
