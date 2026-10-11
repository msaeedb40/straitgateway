# In-Kernel mTLS & Encryption

StraitGateway supports cryptographic mutual TLS (mTLS) without requiring user-space sidecar TLS termination proxies.

---

## 1. Zero-Copy Kernel TLS (kTLS)

Traditional mTLS decrypts traffic in userspace (inside the sidecar container) and re-encrypts or proxies it back to the kernel.

StraitGateway integrates with Linux **kTLS** (`CONFIG_TLS=y`):

```
+-----------------------------------------------------------+
| Pod Application Socket (TCP)                              |
|   | (write syscall)                                       |
|   v                                                       |
| Linux Kernel kTLS Layer                                   |
|   - AES-128-GCM / AES-256-GCM / ChaCha20-Poly1305         |
|   - Direct hardware crypto offload (AES-NI / QAT)         |
|   v                                                       |
| Network Interface / WireGuard                             |
+-----------------------------------------------------------+
```

* **No User-to-Kernel Buffer Copies**: Encrypted payload flows directly through host socket buffers.
* **Handshake Delegation**: Asymmetric TLS handshakes (X.509 certificates and ephemeral key exchange) are conducted via SPIFFE/SPIRE agent, and established session symmetric keys are installed directly into the kernel socket via `setsockopt(fd, SOL_TLS, TLS_TX, ...)`.

---

## 2. SPIFFE/SPIRE Workload Attestation

* Every workload receives a verifiable SPIFFE ID:
  `spiffe://cluster.local/ns/<namespace>/sa/<serviceaccount>`
* Keys and X.509 SVIDs rotate automatically every 60 minutes without restarting containers or dropping long-lived TCP streams.
