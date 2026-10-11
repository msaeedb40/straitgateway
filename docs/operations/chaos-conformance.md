# Chaos Engineering & Conformance Testing

StraitGateway maintains production-grade reliability through two complementary testing disciplines:

1. **Conformance Testing** — validates adherence to CNI and Gateway API specifications across all supported kernel versions
2. **Chaos Engineering** — deliberately injects faults into the eBPF dataplane to verify failure isolation, recovery behaviour, and state-machine correctness

---

## 1. Conformance Test Suite

### 1.1 CNI Conformance (Sonobuoy)

The CNI conformance suite validates StraitGateway's `sg-cni` plugin against the [CNI specification v0.4.0 / v1.0.0](https://github.com/containernetworking/cni/blob/main/SPEC.md) using the [Project Sonobuoy](https://sonobuoy.io/) framework.

**Run the conformance suite:**

```bash
make conformance
# or directly:
sonobuoy run --plugin test/conformance/sonobuoy-plugin.yaml
sonobuoy status
sonobuoy retrieve ./results/
```

**Test inventory:**

| Test | What It Validates |
|---|---|
| `CNI ADD creates interface` | `ADD` call creates a network interface in the container namespace |
| `CNI ADD assigns IP` | Container receives a valid IPv4 address from the IPAM allocator |
| `CNI DEL cleans up` | `DEL` removes the interface and releases the IP |
| `CNI CHECK passes` | `CHECK` returns success immediately after `ADD` |
| `Dual-stack ADD` | Container receives both IPv4 and IPv6 addresses |
| `Concurrent ADD/DEL` | 50 concurrent ADD/DEL operations succeed without IPAM corruption |
| `IP non-overlap` | Two pods on the same node never receive the same IP |
| `IPAM persistence` | IPAM state survives straitd restart without leaking IPs |
| `MTU propagation` | Container interface MTU matches node configuration |
| `Loopback setup` | Container loopback `lo` is configured and UP |

### 1.2 Gateway API Conformance (SIG-Network)

The Gateway API conformance suite runs the official SIG-Network test harness against StraitGateway's Gateway API v1.6.1 implementation:

```bash
make conformance-gateway
# or:
go test -v ./test/conformance/... -run TestGatewayAPIConformance
```

**Conformance profile**: `Gateway` (standard + extended features)

| Feature | Status |
|---|---|
| `HTTPRoute` core routing | ✅ Conformant |
| `HTTPRoute` header matching | ✅ Conformant |
| `HTTPRoute` weight-based splits | ✅ Conformant |
| `HTTPRoute` path rewriting | ✅ Conformant |
| `GRPCRoute` | ✅ Conformant |
| `TCPRoute` | ✅ Conformant |
| `TLSRoute` | ✅ Conformant |
| `BackendTLSPolicy` | ✅ Conformant |
| Cross-namespace references | ✅ Conformant |
| `HTTPRoute` request mirroring | ✅ Conformant |

---

## 2. Chaos Engineering

### 2.1 Chaos Injector (`test/chaos/injector.go`)

The `ChaosInjector` uses `BPF_PROG_TEST_RUN` to execute eBPF programs under synthetic fault conditions without touching production maps:

```go
type ChaosInjector struct {
    cfg     ChaosConfig
    prog    *ebpf.Program
    faultFn FaultFunc
}
```

#### Supported Fault Types

| Fault | `FaultType` | Mechanism | What It Tests |
|---|---|---|---|
| Memory pressure | `FaultENOMEM` | Patches `bpf_map_update_elem` return to `-ENOMEM` | Map-full handling; program doesn't panic |
| Packet drop burst | `FaultDrop` | Returns `XDP_DROP` for every Nth packet | Fast-path recovery; conntrack consistency |
| Latency jitter | `FaultJitter` | Injects artificial delay via spin loop | Timeout expiry handling; ring buffer backpressure |
| BFD session flap | `FaultBFDFlap` | Triggers `handlePeerFailure()` directly | BGP route withdrawal + BFD reconnect |

### 2.2 Running Chaos Tests

```bash
# Requires CAP_BPF + CAP_NET_ADMIN (run as root or with sufficient capabilities)
make chaos

# Run only ENOMEM injection tests
sudo go test -v -run TestENOMEM ./test/chaos/...

# Run BFD flap test
sudo go test -v -run TestBFDFlap ./test/chaos/...

# Run all chaos tests with 5× repetitions
sudo go test -v -count=5 -timeout 300s ./test/chaos/...
```

### 2.3 ENOMEM Injection

Simulates BPF map being full during `bpf_map_update_elem`:

**Expected behaviour:**
- Program returns `XDP_PASS` (fail-open, not crash)
- `sg_counters[MAP_FULL]` increments
- Prometheus metric `sg_map_update_errors_total{error="ENOMEM"}` increments
- straitd logs warning but continues operating

**Test assertion:**
```go
// After ENOMEM injection, packet should be passed (not dropped or panicked)
assert.Equal(t, unix.XDP_PASS, returnCode)
assert.Zero(t, panicCount) // verifier ensures no infinite loops
```

### 2.4 Drop Burst Injection

Injects packet loss at a configurable rate (default: drop every 10th packet):

**Expected behaviour:**
- Conntrack entries for dropped flows age out within `ConntrackTTL` (default: 120s)
- No zombie conntrack entries accumulate
- Flow exporter `sg_drops_total` counter increases monotonically
- TCP connections retransmit and recover (verified via `ss -s` socket statistics)

### 2.5 Latency Jitter Injection

Adds 100µs–10ms artificial latency to the TC ingress path:

