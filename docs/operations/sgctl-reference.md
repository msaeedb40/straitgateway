# Operator CLI Reference (`sgctl`)

`sgctl` is the diagnostic and operational command-line utility for StraitGateway. It communicates directly with local node eBPF maps or remote node agent HTTP endpoints.

---

## 1. Global Flags

| Flag | Default | Description |
|---|---|---|
| `--endpoint` | `http://127.0.0.1:9090` | StraitGateway agent observability endpoint |
| `-h`, `--help` | N/A | Show help for command |

---

## 2. Command Catalog

### 2.1. `sgctl version`
Displays client version, git commit, build date, and compiler information:
```bash
sgctl version
# Output: StraitGateway v1.0.0 (commit: 8afee1a, built: 2026-10-08T07:05:58Z, go: go1.27.1, os/arch: linux/amd64)
```

### 2.2. `sgctl health`
Checks kernel compatibility, BPF filesystem mounts, and pinned map statuses:
```bash
sgctl health
# Output: Status: HEALTHY (Kernel: 6.12.0-generic, BPFFS: true, Maps: true)
```

### 2.3. `sgctl bpf maps list` & `dump`
Inspects pinned maps in `/sys/fs/bpf/straitgateway/`:
```bash
sgctl bpf maps list
sgctl bpf maps dump sg_routes
```

### 2.4. `sgctl routes list`
Lists active L3 forwarding entries across all VRFs:
```bash
sgctl routes list
# Output:
# VRF    DESTINATION        NEXT_HOP         INTERFACE    METRIC  
# 0      10.244.0.0/16      local            netkit       10      
# 0      10.96.0.0/12       service-vip      bpf-lb       10      
# 0      0.0.0.0/0          192.168.1.1      eth0         100
```

### 2.5. `sgctl peers status`
Shows BGP neighbor sessions and WireGuard tunnel links:
```bash
sgctl peers status
```

### 2.6. `sgctl conntrack dump`
Dumps active 5-tuple bidirectional connection tracking records:
```bash
sgctl conntrack dump
```

### 2.7. `sgctl policy list`
Inspects active identity-to-identity ACL rules in `sg_policies`:
```bash
sgctl policy list
```

### 2.8. `sgctl identity list`
Displays registered workload 32-bit identity tokens:
```bash
sgctl identity list
```

### 2.9. `sgctl service list`
Lists active Kubernetes Service VIPs and resolved backend endpoints:
```bash
sgctl service list
```

### 2.10. `sgctl metrics`
Queries dataplane metrics from the local agent:
```bash
sgctl metrics
```

### 2.11. `sgctl flows [--limit N]`
Streams real-time 5-tuple flow records from `sg_events`:
```bash
sgctl flows --limit 50
```

### 2.12. `sgctl clustermesh status` & `services`
Inspects multi-cluster transit federation status and global service mesh catalog:
```bash
sgctl clustermesh status
sgctl clustermesh services
```
