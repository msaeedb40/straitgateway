# StraitGateway Documentation

Welcome to the **StraitGateway** documentation hub. StraitGateway is an eBPF-native Kubernetes networking suite providing next-generation CNI, high-performance kube-proxy replacement, Gateway API v1.6.1 compliance, multi-cluster Transit Gateway mesh networking, and identity-aware security policies.

---

## Documentation Map

Explore the documentation sections below based on your role and operational requirements:

| Section | Description | Target Audience |
| :--- | :--- | :--- |
| **[Overview](overview.md)** | Core vision, problems solved, high-level features, and design philosophy. | All / Evaluators |
| **[Architecture](architecture.md)** | Deep dive into 10 control plane reconcilers (`sg-controller`), node daemon (`straitgatewayd`), CLI (`sg-cli`), 15-section Angular UI (`straitgateway-ui`), and the Intermediate Representation (IR) compiler pipeline. | Architects & Engineers |
| **[Capabilities](capabilities.md)** | Full feature breakdown: NetKit CNI & dynamic IPAM, Maglev service load balancing, DSR, XDP NodePort acceleration, BGP/BFD dynamic routing, and observability. | Network Engineers & SREs |
| **[Security](security.md)** | Identity-based policy enforcement (`StraitNetworkPolicy`), Linux Security Module (LSM) BPF hooks, default-deny postures, and WireGuard/IPsec encryption. | Security Engineers & SecOps |
| **[Transit Gateway](transit-gateway.md)** | Multi-cluster mesh and hub-and-spoke topologies, 32-bit segment isolation (`TransitSegment`), attachments, and automated WireGuard cross-cluster routing. | Cloud & Multi-cluster Architects |
| **[Gateway API](gateway-api.md)** | Gateway API v1.6.1 controller (`straitgateway.io/skgateway`), GatewayClass `skgateway`, and complete routing specifications for `HTTPRoute`, `TCPRoute`, `UDPRoute`, `TLSRoute`, and `GRPCRoute`. | Platform & Ingress Engineers |
| **[Installation & Operations Guide](guide.md)** | Step-by-step guides for installing, upgrading, and uninstalling StraitGateway via development scripts (Kind, Minikube, K3s, Kubeadm), production Helm charts, and operational management with `sg-cli`. | DevOps & Platform Operators |

---

## Fast Links & Quick References

- **Gateway API Controller**: `straitgateway.io/skgateway`
- **Managed GatewayClass**: `skgateway`
- **Helm Chart Directory**: [`straitgateway-helm/`](../straitgateway-helm)
- **API Group**: `straitgateway.io/v1alpha1`
- **Custom Resources (CRDs)**:
  - Multi-Cluster Transit: `TransitGateway`, `TransitSegment`, `TransitSegmentAttachment`, `TransitSegmentRoute`
  - Security & Policy: `StraitNetworkPolicy`
  - Node & Cluster State: `NodeNetworkConfig` (`nnc`), `ClusterNetworkConfig` (`cnc`)
  - Dynamic Routing: `BGPPeer` (`bgpp`), `BFDSession` (`bfd`)
- **Supported Kubernetes Versions**: `>= v1.34.0`
- **Required Linux Kernel**: `>= 6.7.0` (NetKit container datapath, TCX, and BPF LSM support)
- **CLI Utility**: `sg-cli` (17 subcommands for status, BPF maps, transit, nodes, policies, routes)
- **Dashboard UI**: `straitgateway-ui` (Angular 15-module management console)
