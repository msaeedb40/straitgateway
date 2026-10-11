# Distributed Transit Gateway

StraitGateway turns each Kubernetes cluster into a distributed L3 Transit Router capable of multi-VRF isolation, enterprise WAN interconnection, and dynamic route exchange.

---

## 1. Multi-VRF Isolation

Virtual Routing and Forwarding (VRF) enables tenant separation and overlapping IP ranges within the same Kubernetes cluster.

```yaml
apiVersion: straitgateway.io/v1alpha1
kind: TransitGateway
metadata:
  name: tgw-primary
spec:
  asn: 65001
  routerId: "10.0.0.1"
  vrfs:
    - id: 100
      name: "tenant-finance"
      exportRouteTargets: ["65001:100"]
      importRouteTargets: ["65001:100"]
    - id: 200
      name: "tenant-analytics"
      exportRouteTargets: ["65001:200"]
      importRouteTargets: ["65001:200"]
```

* **VRF 0**: Default cluster domain (Kubernetes Pods, Services, and Nodes).
* **VRF 1-4096**: Tenant domains with dedicated LPM lookup tables inside the `sg_routes` BPF map.

---

## 2. Longest Prefix Match (LPM) Trie Engine

The forwarding information base (`FIBEngine` in `internal/routing/fib.go`) maintains routes keyed by `VRFID:Prefix`:

```c
struct lpm_key_v4 {
    __u32 prefixlen;     // 0 to 32
    __u32 vrf_id;        // Tenant VRF ID
    __u32 ip;            // Big-endian IPv4 address
};
```

### Route Precedence Hierarchy

When multiple routes overlap, StraitGateway selects the winning route using strict precedence:

| Precedence Rank | Constant | Value | Description |
|---|---|---|---|
| 1 (Highest) | `PrecedenceLocalPod` | 10 | Netkit device on current node |
| 2 | `PrecedenceClusterRoute` | 20 | WireGuard peer tunnel to remote cluster/node |
| 3 | `PrecedenceStaticRoute` | 30 | Admin-configured `TransitRoute` CR |
| 4 | `PrecedenceBGPRoute` | 40 | Dynamic route learned via RFC 4271 BGP |
| 5 (Lowest) | `PrecedenceDefaultRoute`| 100 | Default gateway (`0.0.0.0/0`) |
