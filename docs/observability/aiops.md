# Autonomous Dataplane AIOps

StraitGateway's AIOps subsystem provides fully autonomous network operations: it continuously observes the dataplane, detects anomalies in real time, and executes coordinated remediation actions — all without operator intervention.

---

## 1. Overview

### 1.1 What AIOps Does

| Capability | Implementation | Trigger |
|---|---|---|
| Micro-burst detection | Dual-EWMA rate monitor (10ms sampling) | EWMA > baseline × 3× for > 20ms |
| BGP route withdrawal | Automated via `BGPWithdrawer` interface | HIGH severity burst (>8× baseline) |
| Route re-advertisement | Background calm-period monitor | EWMA < baseline × 1.5 for 5s after 60s cooldown |
| Alertmanager alerts | HTTP POST JSON payload | MED severity burst (4–8×) |
| W3C Trace Context | Per-flow header inject/extract | All HTTP flows through NodeProxy |
| Flow correlation | `sg_roce_events` ring buffer drain | Every ECN-CE event |

### 1.2 Architecture

```
┌──────────────────────────────────────────────────────────┐
│                    straitd / sg-agent                    │
│                                                          │
│   ┌─────────────────┐    ┌──────────────────────────┐   │
│   │  BurstDetector  │    │    W3C Trace Propagator   │   │
│   │  (10ms ticks)   │    │  (Extract → Inject flow)  │   │
│   └────────┬────────┘    └──────────────────────────┘   │
│            │ BurstEvent                                   │
│            ▼                                             │
│   ┌─────────────────┐    ┌──────────────────────────┐   │
│   │  AIOps Engine   │───►│   BGP Speaker            │   │
│   │  (severity +    │    │   WithdrawRoute()        │   │
│   │   MRAI guard)   │    │   AdvertisePrefix()      │   │
│   └────────┬────────┘    └──────────────────────────┘   │
│            │                                             │
│            ▼                                             │
│   ┌─────────────────┐                                   │
│   │  Alertmanager   │                                   │
│   │  Webhook POST   │                                   │
│   └─────────────────┘                                   │
└──────────────────────────────────────────────────────────┘
```

---

## 2. Micro-Burst Detection

### 2.1 Algorithm

The `BurstDetector` uses a **dual Exponentially Weighted Moving Average (EWMA)** approach:

```
Fast EWMA (detection window, N=20 samples):
  α_fast = 2 / (N + 1) = 0.095
  EWMA_fast_t = α_fast × rate_t + (1 - α_fast) × EWMA_fast_{t-1}

Slow EWMA (baseline reference, N=300 samples):
  α_slow = 2 / (N + 1) = 0.0066
  EWMA_slow_t = α_slow × rate_t + (1 - α_slow) × EWMA_slow_{t-1}

Anomaly condition:
  EWMA_fast > EWMA_slow × AnomalyMultiplier
  AND duration > MinBurstDuration
```

The **fast EWMA** responds quickly to rate spikes; the **slow EWMA** tracks the long-term baseline, self-adapting to legitimate traffic growth without false positives.

### 2.2 Counter Sampling

Every 10ms the `BurstDetector` reads:
```
/sys/class/net/<iface>/statistics/rx_bytes
/sys/class/net/<iface>/statistics/tx_bytes
```

Delta bytes ÷ elapsed seconds = instantaneous byte rate. Counter wrap-around (64-bit rollover) is handled gracefully via `counterDelta()`.

### 2.3 Configuration

```go
cfg := aiops.BurstDetectorConfig{
    Interface:          "eth0",
    SampleInterval:     10 * time.Millisecond,
    WindowSize:         20,                 // fast EWMA samples
    BaselineWindowSize: 300,                // slow EWMA samples
    AnomalyMultiplier:  3.0,               // 3× baseline = burst
    MinBurstDuration:   20 * time.Millisecond,
    BurstThresholdBps:  10_000_000_000,    // 10 Gbps absolute floor
    OnBurst: engine.HandleBurst,
}
```

### 2.4 BurstEvent Fields

| Field | Type | Description |
|---|---|---|
| `Timestamp` | `time.Time` | When the burst was first detected |
| `Interface` | `string` | Network interface name |
| `PeakRateBps` | `float64` | Observed peak rate in bits/second |
| `BaselineRateBps` | `float64` | Slow EWMA at detection time |
| `AnomalyMultiplier` | `float64` | PeakRate ÷ BaselineRate |
| `DurationMs` | `float64` | How long the burst persisted (ms) |
| `BytesObserved` | `uint64` | Total bytes during burst window |

