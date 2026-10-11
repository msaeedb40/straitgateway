# Kernel Compatibility & System Requirements

StraitGateway is engineered for modern high-performance Linux kernels. It makes native use of kernel innovations introduced in Linux 6.8 and Linux 6.12 LTS.

---

## 1. Supported Linux Kernels

| Kernel Version | Support Tier | Primary Features Utilized |
|---|---|---|
| **Linux 6.12+ LTS** | **Tier 1 (Recommended)** | `netkit` L3/L2 driver, TCX links (`BPF_PROG_TYPE_SCHED_CLS`), eBPF ringbuffer v2, kfuncs. |
| **Linux 6.8+** | **Tier 1 (Supported)** | Initial `netkit` support, TCX attachments, CO-RE BTF relocations. |
| **Linux 6.1 - 6.6 LTS** | **Tier 2 (Fallback)** | Fallback from `netkit` to TC on `veth`; standard TC BPF direct-action attachments. |
| **Linux < 6.1** | **Unsupported** | Missing required BTF types and socket hook capabilities. |

---

## 2. Kernel Configuration Prerequisites

The target Linux kernel must be compiled with the following `CONFIG_*` options enabled (built-in `=y` or module `=m`):

```ini
# eBPF Core & BTF
CONFIG_BPF=y
CONFIG_BPF_SYSCALL=y
CONFIG_BPF_JIT=y
CONFIG_BPF_JIT_ALWAYS_ON=y
CONFIG_DEBUG_INFO_BTF=y
CONFIG_BPF_LSM=y

# Networking & TCX
CONFIG_NET_CLS_ACT=y
CONFIG_NET_SCH_INGRESS=y
CONFIG_NET_CLS_BPF=y
CONFIG_BPF_STREAM_PARSER=y

# Netkit & Virtual Devices
CONFIG_NETKIT=m # or =y (Required for native netkit CNI)
CONFIG_WIREGUARD=m # or =y (Required for WireGuard mesh)

# Cgroup & Socket hooks
CONFIG_CGROUPS=y
CONFIG_CGROUP_BPF=y
CONFIG_SOCK_CGROUP_DATA=y
```

---

## 3. Host System Dependencies

* **BPF Filesystem**: Must be mounted at `/sys/fs/bpf` with `shared` propagation:
  ```bash
  mount --make-shared /sys/fs/bpf
  ```
* **Cgroup v2 Unified Hierarchy**: Mounted at `/sys/fs/cgroup`.
* **Kernel BTF Information**: Available at `/sys/kernel/btf/vmlinux`.

---

## 4. Recommended Sysctl Tunings

To support high-throughput packet processing and conntrack scaling:

```ini
# /etc/sysctl.d/99-straitgateway.conf

# Enable IP forwarding across namespaces and interfaces
net.ipv4.ip_forward = 1
net.ipv6.conf.all.forwarding = 1

# eBPF JIT tuning
net.core.bpf_jit_enable = 1
net.core.bpf_jit_harden = 2
net.core.bpf_jit_kallsyms = 1

# Network device queue backlogs
net.core.netdev_max_backlog = 16384
net.core.somaxconn = 32768

# Opt-in for UDP WireGuard socket buffering
net.core.rmem_max = 67108864
net.core.wmem_max = 67108864
```