**Expected behaviour:**
- BFD sessions tolerate jitter below `DetectionMultiplier × MinRxInterval` (default: 150ms)
- Flow ring buffer does not overflow (ring buffer full → oldest events discarded, not blocked)
- AIOps BurstDetector may fire LOW/MED severity events (verified correct classification)

### 2.6 BFD Flap Injection

Directly triggers `handlePeerFailure()` without BFD packets:

**Expected behaviour:**
1. BGP peer transitions to `StateActive` within 1ms
2. All routes from that peer are withdrawn from `sg_routes` within 10ms
3. BFD session reconnect starts within `HoldTime` (default: 3s)
4. On reconnect, BGP session re-establishes and routes are re-installed

**Timing assertions:**
```go
assert.WithinDuration(t, flap.At, withdrawal.At, 10*time.Millisecond)
assert.WithinDuration(t, reconnect.At, reestablish.At, 3*time.Second)
```

---

## 3. Multi-Kernel Verifier Fuzzing Matrix

### 3.1 Overview

The verifier fuzzing matrix (`test/fuzz/verifier_matrix.go`) systematically mutates eBPF bytecode to find programs that pass the verifier on one kernel version but fail on another:

```
Source programs (*.c)
      │ compile with Clang
      ▼
Bytecode mutations (dead stores, no-op insertions, reg reordering)
      │
      ▼
BPF_PROG_LOAD on each kernel version
      │
      ├─► Pass all → safe program
      ├─► Fail one → version-specific regression
      └─► Fail all → genuine bug in mutation engine
```

### 3.2 Kernel Version Matrix

| Kernel | LTS? | Key BPF Features |
|---|---|---|
| 6.8 | Rolling | TCX, netkit, bpf_loop(), kfuncs v1 |
| 6.12 | LTS | CO-RE improvements, arena allocator, bpf_throw() |
| 7.x | Dev | kfuncs v2, open-coded iterators |

### 3.3 Running the Fuzzing Matrix

```bash
# Run in CI-safe mode (no actual BPF loading — unit tests the mutation engine)
make fuzz-unit

# Run full verifier matrix (requires root + BPF support)
sudo make fuzz

# Fuzz only for the current kernel version
sudo go test -v -run TestVerifierMatrix/$(uname -r | cut -d. -f1-2) \
  ./test/fuzz/...
```

### 3.4 Mutation Strategies

The `MutationEngine` applies one or more of:

| Mutation | Description | Verifier Impact |
|---|---|---|
| `DeadStoreInsert` | Insert `MOV R0, 0` before a write | Tests dead-code elimination |
| `NOPSlide` | Insert sequences of `MOV R1, R1` | Tests instruction limit handling |
| `RegReorder` | Swap two identical register uses | Tests liveness analysis |
| `BoundCheck` | Remove one array bounds check | Should fail verifier (use as negative test) |
| `MapLookupNull` | Remove null check after `bpf_map_lookup_elem` | Should fail verifier |

---

## 4. CI Integration

### 4.1 GitHub Actions Matrix

The CI workflow runs the full test suite on multiple kernel versions using KVM:

```yaml
# .github/workflows/ci.yml (excerpt)
strategy:
  matrix:
    kernel: ["6.8", "6.12"]
    arch:   ["amd64", "arm64"]
steps:
  - name: Boot kernel VM
    run: |
      qemu-system-x86_64 -kernel vmlinuz-${{ matrix.kernel }} \
        -initrd initrd.img -append "root=/dev/sda" \
        -hda rootfs.img -m 4G -smp 4 -nographic
  - name: Run conformance
    run: make conformance
  - name: Run chaos tests
    run: make chaos
  - name: Run fuzz unit tests
    run: make fuzz-unit
```

### 4.2 Makefile Targets

| Target | Description |
|---|---|
| `make conformance` | Run CNI + Gateway API conformance (requires Sonobuoy) |
| `make chaos` | Run chaos fault injection tests (requires CAP_BPF) |
| `make fuzz` | Run full multi-kernel verifier fuzzing matrix |
| `make fuzz-unit` | Run mutation engine unit tests (CI-safe, no BPF) |
| `make sonobuoy` | Deploy Sonobuoy CNI plugin to current cluster |
| `make test-phase12` | All Phase 12 tests (fuzz-unit + chaos) |
| `make test-phase13` | Phase 13 RoCEv2 tests |
| `make test-phase14` | Phase 14 AIOps tests |

---

## 5. Interpreting Results

### 5.1 Conformance Report

Sonobuoy generates a tarball with JUnit XML results:

```bash
sonobuoy retrieve ./results/
tar xzf ./results/*.tar.gz
cat ./results/plugins/cni-conformance/results/global/junit_01.xml
```

A conformant run has zero `<failure>` elements.

### 5.2 Chaos Test Report

```bash
go test -v -json ./test/chaos/... | go run golang.org/x/tools/cmd/tparse
```

Expected output:
```
PASS TestENOMEM_FailOpen       0.23s
PASS TestDropBurst_ConntrackGC 1.45s
PASS TestJitter_BFDTolerance   2.11s
PASS TestBFDFlap_RouteWithdraw 0.89s
```

### 5.3 Fuzz Matrix Report

The `VerifierMatrix` outputs a compatibility table:

```
Mutation: DeadStoreInsert
  Kernel 6.8  → PASS (verified 1.2ms)
  Kernel 6.12 → PASS (verified 0.9ms)

Mutation: BoundCheckRemove
  Kernel 6.8  → FAIL (verifier rejected: invalid mem access 'scalar')  ← expected
  Kernel 6.12 → FAIL (verifier rejected: invalid mem access 'scalar')  ← expected
```

`FAIL` on `BoundCheckRemove` is **expected** — it confirms the verifier is correctly catching unsafe programs. A test failure occurs only if a `BoundCheckRemove` mutation is **accepted** by the verifier.