---

## 3. Severity Classification and Remediation

### 3.1 Severity Tiers

| Severity | Multiplier Range | Actions |
|---|---|---|
| `LOW` | 2–4× | Structured log only |
| `MED` | 4–8× | Log + Alertmanager webhook |
| `HIGH` | >8× | Log + Alert + BGP route withdrawal |

### 3.2 BGP Route Withdrawal

When a `HIGH` severity burst is detected, the Engine calls `BGPWithdrawer.WithdrawRoute()` for all configured BGP peers. This shifts traffic to alternate paths within the BGP convergence time.

**MRAI (Minimum Route Advertisement Interval)** guard prevents withdrawal flapping: a minimum interval of 30 seconds (configurable) is enforced between successive withdrawal actions on the same interface.

### 3.3 Route Re-advertisement

After withdrawal, a background goroutine runs every second to check:

```
1. Has CooldownDuration (60s) elapsed since withdrawal?
2. Is EWMA_fast < EWMA_slow × ResumeMultiplier (1.5) for ResumeQuietDuration (5s)?
```

If both conditions are met, routes are re-advertised via `BGPWithdrawer.AdvertisePrefix()`.

### 3.4 Engine Configuration

```go
cfg := aiops.EngineConfig{
    NodeName:            "worker-01",
    MinActionInterval:   30 * time.Second,
    CooldownDuration:    60 * time.Second,
    ResumeQuietDuration: 5 * time.Second,
    ResumeMultiplier:    1.5,
    WithdrawSeverity:    aiops.SeverityHigh,
    AlertWebhookURL:     "http://alertmanager:9093/api/v1/alerts",
    BGP:                 bgpSpeaker,         // implements BGPWithdrawer
    Detectors:           map[string]*aiops.BurstDetector{
        "eth0": det,
    },
}
```

---

## 4. W3C Trace Context Propagation

### 4.1 Standard Reference

StraitGateway implements the **W3C Trace Context Level 1** recommendation (2021-11-23). All HTTP flows traversing the NodeProxy L7 gateway automatically receive traceparent/tracestate headers — no application code changes required.

### 4.2 `traceparent` Header

```
traceparent: 00-4bf92f3577b34da6a3ce929d0e0e4736-00f067aa0ba902b7-01
              ││  ────────────────────────────────  ────────────────  ││
              ││  trace-id (128-bit, 32 hex chars)  span-id (64-bit)  ││
              ││                                                       ││
              └┘ version "00"                              flags "01" ┘
                                                    (bit 0 = sampled)
```

### 4.3 Propagation Flow

```
Incoming HTTP Request
  │
  ├─ Extract(flowKey, headers)
  │    ├─ Parse traceparent from headers
  │    │    ├─ Valid → use upstream TraceID, upstream SpanID as parent
  │    │    └─ Missing/invalid → generate new 128-bit TraceID + SpanID
  │    └─ Parse tracestate
  │
  ├─ Process request (NodeProxy L7)
  │
  └─ Inject(parentContext, outboundHeaders)
       ├─ Mint new 64-bit SpanID (crypto/rand)
       ├─ Preserve TraceID from parent
       ├─ Prepend "sg=straitgateway" to tracestate
       └─ Set traceparent + tracestate in outbound headers
```

### 4.4 Flow Cache

The `Propagator` maintains a bounded in-memory flow cache mapping 5-tuple flow keys to `FlowTraceRecord`:

```
Key:   "srcIP:srcPort→dstIP:dstPort/proto"
Value: { SpanContext, IngressAt, EgressAt }
```

- **Capacity**: 65,536 entries (configurable via `MaxFlowCacheSize`)
- **TTL**: 5 minutes (configurable via `FlowCacheTTL`)
- **Eviction**: `EvictExpired()` runs every 30 seconds from the agent's cleanup goroutine

### 4.5 `tracestate` Format

StraitGateway prepends its vendor entry to the existing tracestate:

```
tracestate: sg=straitgateway,existing-vendor=value
```

Per W3C spec: the leftmost entry is the most recent hop; vendor entries must be valid list-members (max 32 entries, max 256 chars each).

### 4.6 OpenTelemetry Integration

