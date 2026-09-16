# StraitGateway: Installation, Uninstallation, & Lifecycle Guide

This guide provides step-by-step instructions for deploying, managing, upgrading, and removing StraitGateway across local development environments (Kind, Minikube, K3s, Kubeadm) and production Kubernetes clusters via Helm.

---

## Prerequisites

Before installing StraitGateway, ensure your target nodes meet the following requirements:

1. **Linux Kernel**: Version `>= 6.7.0` with eBPF, TCX, and NetKit enabled.
   > [!IMPORTANT]
   > NetKit was introduced in Linux Kernel 6.7.0 to replace legacy `veth` pairs. `straitgatewayd` verifies `MinKernelVersion = "6.7.0"` at startup.
2. **Kernel Modules & Mounts**:
   - BPF filesystem mounted: `/sys/fs/bpf`
   - Linux Security Modules (BPF LSM) enabled (`CONFIG_BPF_LSM=y` and `lsm=...,bpf` in kernel boot parameters) for socket-level enforcement.
3. **Kubernetes Version**: `>= v1.34.0`.
4. **Tools**:
   - `kubectl` configured with cluster administrator privileges.
   - `helm` version `v3.12+`.
   - `clang-22` and `bpftool` (if compiling eBPF programs from source).

---

## 1. Quick Development Setup (Scripts)

StraitGateway includes automated provisioning and installation scripts under the `scripts/` directory for rapid local development.

### Kind
```bash
# Create Kind cluster, build images, and install StraitGateway
make kind-create
make kind-install

# Or directly execute scripts:
bash scripts/kind/create-cluster.sh
bash scripts/kind/install.sh
```

### Minikube
```bash
# Create Minikube profile and install StraitGateway
make minikube-create
make minikube-install

# Or directly execute:
bash scripts/minikube/create-cluster.sh
bash scripts/minikube/install.sh
```

### K3s
```bash
# Install StraitGateway into an existing K3s cluster (with Flannel disabled)
make k3s-install
# Or directly:
bash scripts/k3s/install.sh
```

### Kubeadm Multi-Node Cluster
```bash
# Multi-node kubeadm installation (kube-proxy skipped)
make kubeadm-install
# Or directly:
bash scripts/kubeadm/install.sh
```

---

## 2. Uninstall StraitGateway (Scripts)

To tear down StraitGateway installations created via local scripts:

### Kind
```bash
make kind-uninstall
# Or directly:
bash scripts/kind/uninstall.sh
```

### Minikube
```bash
bash scripts/minikube/uninstall.sh
```

### K3s
```bash
bash scripts/k3s/uninstall.sh
```

---

## 3. Install StraitGateway by Helm

For production environments, StraitGateway is packaged as an enterprise-ready Helm chart located at [`straitgateway-helm/`](../straitgateway-helm).

### Step 1: Install Gateway API v1.6.1 CRDs
StraitGateway requires the standard Kubernetes Gateway API CRDs:

```bash
kubectl apply --server-side -f https://github.com/kubernetes-sigs/gateway-api/releases/download/v1.6.1/standard-install.yaml
```

### Step 2: Install StraitGateway Chart
Deploy the chart into the dedicated `straitgateway-system` namespace:

```bash
helm install straitgateway ./straitgateway-helm \
  --namespace straitgateway-system \
  --create-namespace \
  --set global.clusterName="production-cluster-01" \
  --set global.clusterID=1 \
  --set kubeProxyReplacement.enabled=true \
  --set dataplane.overlay=Native \
  --wait --timeout=15m
```

### Key Helm Configuration Values

