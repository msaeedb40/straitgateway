# StraitGateway CNI Architecture

## Overview
`strait-cni` is a standalone, fully conforming CNI plugin implementing the Container Network Interface specification v1.1.0+. It operates without Calico, Flannel, or Cilium dependencies.

## CNI Operations
- **ADD**:
  1. Reads `NetConf` from standard input.
  2. Allocates IP from per-node cluster-pool IPAM without RFC1918 assumptions.
  3. Creates Netkit link pair (`nk-<pod>` on host, `eth0` in pod netns).
  4. Moves peer into container netns and configures IP address + default route.
  5. Attaches eBPF TCX egress and ingress programs.
  6. Signals `straitd` to initialize pod flow tracking and policy rules.
- **DEL**:
  1. Idempotently cleans up container routes, eBPF attachments, and Netkit interfaces.
  2. Releases IP back to IPAM allocator.
- **CHECK**: Validates interface health, IP assignment, and eBPF link attachment.
- **GC**: Sweeps and cleans up orphaned or dangling interface artifacts after abnormal container crashes.
- **STATUS**: Reports local node readiness, datapath availability, and IP pool utilization.
