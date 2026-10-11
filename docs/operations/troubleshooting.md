# Troubleshooting & Diagnostics

This guide provides operational procedures for diagnosing and resolving issues in StraitGateway.

---

## 1. Quick Diagnostic Runbook

Run the following initial diagnostic checks using `sgctl`:

```bash
# 1. Check node health and BPF filesystem
sgctl health

# 2. Verify all 13 pinned maps exist
sgctl bpf maps list

# 3. Check BGP and WireGuard peer status
sgctl peers status

# 4. Check whether traffic is being dropped by security policies
sgctl policy list
sgctl flows --limit 20
```

---

## 2. Common Issues & Solutions

### 2.1. eBPF Verifier Rejection
* **Symptoms**: Agent fails to start with log `bpf program load: permission denied` or verifier complexity limit exceeded.
* **Root Cause**: Kernel missing BTF (`/sys/kernel/btf/vmlinux`) or kernel version < 6.8.
* **Remediation**:
  1. Verify kernel version: `uname -r`.
  2. Verify BTF support: `ls -l /sys/kernel/btf/vmlinux`.
  3. Ensure `CONFIG_DEBUG_INFO_BTF=y` in kernel config.

### 2.2. Service VIP Connection Timeout
* **Symptoms**: `curl 10.96.x.x:80` hangs or returns `Connection timed out`.
* **Remediation**:
  1. Inspect whether the service VIP exists in BPF maps:
     ```bash
     sgctl service list
     ```
  2. If endpoints are listed as `0`, check if backend pods are ready (`kubectl get pods -l <selector>`).
  3. Check if conntrack entries are exhausted:
     ```bash
     sgctl conntrack dump
     ```

### 2.3. Multi-Cluster WireGuard Link Down
* **Symptoms**: Packets to remote cluster PodCIDR (`10.201.0.0/16`) fail to route.
* **Remediation**:
  1. Check WireGuard peer status:
     ```bash
     sgctl peers status
     ```
  2. Test UDP reachability on port 51820 to remote node:
     ```bash
     nc -zvu <remote-node-ip> 51820
     ```
  3. Verify `ClusterPeering` CR status:
     ```bash
     kubectl describe clusterpeering <peer-name>
     ```

### 2.4. Policy Deny Drops
* **Symptoms**: Workload communication suddenly fails with packet drop logs.
* **Remediation**:
  1. Stream real-time flow logs:
     ```bash
     sgctl flows
     ```
  2. Look for `DENY` verdicts. Check source identity and destination identity.
  3. Add matching ingress/egress rules in the corresponding `SecurityPolicy` resource.
