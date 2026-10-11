// SPDX-License-Identifier: GPL-2.0
/*
 * bpf/roce/headers.h — RoCEv2 protocol header definitions
 *
 * RoCEv2 (RDMA over Converged Ethernet v2) encapsulates InfiniBand Transport
 * Protocol (IB-TL) packets inside UDP/IP. The XDP fast-path parser uses these
 * structs to:
 *   1. Identify RoCEv2 flows by UDP dst-port 4791
 *   2. Parse BTH (Base Transport Header) to extract QPN and OpCode
 *   3. Parse AETH (ACK Extended Transport Header) on RC ACK packets
 *   4. Apply PFC (Priority Flow Control) queue class marking
 *   5. Set ECN (Explicit Congestion Notification) CE bits for DCQCN
 *
 * Reference: IB Spec Vol 1 §9.7; RoCE Annex 16A; RFC 8200 (ECN).
 */

#ifndef __STRAITGATEWAY_ROCE_HEADERS_H__
#define __STRAITGATEWAY_ROCE_HEADERS_H__

#include "../include/common.h"

/* RoCEv2 is UDP port 4791 (IANA-assigned) */
#define ROCE_UDP_PORT       4791

/* -----------------------------------------------------------------------
 * BTH — Base Transport Header (12 bytes, mandatory on all RoCEv2 packets)
 * ----------------------------------------------------------------------- */
struct bth {
    __u8  opcode;       /* OpCode: 0x00=SEND_FIRST, 0x04=RDMA_WRITE_FIRST, ... */
    __u8  se_m_padcnt_tver; /* SolicEvent[7], MigReq[6], PadCnt[5:4], TVer[3:0] */
    __u16 p_key;        /* Partition Key */
    __u8  resv8;
    __u8  dst_qp[3];    /* Destination QP Number (24-bit, big-endian) */
    __u8  ack_f_resv;   /* ACK Req[7], RESV[6:0] */
    __u8  psn[3];       /* Packet Sequence Number (24-bit, big-endian) */
} __attribute__((packed));

/* BTH OpCode classes (upper 5 bits determine transport type) */
#define BTH_OPCODE_RC_SEND_FIRST        0x00
#define BTH_OPCODE_RC_SEND_MIDDLE       0x01
#define BTH_OPCODE_RC_SEND_LAST         0x02
#define BTH_OPCODE_RC_SEND_ONLY         0x04
#define BTH_OPCODE_RC_RDMA_WRITE_FIRST  0x06
#define BTH_OPCODE_RC_RDMA_WRITE_MID    0x07
#define BTH_OPCODE_RC_RDMA_WRITE_LAST   0x08
#define BTH_OPCODE_RC_RDMA_WRITE_ONLY   0x0A
#define BTH_OPCODE_RC_RDMA_READ_REQ     0x0C
#define BTH_OPCODE_RC_RDMA_READ_RESP    0x10
#define BTH_OPCODE_RC_ACK               0x11  /* RC Acknowledge */
#define BTH_OPCODE_RC_ATOMIC_ACK        0x12
#define BTH_OPCODE_RC_CMP_SWAP          0x13
#define BTH_OPCODE_RC_FETCH_ADD         0x14
#define BTH_OPCODE_UD_SEND_ONLY         0x64
#define BTH_OPCODE_UD_SEND_ONLY_IMM     0x65

/* -----------------------------------------------------------------------
 * AETH — ACK Extended Transport Header (4 bytes, follows BTH on RC ACK)
 * ----------------------------------------------------------------------- */
struct aeth {
    __u8  syndrome;     /* [7]=MSN valid, [6:5]=credit type, [4:0]=syndrome/credits */
    __u8  msn[3];       /* Message Sequence Number (24-bit) */
} __attribute__((packed));

/* AETH syndrome field decoding */
#define AETH_SYNDROME_ACK    0x00  /* Pure ACK */
#define AETH_SYNDROME_RNR    0x20  /* Receiver-Not-Ready */
#define AETH_SYNDROME_NAK    0x60  /* NAK (sequence error, etc.) */

/* -----------------------------------------------------------------------
 * RETH — RDMA Extended Transport Header (16 bytes, RDMA WRITE/READ)
 * ----------------------------------------------------------------------- */
struct reth {
    __u64 va;       /* Virtual Address */
    __u32 rkey;     /* Remote Key */
    __u32 dma_len;  /* DMA Length */
} __attribute__((packed));

/* -----------------------------------------------------------------------
 * DETH — Datagram Extended Transport Header (8 bytes, UD transport)
 * ----------------------------------------------------------------------- */
struct deth {
    __u32 q_key;    /* Queue Key */
    __u8  resv8;
    __u8  src_qp[3]; /* Source QPN (24-bit) */
} __attribute__((packed));

/* -----------------------------------------------------------------------
 * IPv6 GRH — Global Route Header (40 bytes, used for RoCEv1 / iWARP)
 * Included here for completeness in parsing path.
 * ----------------------------------------------------------------------- */
struct grh {
    __u32 ip_ver_tc_fl; /* [31:28]=Ver(6), [27:20]=TC, [19:0]=FlowLabel */
    __u16 payload_len;
    __u8  next_hdr;
    __u8  hop_limit;
    __u8  sgid[16];
    __u8  dgid[16];
} __attribute__((packed));

/* -----------------------------------------------------------------------
 * ECN / DCQCN — IP TOS / DSCP field manipulation helpers
 * ----------------------------------------------------------------------- */

/* ECN field values in the lower 2 bits of IP ToS */
#define ECN_NOT_ECT  0x00   /* Non-ECN-capable transport */
#define ECN_ECT1     0x01   /* ECN-capable transport (1) */
#define ECN_ECT0     0x02   /* ECN-capable transport (0) */
#define ECN_CE       0x03   /* Congestion Experienced */

/* PFC IEEE 802.1p priority → DSCP mapping table index */
#define PFC_PRIORITY_ROCE    3   /* PCP=3 for RoCEv2 traffic class */
#define PFC_DSCP_ROCE        26  /* DSCP AF31 = 26 for RoCEv2 by convention */

/* -----------------------------------------------------------------------
 * QP state encoding in sg_qp_table BPF map
 * ----------------------------------------------------------------------- */
#define QP_STATE_RESET       0
#define QP_STATE_INIT        1
#define QP_STATE_RTR         2  /* Ready To Receive */
#define QP_STATE_RTS         3  /* Ready To Send */
#define QP_STATE_SQD         4  /* Send Queue Drain */
#define QP_STATE_SQE         5  /* Send Queue Error */
#define QP_STATE_ERR         6  /* Error */
#define QP_STATE_DRAIN       7  /* Draining (SG internal) */

/* -----------------------------------------------------------------------
 * NCCL Topology domains — used by sg_gpu_topo BPF map
 * ----------------------------------------------------------------------- */
#define NCCL_TOPO_SAME_GPU   0   /* Same GPU (shared memory) */
#define NCCL_TOPO_NVLINK     1   /* NVLink peer */
#define NCCL_TOPO_PCIE       2   /* Same PCIe root complex */
#define NCCL_TOPO_SWITCH     3   /* Same ToR switch */
#define NCCL_TOPO_REMOTE     4   /* Cross-rack / cross-switch */

#endif /* __STRAITGATEWAY_ROCE_HEADERS_H__ */
