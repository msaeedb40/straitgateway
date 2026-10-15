# Gateway API v1.6.3 Implementation

StraitGateway implements full conformance with the Kubernetes **Gateway API v1.6.3** specification, providing next-generation ingress, egress, and cross-cluster transit routing.

---

## 1. Supported Gateway API Resources

* `GatewayClass`: Identifies StraitGateway as the backing data plane controller (`straitgateway.io/gateway-class`).
* `Gateway`: Manages entry points, listening ports, protocol bindings, and TLS termination certificates.
* `HTTPRoute`: Declares layer 7 routing rules, header mutations, path rewrites, and traffic split percentages.
* `GRPCRoute`: High-performance gRPC method and service routing with header filtering.
* `TLSRoute`: Layer 4 SNI passthrough routing without terminating payload encryption.
* `TCPRoute` & `UDPRoute`: Direct Layer 4 raw socket routing and Maglev load balancing.

---

## 2. Advanced Traffic Routing Features

### 2.1. Weighted Canary Releases (`HTTPRoute`)

```yaml
apiVersion: gateway.networking.k8s.io/v1
kind: HTTPRoute
metadata:
  name: store-route
  namespace: ecommerce
spec:
  parentRefs:
    - name: egress-gateway
  rules:
    - matches:
        - path:
            type: PathPrefix
            value: /api/v2
      backendRefs:
        - name: store-v1
          port: 8080
          weight: 90
        - name: store-v2
          port: 8080
          weight: 10
```

### 2.2. Header Manipulation & Path Rewrites

* **Request Header Modification**: Injects tracing headers (`traceparent`), strips sensitive client cookies, and adds security tags.
* **Path Prefix Replacement**: Rewrites external URL namespaces to internal service microservice endpoints cleanly.

---

## 3. Controller Architecture & RBAC

The `sg-controller` reconciles Gateway API resources in real-time, programming eBPF maps and synchronizing routes with node agents (`straitd` / `sg-agent`):

* **RBAC Roles**: Full permissions configured for `gateway.networking.k8s.io` resource groups (`gatewayclasses`, `gateways`, `httproutes`, `grpcroutes`, `tlsroutes`, `tcproutes`, `udproutes`) with status subresources.
* **Readiness Probing**: Controller exposes `/readyz` and `/healthz` endpoints at port 8081 for Kubernetes liveness and readiness monitoring.
* **ClusterMesh Integration**: Seamlessly routes across federated clusters using `ClusterPeering` CRD and WireGuard transport.

---

## 4. Hardware Acceleration & SmartNIC Offload

For high-bandwidth edge clusters, StraitGateway offloads Gateway API route lookups to SmartNICs running `XDP_OFFLOAD`, achieving 100GbE line-rate switching with zero host CPU intervention.