| Value | Default | Description |
| :--- | :--- | :--- |
| `global.clusterName` | `""` | Unique identifier for cluster in transit mesh |
| `global.clusterID` | `1` | Numeric cluster ID (required for transit mesh) |
| `cni.enabled` | `true` | Installs StraitGateway CNI plugin binary |
| `cni.disableDefaultCNI` | `true` | Disables pre-existing CNI plugins |
| `dataplane.overlay` | `Native` | Pod traffic routing (`Native`, `VXLAN`, `Geneve`) |
| `kubeProxyReplacement.enabled` | `true` | Replaces kube-proxy with eBPF service dataplane |
| `serviceLoadBalancer.algorithm` | `maglev` | LB algorithm (`maglev`, `roundrobin`, `leastconn`, `iphash`, `random`) |
| `serviceLoadBalancer.maglevTableSize` | `128` | Maglev consistent hash table size (must be prime) |
| `serviceLoadBalancer.dsr` | `false` | Enables Direct Server Return for backends |
| `serviceLoadBalancer.nodePort.enabled`| `true` | Accelerates NodePort handling via XDP |
| `networkPolicy.enabled` | `true` | Activates eBPF identity network policy engine |
| `networkPolicy.lsm` | `true` | Enables Linux Security Module socket hooks |
| `networkPolicy.defaultDeny` | `false` | Sets cluster-wide default-deny zero trust posture |
| `transitGateway.enabled` | `false` | Enables multi-cluster transit gateway mesh |
| `transitGateway.topology` | `Mesh` | Transit topology (`Mesh`, `HubAndSpoke`, `PeerToPeer`, `GatewayToGateway`) |
| `transitGateway.encryption` | `WireGuard` | Transit tunnel encryption (`WireGuard`, `IPsec`, `None`) |
| `bgp.enabled` | `false` | Activates BGP dynamic routing controller |
| `bgp.bfd` | `false` | Enables BFD sub-second link failure detection |
| `ui.enabled` | `false` | Deploys the Angular UI dashboard (accessible via ClusterIP) |

---

## 4. Uninstall StraitGateway by Helm

To remove StraitGateway cleanly from your cluster:

### Step 1: Uninstall the Helm Release
```bash
helm uninstall straitgateway --namespace straitgateway-system
```

### Step 2: Delete Namespace (Optional)
```bash
kubectl delete namespace straitgateway-system --ignore-not-found
```

### Step 3: Remove CRDs (Optional)
> [!WARNING]
> Deleting CRDs will delete all active `TransitGateway`, `TransitSegment`, `StraitNetworkPolicy`, `NodeNetworkConfig`, `ClusterNetworkConfig`, and `BFDSession` custom resources.

```bash
kubectl delete -f straitgateway-helm/charts/crds/
```

---

## 5. Update / Reinstall StraitGateway by Helm (`helm upgrade --install`)

To upgrade an existing installation, apply new configuration values, or reinstall idempotently, use `helm upgrade --install`. This guarantees atomic rollouts with zero downtime.

### Basic Upgrade / Reinstallation
```bash
helm upgrade --install straitgateway ./straitgateway-helm \
  --namespace straitgateway-system \
  --create-namespace \
  --set global.clusterName="production-cluster-01" \
  --set kubeProxyReplacement.enabled=true \
  --wait --timeout=15m
```

### Enabling Multi-Cluster Transit Gateway & WireGuard
To upgrade an existing cluster and enable multi-cluster transit mesh with WireGuard encryption:

```bash
helm upgrade --install straitgateway ./straitgateway-helm \
  --namespace straitgateway-system \
  --set global.clusterName="production-cluster-01" \
  --set global.clusterID=1 \
  --set transitGateway.enabled=true \
  --set transitGateway.topology="Mesh" \
  --set transitGateway.encryption="WireGuard" \
  --wait --timeout=15m
```

### Enabling the UI Dashboard
To enable and access the Angular management console:

```bash
helm upgrade --install straitgateway ./straitgateway-helm \
  --namespace straitgateway-system \
  --set ui.enabled=true \
  --wait --timeout=15m

# Forward UI port to local machine
kubectl port-forward -n straitgateway-system svc/straitgateway-ui 8080:80
```

---

## 6. Verification & Operational Health Checks

Once installed or updated, verify cluster health with `kubectl` and `sg-cli`:

```bash
# 1. Check pod status in straitgateway-system
kubectl get pods -n straitgateway-system -o wide

# 2. Check DaemonSet rollout
kubectl rollout status daemonset/straitgateway-agent -n straitgateway-system

# 3. Check controller deployment rollout
kubectl rollout status deployment/straitgateway-controller -n straitgateway-system

# 4. Verify ClusterNetworkConfig and NodeNetworkConfig CRDs
kubectl get cnc
kubectl get nnc -o wide

# 5. Verify GatewayClass controller registration
kubectl get gatewayclass skgateway -o yaml

# 6. Check eBPF agent logs
kubectl logs -n straitgateway-system -l app.kubernetes.io/name=straitgateway-agent -c straitgatewayd --tail=50

# 7. Check status via sg-cli
sg-cli status
sg-cli node list
sg-cli gateway list
```
