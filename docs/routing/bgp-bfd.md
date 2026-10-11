# Dynamic Routing with BGP & BFD

StraitGateway features an embedded RFC 4271 BGP-4 speaker and an RFC 5880 Bidirectional Forwarding Detection (BFD) monitor for sub-second failure detection.

---

## 1. Embedded BGP Speaker (RFC 4271)

Implemented in `internal/routing/bgp/speaker.go`, the BGP speaker communicates natively with upstream Top-of-Rack (ToR) switches, cloud transit gateways, and edge routers.

### 1.1. Peering Configuration (`GatewayPeer` CRD)

```yaml
apiVersion: straitgateway.io/v1alpha1
kind: GatewayPeer
metadata:
  name: peer-tor-1
spec:
  transitGatewayRef: "tgw-primary"
  peerType: "BGP"
  peerIp: "192.168.10.1"
  peerAsn: 65002
  holdTimeSeconds: 90
  keepaliveSeconds: 30
  enableBfd: true
  advertisedPrefixes:
    - "10.244.0.0/16" # Pod CIDR
    - "10.96.0.0/12"  # Service VIPs
```

### 1.2. Prefix Advertisements & Withdrawals
* **Pod CIDRs**: Advertised upon node agent startup.
* **Service VIPs**: Advertised dynamically when Services with `type: LoadBalancer` or Transit annotations are created.
* **Learned Routes**: Injected directly into the `sg_routes` BPF map for immediate in-kernel forwarding.

---

## 2. Sub-50ms BFD Failure Detection (RFC 5880)

Traditional BGP sessions take 90 seconds (hold time expiry) to detect a silent link drop or dead upstream switch. StraitGateway eliminates this failover delay using BFD:

```
[StraitGateway BFD]  <--- 50ms Echo Heartbeats --->  [Upstream ToR Switch]
         |
    Link Lost (3 missed heartbeats = 150ms)
         |
         v
[Immediate BPF Invalidation] (0ms reconciliation delay)
         |
         v
Retract next-hop from sg_routes -> Fallback to redundant peer in < 50ms
```

* **Heartbeat Interval**: Configured down to 50ms asynchronous mode.
* **Failover Convergence**: BFD timeout triggers an atomic retraction of affected routes from the kernel's `sg_routes` map directly, preventing black-holing of cluster traffic.
