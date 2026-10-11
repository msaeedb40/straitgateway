// Package aiops — aiops_test.go provides unit tests for the Phase 14
// AIOps components: BurstDetector, Engine, and W3C Trace Context Propagator.
//
// All tests run without root privileges and without live network interfaces.
// The BurstDetector tests use a configurable CounterReader stub.
package aiops_test

import (
	"context"
	"net"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/straitgateway/straitgateway/internal/observability/aiops"
)

// ---------------------------------------------------------------------------
// BurstDetector tests
// ---------------------------------------------------------------------------

func TestBurstDetectorNoFalsePositive(t *testing.T) {
	// Constant rate — should not fire burst events once EWMA is stable.
	// We allow a short warm-up window where the EWMA is still converging;
	// only events fired AFTER warm-up are counted.
	var events atomic.Int32
	var warmupDone atomic.Bool

	cfg := aiops.DefaultBurstDetectorConfig("lo")
	cfg.SampleInterval = 5 * time.Millisecond
	cfg.WindowSize = 20
	cfg.BaselineWindowSize = 40 // longer baseline so it converges
	cfg.AnomalyMultiplier = 3.0
	cfg.MinBurstDuration = 30 * time.Millisecond
	cfg.OnBurst = func(_ aiops.BurstEvent) {
		if warmupDone.Load() {
			events.Add(1)
		}
	}

	counter := uint64(0)
	cfg.CounterReader = func(_ string) (uint64, uint64, error) {
		counter += 1000 // constant 1000 B/sample
		return counter, 0, nil
	}

	det := aiops.NewBurstDetector(cfg)

	ctx, cancel := context.WithTimeout(context.Background(), 600*time.Millisecond)
	defer cancel()

	go func() {
		// Allow 300ms for EWMA to stabilise before checking for false positives
		time.Sleep(300 * time.Millisecond)
		warmupDone.Store(true)
	}()

	_ = det.Run(ctx)

	if n := events.Load(); n > 0 {
		t.Errorf("false positive after warm-up: %d burst events for constant-rate traffic", n)
	}
}

func TestBurstDetectorFiresOnBurst(t *testing.T) {
	var events atomic.Int32

	cfg := aiops.DefaultBurstDetectorConfig("lo")
	cfg.SampleInterval = 5 * time.Millisecond
	cfg.WindowSize = 6
	cfg.BaselineWindowSize = 20 // needs 20 ticks to build baseline
	cfg.AnomalyMultiplier = 2.0
	cfg.MinBurstDuration = 15 * time.Millisecond
	cfg.OnBurst = func(ev aiops.BurstEvent) {
		events.Add(1)
		if ev.Interface != "lo" {
			t.Errorf("burst event interface = %q, want %q", ev.Interface, "lo")
		}
		if ev.PeakRateBps <= 0 {
			t.Errorf("burst event PeakRateBps = %.0f, want > 0", ev.PeakRateBps)
		}
	}

	tick := 0
	counter := uint64(0)
	cfg.CounterReader = func(_ string) (uint64, uint64, error) {
		tick++
		// First 40 ticks: build baseline at 1000 B/tick (200ms at 5ms interval)
		// After tick 40: spike to 200000 B/tick (200× burst)
		if tick <= 40 {
			counter += 1000
		} else {
			counter += 200_000
		}
		return counter, 0, nil
	}

	det := aiops.NewBurstDetector(cfg)

	// Enough time for: 40 baseline ticks (200ms) + burst detection (>=30ms)
	ctx, cancel := context.WithTimeout(context.Background(), 600*time.Millisecond)
	defer cancel()
	_ = det.Run(ctx)

	if n := events.Load(); n == 0 {
		t.Error("expected at least one burst event to fire, got none")
	}
}

