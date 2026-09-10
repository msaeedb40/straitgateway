# StraitGateway Documentation

Welcome to the **StraitGateway** documentation hub. StraitGateway is an eBPF-native Kubernetes networking suite providing next-generation CNI, high-performance kube-proxy replacement, Gateway API v1.6.1 compliance, multi-cluster Transit Gateway mesh networking, and identity-aware security policies.

---

## Documentation Map

Explore the documentation sections below based on your role and operational requirements:

| Section | Description | Target Audience |
| :--- | :--- | :--- |
| **[Overview](overview.md)** | Core vision, problems solved, high-level features, and design philosophy. | All / Evaluators |
| **[Architecture](architecture.md)** | Deep dive into control plane (`sg-controller`), node daemon (`straitgatewayd`), CLI (`sg-cli`), UI (`straitgateway-ui`), and the Intermediate Representation (IR) compiler pipeline. | Architects & Engineers |
| **[Capabilities](capabilities.md)** | Full feature breakdown: CNI & IPAM, Maglev service load balancing, DSR, XDP NodePort acceleration, BGP/BFD, and observability. | Network Engineers & SREs |
| **[Security](security.md)** | Identity-based policy enforcement (`StraitNetworkPolicy`), Linux Security Module (LSM) BPF hooks, default-deny postures, and WireGuard/IPsec encryption. | Security Engineers & SecOps |
| **[Transit Gateway](transit-gateway.md)** | Multi-cluster mesh and hub-and-spoke topologies, 32-bit segment isolation (`TransitSegment`), attachments, and automated WireGuard cross-cluster routing. | Cloud & Multi-cluster Architects |
| **[Gateway API](gateway-api.md)** | Gateway API v1.6.1 controller (`straitgateway.io/skgateway`), GatewayClass `skgateway`, and complete routing specifications for `HTTPRoute`, `TCPRoute`, `UDPRoute`, `TLSRoute`, and `GRPCRoute`. | Platform & Ingress Engineers |
| **[Installation & Operations Guide](guide.md)** | Step-by-step guides for installing, upgrading, and uninstalling StraitGateway via development scripts (Kind, Minikube, K3s, Kubeadm) and production Helm charts. | DevOps & Platform Operators |

---

## Fast Links & Quick References

- **Gateway API Controller**: `straitgateway.io/skgateway`
- **Managed GatewayClass**: `skgateway`
- **Helm Chart Directory**: [`straitgateway-helm/`](../straitgateway-helm)
- **API Group**: `straitgateway.io/v1alpha1`
- **Supported Kubernetes Versions**: `>= v1.34.0`
- **Required Linux Kernel**: `>= 5.15` (eBPF, TCX, NetKit, BPF LSM support)
