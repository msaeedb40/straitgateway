# Linux Capabilities & Kernel Compatibility

## Overview
StraitGateway operates close to the Linux kernel to leverage Netkit and eBPF. This document details privilege requirements and capability detection invariants.

## Required Linux Capabilities
StraitGateway daemons require specific capabilities rather than full privileged root access:

| Capability | Daemon | Reason |
| :--- | :--- | :--- |
| `CAP_NET_ADMIN` | `straitd`, `strait-cni`, `tgwd` | Creating and configuring Netkit devices, IP addresses, TCX attachments, and routes |
| `CAP_SYS_ADMIN` | `straitd` | Legacy BPF loading on older kernels, mounting `/sys/fs/bpf` |
| `CAP_BPF` | `straitd` | Loading eBPF programs, creating BPF maps (Linux 5.8+) |
| `CAP_PERFMON` | `straitd` | Accessing performance counters and tracepoints |
| `CAP_NET_RAW` | `sgpktcap` | Packet capture and raw socket inspection |

## Kernel Compatibility Matrix
StraitGateway uses capability detection at startup to verify feature availability before attaching programs:

| Feature | Minimum Kernel | Recommended Kernel | Fallback Strategy |
| :--- | :--- | :--- | :--- |
| **Netkit** | Linux 6.7+ | Linux 6.8+ | Fall back to optimized veth pairs if Netkit module absent |
| **TCX Hooks** | Linux 6.6+ | Linux 6.8+ | Fall back to TC BPF (`tc filter`) |
| **cgroup Socket LB** | Linux 5.4+ | Linux 6.1+ | Fall back to TCX-level DNAT |
| **BPF Ring Buffer** | Linux 5.8+ | Linux 6.1+ | Fall back to perf event buffers |
| **BPF LSM** | Linux 5.7+ | Linux 6.1+ | Disabled with audit warning |
