# StraitGateway: Security Architecture & Policies

StraitGateway implements an identity-based zero-trust security framework. By shifting policy enforcement from IP-based access control lists (ACLs) to cryptographic and cryptographic-grade numeric identities evaluated in eBPF, StraitGateway eliminates IP-reuse vulnerabilities and enforces strict least-privilege security across single- and multi-cluster environments.

---

## Identity-Based Policy Engine

Traditional firewall and iptables policies match against dynamic Pod IP addresses. In elastic Kubernetes clusters with frequent pod restarts and auto-scaling, IP churn leads to race conditions where stale rules temporarily allow unauthorized traffic.

### How StraitGateway Identities Work
1. **Label Hash Allocation**: When a pod starts, `sg-controller` hashes its immutable metadata labels and namespace into a unique 32-bit numeric security **Identity** (`sgtypes.Identity`).
2. **Cluster-Wide Identity Map**: The identity is pushed to the node's eBPF `identity_map` linking pod IP -> Identity.
3. **In-Packet Identification**: Network packets carry or resolve to this identity at ingress and egress hooks.
4. **O(1) Policy Lookups**: Enforcement rules are stored in `policy_map` as `(SrcIdentity, DstIdentity, Port, Proto) -> Action`. Lookups occur in constant time regardless of how many pods share that identity.

---

## The `StraitNetworkPolicy` CRD

StraitGateway provides advanced security policy capabilities beyond standard Kubernetes `NetworkPolicy` resources via the `StraitNetworkPolicy` custom resource (`straitgateway.io/v1alpha1`).

### Multi-Dimensional Selectors
A single `StraitNetworkPolicy` can select endpoints across multiple operational dimensions:

- **`namespaceSelector`**: Standard label selector matching namespaces.
- **`podSelector`**: Standard label selector matching pods within namespaces.
- **`clusterSelector`**: Matches remote clusters in a multi-cluster transit mesh.
- **`segmentSelector`**: Restricts traffic to specific 32-bit network segments (`TransitSegment`).
- **`gatewaySelector`**: Matches traffic originating from or targeting specific Gateway API Gateways.
- **`httprouteSelector`**: Binds security rules directly to specific HTTPRoute resources.
- **`grpcrouteSelector`**: Binds security rules directly to specific GRPCRoute resources.

### Policy Actions & Protocols
- **Actions**:
  - `Allow`: Permits matched traffic.
  - `Deny`: Silently drops packets (stealth mode).
  - `Reject`: Actively terminates connection by sending TCP RST or ICMP Port Unreachable.
- **Directions**: `Ingress` and `Egress`.
- **Protocols**: `TCP`, `UDP`, `ICMP`, `SCTP`, `Any`.

---

## Example: Microservice & Transit Segmentation Policy

The following policy allows `frontend` pods in namespace `production` to communicate with `payment` service backends, while strictly denying cross-segment traffic and blocking access from untrusted clusters:

```yaml
apiVersion: straitgateway.io/v1alpha1
kind: StraitNetworkPolicy
metadata:
  name: secure-payment-policy
  namespace: production
spec:
  # Target pods protected by this policy
  podSelector:
    matchLabels:
      app.kubernetes.io/name: payment-service

  ingress:
    # Rule 1: Allow frontend pods in production namespace over TCP port 8443
    - action: Allow
      direction: Ingress
      protocol: TCP
      ports:
        - 8443
      from:
        - namespaceSelector:
            matchLabels:
              kubernetes.io/metadata.name: production
          podSelector:
            matchLabels:
              app.kubernetes.io/name: frontend

    # Rule 2: Allow authorized payments cluster in transit mesh on secure segment 100
    - action: Allow
      direction: Ingress
      protocol: TCP
      ports:
        - 8443
      from:
        - clusterSelector:
            matchLabels:
              region: eu-west-1
              environment: prod
          segmentSelector:
            matchLabels:
              segment: payment-pci-dss

  egress:
    # Rule 3: Allow payment service to reach database on segment 100
    - action: Allow
      direction: Egress
      protocol: TCP
      ports:
        - 5432
      to:
        - podSelector:
            matchLabels:
              app.kubernetes.io/name: payment-db
```

---

## Linux Security Module (LSM) BPF Enforcement

In addition to packet-level filtering on network interfaces, StraitGateway utilizes **BPF LSM (Linux Security Module)** hooks directly within the kernel socket layer:

- **`bpf_lsm_socket_connect`**: Evaluates egress network policies at the exact moment a process calls `connect()`. If the connection violates policy, the syscall returns `-EPERM` immediately before any SYN packet is constructed or placed on the wire.
- **`bpf_lsm_socket_bind`**: Restricts which ports and IP addresses containers can bind to, preventing unauthorized port interception or privilege escalation inside compromised pods.

---

## Transparent Overlay & Transit Encryption

StraitGateway provides native, kernel-level encryption for pod-to-pod, node-to-node, and cross-cluster transit traffic:

1. **WireGuard (Default)**:
   - Built into the Linux kernel using the state-of-the-art **ChaCha20-Poly1305** cipher suite.
   - Public/private keys are rotated and distributed automatically by `sg-controller`.
   - Offers near-line-rate encryption throughput with minimal CPU overhead.
2. **IPsec**:
   - Optional enterprise cipher suite with hardware crypto acceleration (AES-GCM-256).
   - Designed for compliance with strict regulatory mandates requiring FIPS-validated IPsec tunnels.
3. **Node-to-Node Encryption**:
   - When enabled via `encryption.nodeEncryption=true`, all host-level traffic between Kubernetes nodes is encrypted, safeguarding control-plane and host-networked communications.

---

## Default Deny Architecture

By configuring `networkPolicy.defaultDeny=true` in Helm:
- All ingress and egress traffic is dropped by default across the entire cluster.
- Pods are fully isolated upon creation until an explicit `StraitNetworkPolicy` or standard Kubernetes `NetworkPolicy` allows the flow.
- Health checks (Kubelet liveness and readiness probes) are automatically whitelisted via kernel identity rules to maintain pod lifecycle stability.
