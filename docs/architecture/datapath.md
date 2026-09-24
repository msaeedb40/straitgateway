# StraitGateway eBPF Datapath Architecture

## Overview
StraitGateway bypasses iptables, nftables, and IPVS entirely, implementing an eBPF-first datapath directly attached to network interfaces and Netkit pod devices.

## Hook Points & BPF Programs
1. **XDP Ingress (`bpf/programs/xdp/sg_xdp.bpf.c`)**:
   - Ingress fast-path filtering and DDoS mitigation at driver level.
   - Falls back to generic XDP when driver support is absent.
2. **TCX Egress/Ingress (`bpf/programs/tcx/sg_tcx.bpf.c`)**:
   - Primary Netkit pod link attachment.
   - Evaluates security policies and routes packets to node or pod endpoints.
3. **Socket LB (`bpf/programs/socket/sg_sock_lb.bpf.c`)**:
   - Attaches to cgroup v2 hooks (`connect4/6`, `sendmsg4/6`).
   - Translates Service VIPs to backend Pod IPs before TCP/UDP handshake, eliminating DNAT overhead in the packet path.
4. **LSM Security (`bpf/programs/lsm/sg_lsm.bpf.c`)**:
   - Enforces workload network isolation at the kernel socket layer.

## BPF Maps
- `services`: VIP to backend pod endpoint mapping.
- `routes`: LPM trie routing table.
- `policy`: Identity-based policy rules.
- `flow`: LRU flow connection tracking and ring buffer.
- `peers`: Transit gateway and mesh peer connections.