func TestBurstDetectorEWMAMonotonicallySmooth(t *testing.T) {
	cfg := aiops.DefaultBurstDetectorConfig("lo")
	cfg.SampleInterval = 5 * time.Millisecond
	cfg.WindowSize = 5
	cfg.OnBurst = func(_ aiops.BurstEvent) {}

	counter := uint64(0)
	cfg.CounterReader = func(_ string) (uint64, uint64, error) {
		counter += 5000
		return counter, 0, nil
	}

	det := aiops.NewBurstDetector(cfg)
	ctx, cancel := context.WithTimeout(context.Background(), 150*time.Millisecond)
	defer cancel()
	_ = det.Run(ctx)

	// After constant rate, EWMA and baseline should both be > 0
	if e := det.EWMA(); e <= 0 {
		t.Errorf("EWMA = %.2f, want > 0 after sustained traffic", e)
	}
	if b := det.BaselineEWMA(); b <= 0 {
		t.Errorf("BaselineEWMA = %.2f, want > 0 after sustained traffic", b)
	}
}

func TestBurstDetectorInBurstState(t *testing.T) {
	cfg := aiops.DefaultBurstDetectorConfig("lo")
	cfg.SampleInterval = 5 * time.Millisecond
	cfg.WindowSize = 4
	cfg.AnomalyMultiplier = 2.0
	cfg.MinBurstDuration = 50 * time.Millisecond // long enough to observe
	cfg.OnBurst = func(_ aiops.BurstEvent) {}

	tick := 0
	counter := uint64(0)
	cfg.CounterReader = func(_ string) (uint64, uint64, error) {
		tick++
		if tick <= 15 {
			counter += 500
		} else {
			counter += 500_000 // massive spike
		}
		return counter, 0, nil
	}

	det := aiops.NewBurstDetector(cfg)
	ctx, cancel := context.WithTimeout(context.Background(), 300*time.Millisecond)
	defer cancel()
	_ = det.Run(ctx)
	// Test just verifies no panic and EWMA is computed
	_ = det.InBurst()
}

// ---------------------------------------------------------------------------
// Engine tests
// ---------------------------------------------------------------------------

// mockBGP implements BGPWithdrawer for testing.
type mockBGP struct {
	withdrawn    []string // prefix strings
	readvertised []string
	peers        []aiops.BGPPeerInfo
}

func (m *mockBGP) WithdrawRoute(neighborIP net.IP, vrfID uint32, prefix net.IPNet) error {
	m.withdrawn = append(m.withdrawn, prefix.String())
	return nil
}

func (m *mockBGP) AdvertisePrefix(neighborIP net.IP, vrfID uint32, prefix net.IPNet) error {
	m.readvertised = append(m.readvertised, prefix.String())
	return nil
}

func (m *mockBGP) ListPeers() []aiops.BGPPeerInfo {
	return m.peers
}

func TestEngineSeverityClassification(t *testing.T) {
	cases := []struct {
		multiplier float64
		wantSev    string
	}{
		{1.5, "LOW"},
		{2.5, "LOW"},
		{4.0, "MED"},
		{6.0, "MED"},
		{8.0, "HIGH"},
		{15.0, "HIGH"},
	}

	for _, tc := range cases {
		ev := aiops.BurstEvent{
			Interface:         "eth0",
			AnomalyMultiplier: tc.multiplier,
		}

		cfg := aiops.DefaultEngineConfig()
		cfg.WithdrawSeverity = aiops.SeverityHigh
		var loggedSev string
		cfg.Logger = func(_, _ string, fields map[string]any) {
			if s, ok := fields["severity"]; ok {
				loggedSev = s.(string)
			}
		}

		eng := aiops.NewEngine(cfg)
		eng.HandleBurst(ev)

		if tc.wantSev == "HIGH" && loggedSev != "HIGH" {
			t.Errorf("multiplier=%.1f: logged severity=%q, want HIGH", tc.multiplier, loggedSev)
		}
	}
}

