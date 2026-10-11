# Zero-Trust Security Policies

StraitGateway enforces fine-grained, identity-based microsegmentation across all L3, L4, and L7 traffic in Kubernetes.

---

## 1. 32-Bit Workload Identities

Rather than filtering by ephemeral IP addresses, StraitGateway assigns a permanent 32-bit cryptographic identity token to each workload.

* **Identity Derivation Factors**:
  * Kubernetes Namespace
  * ServiceAccount Name
  * Pod Labels (e.g. `app=checkout`, `env=prod`)
* **Identity Cache Map (`sg_identities`)**:
  * Synchronizes each running Pod's allocated IP with its derived 32-bit ID.

---

## 2. Declarative `SecurityPolicy` CRD

```yaml
apiVersion: straitgateway.io/v1alpha1
kind: SecurityPolicy
metadata:
  name: protect-payment-service
  namespace: finance
spec:
  priority: 10
  defaultAction: Deny
  selector:
    matchLabels:
      app: payment
  ingress:
    - from:
        - namespace: store
          podSelector:
            matchLabels:
              app: checkout
          serviceAccount: checkout-sa
      ports:
        - protocol: TCP
          port: 8443
      http:
        - methods: ["POST"]
          paths: ["/api/v1/charge"]
  egress:
    - to:
        - namespace: database
          podSelector:
            matchLabels:
              role: postgres
      ports:
        - protocol: TCP
          port: 5432
```

---

## 3. Fast In-Kernel Evaluation

Policies compile into BPF entries in `sg_policies`:

```
[Key: src_id=1002, dst_id=1005, port=8443, proto=TCP] ---> [Value: Action=ALLOW, RuleID=42]
```

* **Microsecond Decision Speed**: BPF map lookups execute in under $1\mu s$ on TCX ingress.
* **Default-Deny Model**: Any packet between workloads lacking an explicit matching policy is dropped at line rate (`TC_ACT_SHOT`) and logged to `sg_events`.
