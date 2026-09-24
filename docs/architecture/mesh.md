# StraitGateway Sidecarless Service Mesh Architecture

## Overview
StraitGateway delivers service mesh capabilities—including intelligent traffic routing, L4/L7 policy enforcement, and observability—without running a sidecar proxy inside each workload pod.

## Sidecar vs Sidecarless Comparison
Traditional sidecar meshes (e.g. Envoy sidecars) inject a proxy container per pod, causing:
- 2x socket and context switches per connection.
- High memory footprint per pod (50MB+ overhead per container).
- Lifecycle coupling and deployment ordering complexities.

StraitGateway replaces sidecars by intercepting traffic at the kernel socket and Netkit layers:
```
Workload Socket (connect)
       │
       ▼
eBPF Socket LB (cgroup connect4)  ──> Translates VIP to Endpoint IP
       │
       ▼
Pod Netkit Interface
       │
       ▼
Node TCX Hook                     ──> Enforces Policy & Collects Telemetry
       │
       ▼
Target Pod (direct memory fast-path)
```

## Key Capabilities
1. **Zero-Copy Forwarding**: Netkit links operate in passthrough mode, moving skbs between network namespaces with minimal overhead.
2. **Gateway API Support**: Integration with Kubernetes Gateway API (HTTPRoute, GRPCRoute, TCPRoute) translated directly into datapath map states.
3. **Mutual TLS (mTLS)**: WireGuard/IPsec kernel encryption or transparent node-level TLS terminators without per-pod proxies.
4. **Kernel Flow Visibility**: Flow state exported via BPF ring buffers directly to OpenTelemetry collectors.