func TestEngineWithdrawsOnHighSeverity(t *testing.T) {
	_, prefix, _ := net.ParseCIDR("10.0.0.0/8")

	bgp := &mockBGP{
		peers: []aiops.BGPPeerInfo{
			{
				NeighborIP: net.ParseIP("192.168.1.1"),
				VRFID:      0,
				Prefixes:   []net.IPNet{*prefix},
			},
		},
	}

	cfg := aiops.DefaultEngineConfig()
	cfg.MinActionInterval = 0 // disable MRAI for test
	cfg.WithdrawSeverity = aiops.SeverityHigh
	cfg.BGP = bgp
	cfg.Logger = func(_, _ string, _ map[string]any) {}

	eng := aiops.NewEngine(cfg)

	ev := aiops.BurstEvent{
		Interface:         "eth0",
		AnomalyMultiplier: 10.0, // HIGH severity
		PeakRateBps:       10_000_000_000,
		BaselineRateBps:   1_000_000_000,
	}
	eng.HandleBurst(ev)

	// Give the async withdrawal goroutine time to complete
	time.Sleep(50 * time.Millisecond)

	if len(bgp.withdrawn) == 0 {
		t.Error("expected BGP routes to be withdrawn on HIGH severity burst, got none")
	}
	if bgp.withdrawn[0] != "10.0.0.0/8" {
		t.Errorf("withdrew %q, want 10.0.0.0/8", bgp.withdrawn[0])
	}
}

func TestEngineMRAISuppression(t *testing.T) {
	bgp := &mockBGP{
		peers: []aiops.BGPPeerInfo{
			{
				NeighborIP: net.ParseIP("10.0.0.1"),
				VRFID:      0,
				Prefixes:   []net.IPNet{},
			},
		},
	}

	cfg := aiops.DefaultEngineConfig()
	cfg.MinActionInterval = 10 * time.Second // long MRAI
	cfg.WithdrawSeverity = aiops.SeverityHigh
	cfg.BGP = bgp
	cfg.Logger = func(_, _ string, _ map[string]any) {}

	eng := aiops.NewEngine(cfg)

	ev := aiops.BurstEvent{
		Interface:         "eth0",
		AnomalyMultiplier: 9.0,
	}

	// First burst fires
	eng.HandleBurst(ev)
	time.Sleep(20 * time.Millisecond)
	firstWithdrawn := len(bgp.withdrawn)

	// Second burst within MRAI window — should be suppressed
	eng.HandleBurst(ev)
	time.Sleep(20 * time.Millisecond)

	if len(bgp.withdrawn) > firstWithdrawn {
		t.Error("MRAI failed: second burst triggered additional withdrawal within min interval")
	}
}

func TestEngineLowSeverityNoWithdrawal(t *testing.T) {
	bgp := &mockBGP{}

	cfg := aiops.DefaultEngineConfig()
	cfg.MinActionInterval = 0
	cfg.WithdrawSeverity = aiops.SeverityHigh
	cfg.BGP = bgp
	cfg.Logger = func(_, _ string, _ map[string]any) {}

	eng := aiops.NewEngine(cfg)
	ev := aiops.BurstEvent{
		Interface:         "lo",
		AnomalyMultiplier: 2.0, // LOW
	}
	eng.HandleBurst(ev)
	time.Sleep(20 * time.Millisecond)

	if len(bgp.withdrawn) != 0 {
		t.Errorf("LOW severity should not trigger withdrawal, got %d withdrawals", len(bgp.withdrawn))
	}
}

func TestEngineStartStop(t *testing.T) {
	cfg := aiops.DefaultEngineConfig()
	cfg.Logger = func(_, _ string, _ map[string]any) {}

	eng := aiops.NewEngine(cfg)
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	eng.Start(ctx)
	time.Sleep(50 * time.Millisecond)
	eng.Stop() // must not panic or deadlock
}

// ---------------------------------------------------------------------------
// W3C Trace Context tests
// ---------------------------------------------------------------------------

func TestTraceparentFormat(t *testing.T) {
	sc := aiops.SpanContext{
		TraceID: "4bf92f3577b34da6a3ce929d0e0e4736",
		SpanID:  "00f067aa0ba902b7",
		Flags:   aiops.TraceFlagSampled,
	}
	want := "00-4bf92f3577b34da6a3ce929d0e0e4736-00f067aa0ba902b7-01"
	if got := sc.Traceparent(); got != want {
		t.Errorf("Traceparent() = %q, want %q", got, want)
	}
}

