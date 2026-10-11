# WireGuard Overlay Mesh

StraitGateway provides an automated kernel-native WireGuard mesh (`wg0`) that connects nodes across multiple Kubernetes clusters, multi-cloud regions, and on-premises datacenters.

---

## 1. Mesh Architecture

Unlike user-space VPN tunnels, StraitGateway creates and manages kernel WireGuard interfaces directly using `netlink` and `wgctrl`:

```
+---------------------------------------------------------------+
|                      StraitGateway Mesh                       |
|                                                               |
|  [Cluster-East Node 1]                     [Cluster-West Node 1]
|   - WireGuard IP: 10.200.0.1                - WireGuard IP: 10.200.0.2
|   - Interface: wg0                          - Interface: wg0  |
|            \                                     /            |
|             \== Encrypted ChaCha20-Poly1305 ====/             |
|              (Kernel wg0 Fast Path UDP 51820)                 |
+---------------------------------------------------------------+
```

---

## 2. Configuration & Discovery (`api/v1alpha1/peering_types.go`)

### 2.1. Standalone Peer Definition

```yaml
apiVersion: straitgateway.io/v1alpha1
kind: GatewayPeer
metadata:
  name: peer-remote-dc
spec:
  transitGatewayRef: "tgw-primary"
  peerType: "WireGuard"
  peerIp: "192.168.20.1"
  wireguardPeer:
    publicKey: "xT9vLqG5uR8fO2eY1wK3mZ7sA4bC6dE8hJ0kL2n="
    endpoint: "203.0.113.10:51820"
    allowedIPs:
      - "10.244.2.0/24" # Remote Pod CIDR
      - "10.96.2.0/24"  # Remote Service CIDR
    persistentKeepaliveSeconds: 25
```

### 2.2. Multi-Cluster Peering (`ClusterPeering`)

```yaml
apiVersion: straitgateway.io/v1alpha1
kind: ClusterPeering
metadata:
  name: peering-aws-us-east
spec:
  clusterId: "aws-us-east-1"
  wireguardEndpoint: "54.210.12.34:51820"
  wireguardPublicKey: "B9q5fA71kL2nM3oP4rS5tU6vW7xY8z0123456789abc="
  podCIDRs:
    - "10.201.0.0/16"
  serviceCIDRs:
    - "10.101.0.0/16"
```

---

## 3. Kernel Integration with BPF Dataplane

1. **Direct Routing**:
   * Remote PodCIDRs are installed into `sg_routes` pointing to the `wg0` interface index.
2. **Hardware Acceleration**:
   * WireGuard encryption utilizes ChaCha20-Poly1305 with AVX-512 and ARM NEON hardware acceleration in the Linux kernel.
3. **Zero Copy**:
   * In-kernel routing routes directly from `netkit` to `wg0` without userspace buffers or daemon context switches.
