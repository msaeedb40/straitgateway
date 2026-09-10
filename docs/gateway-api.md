# StraitGateway: Gateway API Implementation

StraitGateway includes a native, high-performance controller for the Kubernetes **Gateway API v1.6.1** specification. Instead of provisioning heavyweight reverse-proxy sidecars or standalone NGINX/Envoy pods, StraitGateway reconciles Gateway API resources into its unified **Intermediate Representation (IR)** and accelerates routing directly in the Linux kernel via eBPF.

---

## Gateway API Identifiers & Class

To route traffic through StraitGateway, Gateway resources must reference the managed GatewayClass:

| Property | Value |
| :--- | :--- |
| **GatewayClass Controller Name** | `straitgateway.io/skgateway` |
| **Managed GatewayClass Name** | `skgateway` |
| **API Version** | `gateway.networking.k8s.io/v1` |

---

## 1. GatewayClass Manifest

When StraitGateway is installed, it reconciles or provisions the default `GatewayClass`:

```yaml
apiVersion: gateway.networking.k8s.io/v1
kind: GatewayClass
metadata:
  name: skgateway
spec:
  controllerName: "straitgateway.io/skgateway"
  description: "StraitGateway eBPF-accelerated Gateway API controller"
```

---

## 2. Gateway Resource

The `Gateway` resource defines the network entry points (listeners) and ports for incoming traffic.

```yaml
apiVersion: gateway.networking.k8s.io/v1
kind: Gateway
metadata:
  name: edge-gateway
  namespace: straitgateway-system
spec:
  gatewayClassName: skgateway
  listeners:
    # HTTP Listener
    - name: http
      protocol: HTTP
      port: 80
      allowedRoutes:
        namespaces:
          from: All

    # HTTPS Listener (TLS Termination)
    - name: https
      protocol: HTTPS
      port: 443
      tls:
        mode: Terminate
        certificateRefs:
          - kind: Secret
            name: edge-tls-secret
      allowedRoutes:
        namespaces:
          from: All

    # Raw TCP Listener
    - name: tcp-db
      protocol: TCP
      port: 5432
      allowedRoutes:
        namespaces:
          from: Same

    # Raw UDP Listener
    - name: udp-dns
      protocol: UDP
      port: 5353
      allowedRoutes:
        namespaces:
          from: Same

    # TLS Passthrough Listener (SNI)
    - name: tls-sni
      protocol: TLS
      port: 8443
      tls:
        mode: Passthrough
      allowedRoutes:
        namespaces:
          from: All
```

---

## 3. Route Types & Specifications

### A. `HTTPRoute`: Path, Header, and Weighted Canary Routing
Directs HTTP/HTTPS traffic to backends with path-prefix matching, custom header rules, and weighted traffic splits:

```yaml
apiVersion: gateway.networking.k8s.io/v1
kind: HTTPRoute
metadata:
  name: web-app-route
  namespace: default
spec:
  parentRefs:
    - name: edge-gateway
      namespace: straitgateway-system
      sectionName: http
  hostnames:
    - "app.example.com"
  rules:
    # Rule 1: API traffic routed to backend API
    - matches:
        - path:
            type: PathPrefix
            value: /api/v1
          headers:
            - type: Exact
              name: X-Client-Version
              value: "v2"
      backendRefs:
        - name: api-service-v2
          port: 8080

    # Rule 2: Canary traffic splitting (90% stable, 10% canary)
    - matches:
        - path:
            type: PathPrefix
            value: /
      backendRefs:
        - name: web-frontend-stable
          port: 80
          weight: 90
        - name: web-frontend-canary
          port: 80
          weight: 10
```

---

### B. `TCPRoute`: High-Speed L4 Stream Routing
Routes raw TCP streams directly to backend services (e.g., databases, Redis, message brokers):

```yaml
apiVersion: gateway.networking.k8s.io/v1alpha2
kind: TCPRoute
metadata:
  name: postgres-route
  namespace: default
spec:
  parentRefs:
    - name: edge-gateway
      namespace: straitgateway-system
      sectionName: tcp-db
  rules:
    - backendRefs:
        - name: postgres-service
          port: 5432
```

---

### C. `UDPRoute`: Datagram Proxying
Accelerates stateless or real-time UDP protocols (DNS, game servers, VoIP, streaming) via eBPF:

```yaml
apiVersion: gateway.networking.k8s.io/v1alpha2
kind: UDPRoute
metadata:
  name: custom-dns-route
  namespace: default
spec:
  parentRefs:
    - name: edge-gateway
      namespace: straitgateway-system
      sectionName: udp-dns
  rules:
    - backendRefs:
        - name: coredns-custom
          port: 5353
```

---

### D. `TLSRoute`: SNI-Based Passthrough
Directs encrypted TLS traffic based on the Server Name Indication (SNI) header without decrypting payload data:

```yaml
apiVersion: gateway.networking.k8s.io/v1alpha2
kind: TLSRoute
metadata:
  name: secure-passthrough-route
  namespace: default
spec:
  parentRefs:
    - name: edge-gateway
      namespace: straitgateway-system
      sectionName: tls-sni
  hostnames:
    - "secure.vault.example.com"
  rules:
    - backendRefs:
        - name: vault-service
          port: 8443
```

---

### E. `GRPCRoute`: gRPC Service & Method Routing
Routes gRPC calls based on RPC service and method names:

```yaml
apiVersion: gateway.networking.k8s.io/v1
kind: GRPCRoute
metadata:
  name: payment-grpc-route
  namespace: default
spec:
  parentRefs:
    - name: edge-gateway
      namespace: straitgateway-system
      sectionName: https
  hostnames:
    - "grpc.payment.example.com"
  rules:
    - matches:
        - method:
            service: "payment.v1.PaymentService"
            method: "ProcessTransaction"
      backendRefs:
        - name: transaction-processor
          port: 9000
    - matches:
        - method:
            service: "payment.v1.PaymentService"
      backendRefs:
        - name: general-payment-service
          port: 9000
```

---

### F. `ReferenceGrant`: Cross-Namespace Routing
Enables a Gateway in `straitgateway-system` to securely route traffic to backend Services located in other namespaces:

```yaml
apiVersion: gateway.networking.k8s.io/v1beta1
kind: ReferenceGrant
metadata:
  name: allow-edge-gateway-backends
  namespace: default
spec:
  from:
    - group: gateway.networking.k8s.io
      kind: Gateway
      namespace: straitgateway-system
  to:
    - group: ""
      kind: Service
```