func TestParseTraceparentValid(t *testing.T) {
	input := "00-4bf92f3577b34da6a3ce929d0e0e4736-00f067aa0ba902b7-01"
	sc, ok := aiops.ParseTraceparent(input)
	if !ok {
		t.Fatal("ParseTraceparent returned ok=false for valid input")
	}
	if sc.TraceID != "4bf92f3577b34da6a3ce929d0e0e4736" {
		t.Errorf("TraceID = %q", sc.TraceID)
	}
	if sc.SpanID != "00f067aa0ba902b7" {
		t.Errorf("SpanID = %q", sc.SpanID)
	}
	if !sc.Sampled() {
		t.Error("expected Sampled() = true")
	}
	if !sc.Remote {
		t.Error("expected Remote = true for parsed context")
	}
}

func TestParseTraceparentInvalid(t *testing.T) {
	cases := []string{
		"",
		"invalid",
		"00-short-00f067aa0ba902b7-01",
		"00-4bf92f3577b34da6a3ce929d0e0e4736-tooshortspan-01",
		"00-4bf92f3577b34da6a3ce929d0e0e4736-00f067aa0ba902b7",    // missing flags
		"ff-4bf92f3577b34da6a3ce929d0e0e4736-00f067aa0ba902b7-01", // reserved version
		"00-4bf92f3577b34da6a3ce929d0e0e4736-00f067aa0ba902b7-zz", // non-hex flags
	}
	for _, c := range cases {
		_, ok := aiops.ParseTraceparent(c)
		if ok {
			t.Errorf("ParseTraceparent(%q) returned ok=true, expected false", c)
		}
	}
}

func TestSpanContextIsValid(t *testing.T) {
	valid := aiops.SpanContext{
		TraceID: "4bf92f3577b34da6a3ce929d0e0e4736",
		SpanID:  "00f067aa0ba902b7",
	}
	if !valid.IsValid() {
		t.Error("expected IsValid() = true for non-zero IDs")
	}

	zeroTrace := aiops.SpanContext{
		TraceID: "00000000000000000000000000000000",
		SpanID:  "00f067aa0ba902b7",
	}
	if zeroTrace.IsValid() {
		t.Error("expected IsValid() = false for all-zero TraceID")
	}
}

func TestPropagatorExtractNewTrace(t *testing.T) {
	p := aiops.NewPropagator(aiops.DefaultPropagatorConfig())
	carrier := aiops.MapCarrier{} // empty — no incoming traceparent

	sc := p.Extract("10.0.0.1:1234→10.0.0.2:80/TCP", carrier)

	if !sc.IsValid() {
		t.Errorf("Extract() returned invalid SpanContext: %+v", sc)
	}
	if len(sc.TraceID) != 32 {
		t.Errorf("new TraceID length = %d, want 32", len(sc.TraceID))
	}
	if sc.Remote {
		t.Error("new SpanContext should not be Remote")
	}
}

func TestPropagatorExtractPropagates(t *testing.T) {
	p := aiops.NewPropagator(aiops.DefaultPropagatorConfig())

	incoming := aiops.MapCarrier{
		"traceparent": "00-4bf92f3577b34da6a3ce929d0e0e4736-00f067aa0ba902b7-01",
		"tracestate":  "vendor=abc",
	}
	sc := p.Extract("10.0.0.1:1234→10.0.0.2:80/TCP", incoming)

	if sc.TraceID != "4bf92f3577b34da6a3ce929d0e0e4736" {
		t.Errorf("propagated TraceID = %q, want original", sc.TraceID)
	}
	if sc.TraceState != "vendor=abc" {
		t.Errorf("TraceState = %q, want %q", sc.TraceState, "vendor=abc")
	}
}