The `Propagator` is compatible with the OpenTelemetry `TextMapPropagator` interface. When used with `otelhttp`:

```go
import "go.opentelemetry.io/otel/propagation"

// Adapt StraitGateway Propagator to OTel interface
otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(
    propagation.TraceContext{},  // W3C standard
    propagation.Baggage{},
))
```

Traces automatically appear in Jaeger, Grafana Tempo, and Google Cloud Trace.

---

## 5. Prometheus Metrics

AIOps metrics exposed at `/metrics`:

| Metric | Type | Labels | Description |
|---|---|---|---|
| `sg_aiops_burst_events_total` | Counter | `interface`, `severity`, `node` | Burst events fired |
| `sg_aiops_bgp_withdrawals_total` | Counter | `interface`, `node` | BGP withdrawal actions |
| `sg_aiops_bgp_readvertisements_total` | Counter | `interface`, `node` | Route re-advertisement actions |
| `sg_aiops_mrai_suppressions_total` | Counter | `interface`, `node` | Withdrawals suppressed by MRAI |
| `sg_aiops_ewma_rate_bps` | Gauge | `interface`, `node` | Current fast EWMA rate |
| `sg_aiops_baseline_rate_bps` | Gauge | `interface`, `node` | Current slow EWMA baseline rate |
| `sg_trace_flows_active` | Gauge | `node` | Active flow→trace cache entries |
| `sg_trace_created_total` | Counter | `node` | New traces generated (no incoming context) |
| `sg_trace_propagated_total` | Counter | `node` | Traces forwarded from upstream |

---

## 6. Alertmanager Integration

Burst events at `MED` or `HIGH` severity trigger an HTTP POST to the configured Alertmanager webhook URL with an Alertmanager-compatible JSON payload:

```json
[{
  "labels": {
    "alertname": "MicroBurst",
    "severity":  "HIGH",
    "interface": "eth0",
    "node":      "worker-01"
  },
  "annotations": {
    "summary": "Micro-burst on eth0: peak 12340.0 Mbps (10.3x baseline)"
  }
}]
```

### Alertmanager Route Configuration

```yaml
# alertmanager.yml
route:
  group_by: ['alertname', 'node']
  receiver: 'pagerduty'
  routes:
    - match:
        alertname: MicroBurst
        severity: HIGH
      receiver: 'oncall-critical'
      continue: true
    - match:
        alertname: MicroBurst
        severity: MED
      receiver: 'slack-warnings'
```

---

## 7. Operational Reference

### Monitor Live Burst Detection

```bash
# Watch AIOps event log in real time
sgctl aiops watch --interface eth0

# Show current EWMA rate and baseline
sgctl aiops rate --interface eth0
```

### BGP Withdrawal Status

```bash
# List interfaces with active BGP withdrawals
sgctl aiops withdrawals

# Manually trigger route re-advertisement (override cooldown)
sgctl aiops readvertise --interface eth0 --force
```

### Trace Context Diagnostics

```bash
# Show flow cache statistics
sgctl trace stats

# Look up trace context for a specific flow
sgctl trace get --src 10.0.0.1:54321 --dst 10.0.0.2:80 --proto TCP

# Force eviction of expired flow cache entries
sgctl trace evict
```

---

## 8. Tuning Guide

### Reducing False Positives

If the burst detector fires frequently during normal traffic:

1. **Increase `AnomalyMultiplier`** from 3.0 to 4.0–5.0
2. **Increase `MinBurstDuration`** from 20ms to 50ms
3. **Increase `BaselineWindowSize`** from 300 to 600 samples (slower adaptation)
4. **Increase `WithdrawSeverity`** from `HIGH` to never trigger BGP withdrawal on MED events

### Improving Detection Sensitivity

For environments with strict latency SLAs where even short micro-bursts cause tail latency:

1. **Decrease `SampleInterval`** from 10ms to 5ms
2. **Decrease `AnomalyMultiplier`** to 2.0
3. **Decrease `MinBurstDuration`** to 10ms
4. **Set `BurstThresholdBps`** to interface line rate × 0.8 (absolute ceiling)

### MRAI Tuning

The default MRAI of 30s balances between fast remediation and BGP stability. In environments with fast BGP convergence (< 1s via BFD):

```go
cfg.MinActionInterval = 5 * time.Second  // allow faster successive actions
cfg.CooldownDuration  = 15 * time.Second // shorter recovery check period
```
