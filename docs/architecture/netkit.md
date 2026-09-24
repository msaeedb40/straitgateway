# Netkit Pod Networking Architecture

## Why Netkit Replaces veth
Traditional Kubernetes CNIs rely on Linux `veth` (virtual ethernet) pairs. Each packet transmitted across a veth pair incurs:
- Two traversals of the Linux networking stack (`netif_rx` / `dev_queue_xmit`)
- Queue overhead and context switching
- Double packet checksumming

Netkit (introduced in Linux 6.7) provides a dedicated, lightweight pod-to-node link mechanism specifically optimized for BPF:
- Operates at L3/L2 with BPF programs attached directly at the device boundary
- Direct packet redirection between container namespace and host datapath
- Up to 40% reduction in packet processing latency compared to veth pairs

## Kernel Compatibility & Fallback
- For Linux kernels ≥ 6.7: Native Netkit devices (`nk-host` / `nk-pod`) are created.
- For Linux kernels < 6.7: StraitGateway automatically detects missing Netkit support and creates standard veth pairs with TCX/TC attachments.