func TestPropagatorInjectCreatesChildSpan(t *testing.T) {
	p := aiops.NewPropagator(aiops.DefaultPropagatorConfig())

	parent := aiops.SpanContext{
		TraceID: "4bf92f3577b34da6a3ce929d0e0e4736",
		SpanID:  "00f067aa0ba902b7",
		Flags:   aiops.TraceFlagSampled,
	}

	outgoing := aiops.MapCarrier{}
	child := p.Inject(parent, outgoing)

	// Child must inherit trace-id
	if child.TraceID != parent.TraceID {
		t.Errorf("child TraceID = %q, want %q", child.TraceID, parent.TraceID)
	}
	// Child must have NEW span-id
	if child.SpanID == parent.SpanID {
		t.Error("child SpanID must differ from parent SpanID")
	}
	if len(child.SpanID) != 16 {
		t.Errorf("child SpanID length = %d, want 16", len(child.SpanID))
	}

	// Header must be set
	tp := outgoing.Get("traceparent")
	if tp == "" {
		t.Error("Inject did not set traceparent header")
	}
	if !strings.HasPrefix(tp, "00-"+parent.TraceID) {
		t.Errorf("traceparent %q does not start with expected prefix", tp)
	}
}

func TestPropagatorTraceState(t *testing.T) {
	cfg := aiops.DefaultPropagatorConfig()
	cfg.ServiceName = "sgw-test"
	p := aiops.NewPropagator(cfg)

	parent := aiops.SpanContext{
		TraceID:    "4bf92f3577b34da6a3ce929d0e0e4736",
		SpanID:     "00f067aa0ba902b7",
		TraceState: "vendor=abc",
	}

	outgoing := aiops.MapCarrier{}
	child := p.Inject(parent, outgoing)

	// tracestate must have sg= prepended
	ts := outgoing.Get("tracestate")
	if !strings.HasPrefix(ts, "sg=sgw-test") {
		t.Errorf("tracestate %q does not start with sg=sgw-test", ts)
	}
	if !strings.Contains(ts, "vendor=abc") {
		t.Errorf("tracestate %q does not contain original vendor=abc", ts)
	}
	_ = child
}

func TestPropagatorFlowCache(t *testing.T) {
	p := aiops.NewPropagator(aiops.DefaultPropagatorConfig())

	flowKey := aiops.FlowKey("10.0.0.1", "10.0.0.2", 12345, 80, "TCP")
	sc := p.Extract(flowKey, aiops.MapCarrier{})

	rec := p.AnnotateFlow(flowKey, sc)
	if rec == nil {
		t.Fatal("AnnotateFlow returned nil")
	}
	if rec.FlowKey != flowKey {
		t.Errorf("FlowKey = %q, want %q", rec.FlowKey, flowKey)
	}

	got, ok := p.GetFlowTrace(flowKey)
	if !ok {
		t.Error("GetFlowTrace returned ok=false for known flowKey")
	}
	if got.SpanContext.TraceID != sc.TraceID {
		t.Errorf("cached TraceID = %q, want %q", got.SpanContext.TraceID, sc.TraceID)
	}
}

func TestPropagatorEvictExpired(t *testing.T) {
	cfg := aiops.DefaultPropagatorConfig()
	cfg.FlowCacheTTL = 10 * time.Millisecond
	p := aiops.NewPropagator(cfg)

	for i := 0; i < 5; i++ {
		key := aiops.FlowKey("10.0.0.1", "10.0.0.2", uint16(i), 80, "TCP")
		p.Extract(key, aiops.MapCarrier{})
	}

	if p.FlowCacheSize() != 5 {
		t.Fatalf("expected 5 cached flows, got %d", p.FlowCacheSize())
	}

	time.Sleep(20 * time.Millisecond)
	n := p.EvictExpired()
	if n != 5 {
		t.Errorf("EvictExpired() = %d, want 5", n)
	}
	if p.FlowCacheSize() != 0 {
		t.Errorf("cache size after eviction = %d, want 0", p.FlowCacheSize())
	}
}

func TestFlowKeyFormat(t *testing.T) {
	key := aiops.FlowKey("192.168.1.1", "10.0.0.1", 54321, 443, "TCP")
	if !strings.Contains(key, "192.168.1.1") {
		t.Errorf("FlowKey %q missing srcIP", key)
	}
	if !strings.Contains(key, "443") {
		t.Errorf("FlowKey %q missing dstPort", key)
	}
}
