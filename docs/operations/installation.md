# Installation & Deployment Guide

StraitGateway can be deployed onto any modern Kubernetes cluster (Linux 6.8+ or 6.12+ LTS) using the official Helm chart or standalone manifests.

---

## 1. Prerequisites

1. **Linux Kernel**: Version 6.8+ or 6.12+ LTS on all worker and control plane nodes.
2. **BPF Mount**: Verify `/sys/fs/bpf` is mounted:
   ```bash
   mount | grep /sys/fs/bpf
   # If not mounted:
   mount -t bpf bpffs /sys/fs/bpf
   ```
3. **Cgroup v2**: Verify unified hierarchy at `/sys/fs/cgroup`.

---

## 2. Option A: Helm Deployment (Recommended)

```bash
# Add StraitGateway repository
helm repo add straitgateway https://charts.straitgateway.io
helm repo update

# Install StraitGateway in straitgateway-system namespace
helm install straitgateway charts/straitgateway \
  --namespace straitgateway-system \
  --create-namespace \
  --set bpf.mountPath="/sys/fs/bpf" \
  --set cni.mode="netkit" \
  --set wireguard.enabled=true \
  --set routing.bgp.localAsn=65001
```

### Key Values Configuration (`charts/straitgateway/values.yaml`)

| Parameter | Default | Description |
|---|---|---|
| `cni.mode` | `netkit` | CNI driver (`netkit` or `veth`) |
| `bpf.mountPath` | `/sys/fs/bpf` | Host BPF filesystem mount path |
| `wireguard.listenPort` | `51820` | Port for encrypted overlay mesh tunnels |
| `routing.bgp.routerId` | `""` | BGP Router ID (defaults to Node IP) |
| `metrics.port` | `9090` | Observability metrics and flow port |

---

## 3. Option B: Standalone All-In-One Manifest

Deploy all 18 core resources (Namespace, CRDs, RBAC, ConfigMap, Controller Deployment, Agent DaemonSet, and CNI DaemonSet) in a single command:

```bash
kubectl apply -f deploy/all-in-one.yaml
```

### Verify Deployment Status

```bash
kubectl get pods -n straitgateway-system -o wide
kubectl get crds | grep straitgateway
```
