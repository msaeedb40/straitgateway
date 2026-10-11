// Package aiops — trace_context.go implements zero-code W3C Trace Context
// (traceparent / tracestate) propagation for all HTTP flows traversing the
// StraitGateway NodeProxy.
//
// # W3C Trace Context (Level 1 — RFC 7230 / W3C Recommendation 2021-11-23)
//
// The traceparent header format:
//
//	{version}-{trace-id}-{parent-id}-{flags}
//	  version  : 2 hex chars (currently "00")
//	  trace-id : 32 hex chars (128-bit)
//	  parent-id: 16 hex chars (64-bit span ID)
//	  flags    : 2 hex chars  (bit 0 = sampled)
//
// # Zero-Code Propagation Strategy
//
// StraitGateway injects Trace Context at the network boundary without
// modifying application code:
//
//  1. Ingress path (Propagator.InjectIncoming):
//     Parses the existing traceparent/tracestate from the incoming HTTP
//     request headers. If absent, generates a new trace-id and parent-id.
//     Stores context in the per-flow map for correlation with flow logs.
//
//  2. Egress path (Propagator.InjectOutgoing):
//     Writes the propagated traceparent/tracestate into outgoing HTTP
//     requests. The parent-id is a freshly minted span-id representing
//     StraitGateway's own processing span. The trace-id is preserved.
//
//  3. Flow correlation (Propagator.AnnotateFlowLog):
//     Attaches trace-id and span-id to FlowLogEntry so traces are
//     cross-referenceable in Jaeger / Tempo / Cloud Trace.
//
// # OpenTelemetry Compatibility
//
// The Propagator is compatible with the OpenTelemetry TextMapPropagator
// interface and can be used directly with otelhttp middleware.  No
// opentelemetry-go import is required — the W3C format is plain text.
//
// Reference: https://www.w3.org/TR/trace-context/
package aiops

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"strings"
	"sync"
	"time"
)

// TraceFlags mirrors W3C trace flags byte (bit 0 = sampled).
type TraceFlags uint8

const (
	TraceFlagSampled TraceFlags = 0x01
)

// SpanContext holds the immutable identity of a single distributed trace span.
type SpanContext struct {
	// TraceID is the 128-bit globally unique trace identifier (32 hex chars).
	TraceID string
	// SpanID is the 64-bit span identifier for the current hop (16 hex chars).
	SpanID string
	// Flags holds sampling and other trace flags.
	Flags TraceFlags
	// TraceState carries vendor-specific key=value pairs (tracestate header).
	TraceState string
	// Remote is true when this SpanContext was received from an upstream hop.
	Remote bool
}

// IsValid returns true when the SpanContext has non-zero IDs.
func (sc SpanContext) IsValid() bool {
	return len(sc.TraceID) == 32 && len(sc.SpanID) == 16 &&
		sc.TraceID != "00000000000000000000000000000000" &&
		sc.SpanID != "0000000000000000"
}

// Sampled returns true if the sampled flag is set.
func (sc SpanContext) Sampled() bool {
	return sc.Flags&TraceFlagSampled != 0
}

// Traceparent returns the W3C traceparent header value for this SpanContext.
func (sc SpanContext) Traceparent() string {
	flags := fmt.Sprintf("%02x", uint8(sc.Flags))
	return "00-" + sc.TraceID + "-" + sc.SpanID + "-" + flags
}

// ParseTraceparent parses a W3C traceparent header value.
// Returns (SpanContext, true) on success.
func ParseTraceparent(value string) (SpanContext, bool) {
	// Format: 00-<trace-id>-<parent-id>-<flags>
	parts := strings.Split(strings.TrimSpace(value), "-")
	if len(parts) != 4 {
		return SpanContext{}, false
	}

	version := parts[0]
	if len(version) != 2 || !isHex(version) {
		return SpanContext{}, false
	}
	// We accept version "00" only (ignore future versions gracefully)
	if version == "ff" {
		return SpanContext{}, false
	}

	traceID := parts[1]
	if len(traceID) != 32 || !isHex(traceID) {
		return SpanContext{}, false
	}

	parentID := parts[2]
	if len(parentID) != 16 || !isHex(parentID) {
		return SpanContext{}, false
	}

	flagsHex := parts[3]
	if len(flagsHex) != 2 || !isHex(flagsHex) {
		return SpanContext{}, false
	}

	flagByte, _ := hex.DecodeString(flagsHex)
	return SpanContext{
		TraceID: traceID,
		SpanID:  parentID,
		Flags:   TraceFlags(flagByte[0]),
		Remote:  true,
	}, true
}

// ParseTracestate parses a W3C tracestate header value.
// Returns the raw string (stored as-is per spec — vendors manage their entries).
func ParseTracestate(value string) string {
	return strings.TrimSpace(value)
}

// FlowTraceRecord correlates a network flow with its W3C trace context.
type FlowTraceRecord struct {
	// FlowKey is a 5-tuple string (srcIP:srcPort→dstIP:dstPort/proto)
	FlowKey string
	// SpanContext is the trace context associated with this flow.
	SpanContext SpanContext
	// IngressAt is when the first packet of the flow was observed.
	IngressAt time.Time
	// EgressAt is the last packet timestamp.
	EgressAt time.Time
}

// PropagatorConfig controls the W3C Trace Context propagator behaviour.
type PropagatorConfig struct {
	// SampleRate is the fraction (0.0–1.0) of new traces to sample. Default: 1.0.
	SampleRate float64
	// ServiceName is included in the tracestate sg= entry. Default: "straitgateway".
	ServiceName string
	// MaxFlowCacheSize is the max number of flow→trace mappings cached. Default: 65536.
	MaxFlowCacheSize int
	// FlowCacheTTL is the duration a flow entry lives in the cache. Default: 5 min.
	FlowCacheTTL time.Duration
}

// DefaultPropagatorConfig returns sensible defaults.
func DefaultPropagatorConfig() PropagatorConfig {
	return PropagatorConfig{
		SampleRate:       1.0,
		ServiceName:      "straitgateway",
		MaxFlowCacheSize: 65536,
		FlowCacheTTL:     5 * time.Minute,
	}
}

// Propagator implements W3C Trace Context injection and extraction.
type Propagator struct {
	mu        sync.RWMutex
	cfg       PropagatorConfig
	flowCache map[string]*FlowTraceRecord // key: flow 5-tuple string

	// Metrics
	TracesCreated    uint64
	TracesPropagated uint64
	TracesSampled    uint64
}

// NewPropagator creates a W3C Trace Context propagator.
func NewPropagator(cfg PropagatorConfig) *Propagator {
	if cfg.SampleRate <= 0 {
		cfg.SampleRate = 1.0
	}
	if cfg.ServiceName == "" {
		cfg.ServiceName = "straitgateway"
	}
	if cfg.MaxFlowCacheSize <= 0 {
		cfg.MaxFlowCacheSize = 65536
	}
	if cfg.FlowCacheTTL <= 0 {
		cfg.FlowCacheTTL = 5 * time.Minute
	}
	return &Propagator{
		cfg:       cfg,
		flowCache: make(map[string]*FlowTraceRecord),
	}
}

// TextMapCarrier abstracts HTTP headers / gRPC metadata for reading/writing.
type TextMapCarrier interface {
	Get(key string) string
	Set(key string, value string)
}

// Extract reads the incoming traceparent/tracestate from carrier headers.
// If no valid traceparent is present, a new SpanContext is generated.
// The resulting SpanContext represents the upstream context (parent span).
func (p *Propagator) Extract(flowKey string, carrier TextMapCarrier) SpanContext {
	p.mu.Lock()
	defer p.mu.Unlock()

	raw := carrier.Get("traceparent")
	var sc SpanContext
	var ok bool

	if raw != "" {
		sc, ok = ParseTraceparent(raw)
	}

	if !ok || !sc.IsValid() {
		// Generate a new trace — this is the ingress point
		sc = p.newSpanContext()
		p.TracesCreated++
	} else {
		p.TracesPropagated++
	}

	ts := carrier.Get("tracestate")
	if ts != "" {
		sc.TraceState = ParseTracestate(ts)
	}

	// Store in flow cache
	if len(p.flowCache) < p.cfg.MaxFlowCacheSize {
		p.flowCache[flowKey] = &FlowTraceRecord{
			FlowKey:     flowKey,
			SpanContext: sc,
			IngressAt:   time.Now(),
		}
	}

	if sc.Sampled() {
		p.TracesSampled++
	}

	return sc
}

// Inject writes a new child span's traceparent/tracestate into the outgoing
// carrier. A fresh SpanID is minted for StraitGateway's own processing span;
// the TraceID from the upstream context is preserved.
func (p *Propagator) Inject(parent SpanContext, carrier TextMapCarrier) SpanContext {
	// Mint a new span-id for our processing hop
	childSpanID := newSpanID()

	child := SpanContext{
		TraceID:    parent.TraceID,
		SpanID:     childSpanID,
		Flags:      parent.Flags,
		TraceState: p.buildTraceState(parent.TraceState),
	}

	carrier.Set("traceparent", child.Traceparent())
	if child.TraceState != "" {
		carrier.Set("tracestate", child.TraceState)
	}

	return child
}

// AnnotateFlow adds trace context to the flow cache record identified by flowKey,
// and returns the FlowTraceRecord for embedding in flow log entries.
func (p *Propagator) AnnotateFlow(flowKey string, sc SpanContext) *FlowTraceRecord {
	p.mu.Lock()
	defer p.mu.Unlock()

	rec, ok := p.flowCache[flowKey]
	if !ok {
		rec = &FlowTraceRecord{
			FlowKey:     flowKey,
			SpanContext: sc,
			IngressAt:   time.Now(),
		}
		if len(p.flowCache) < p.cfg.MaxFlowCacheSize {
			p.flowCache[flowKey] = rec
		}
	}
	rec.EgressAt = time.Now()
	return rec
}

// GetFlowTrace returns the trace record for a flow, if present.
func (p *Propagator) GetFlowTrace(flowKey string) (*FlowTraceRecord, bool) {
	p.mu.RLock()
	defer p.mu.RUnlock()
	rec, ok := p.flowCache[flowKey]
	return rec, ok
}

// EvictExpired removes flow trace entries older than FlowCacheTTL.
// Should be called periodically (e.g., from a 30s ticker in the agent).
func (p *Propagator) EvictExpired() int {
	p.mu.Lock()
	defer p.mu.Unlock()

	cutoff := time.Now().Add(-p.cfg.FlowCacheTTL)
	n := 0
	for key, rec := range p.flowCache {
		if rec.IngressAt.Before(cutoff) {
			delete(p.flowCache, key)
			n++
		}
	}
	return n
}

// FlowCacheSize returns the current number of cached flow→trace mappings.
func (p *Propagator) FlowCacheSize() int {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return len(p.flowCache)
}

// ---------------------------------------------------------------------------
// MapCarrier — a simple map-backed TextMapCarrier for HTTP headers.
// ---------------------------------------------------------------------------

// MapCarrier implements TextMapCarrier over a string map (e.g., http.Header).
type MapCarrier map[string]string

func (m MapCarrier) Get(key string) string {
	// Case-insensitive lookup
	k := strings.ToLower(key)
	for ck, v := range m {
		if strings.ToLower(ck) == k {
			return v
		}
	}
	return ""
}

func (m MapCarrier) Set(key, value string) {
	m[key] = value
}

// ---------------------------------------------------------------------------
// Internal helpers
// ---------------------------------------------------------------------------

func (p *Propagator) newSpanContext() SpanContext {
	flags := TraceFlags(0)
	if p.shouldSample() {
		flags |= TraceFlagSampled
	}
	return SpanContext{
		TraceID: newTraceID(),
		SpanID:  newSpanID(),
		Flags:   flags,
	}
}

func (p *Propagator) buildTraceState(existing string) string {
	// Prepend our vendor entry: sg=<serviceName>
	// W3C spec: max 32 list-members, each max 256 chars
	entry := "sg=" + p.cfg.ServiceName
	if existing == "" {
		return entry
	}
	// Prepend (leftmost = most recent hop per spec)
	return entry + "," + existing
}

func (p *Propagator) shouldSample() bool {
	if p.cfg.SampleRate >= 1.0 {
		return true
	}
	if p.cfg.SampleRate <= 0 {
		return false
	}
	// Use a random byte for probabilistic sampling
	b := make([]byte, 1)
	_, _ = rand.Read(b)
	return float64(b[0])/255.0 < p.cfg.SampleRate
}

// newTraceID generates a cryptographically random 128-bit trace ID (32 hex chars).
func newTraceID() string {
	b := make([]byte, 16)
	_, _ = rand.Read(b)
	return hex.EncodeToString(b)
}

// newSpanID generates a cryptographically random 64-bit span ID (16 hex chars).
func newSpanID() string {
	b := make([]byte, 8)
	_, _ = rand.Read(b)
	return hex.EncodeToString(b)
}

// isHex returns true if all characters in s are valid hexadecimal digits.
func isHex(s string) bool {
	for _, c := range s {
		if !((c >= '0' && c <= '9') || (c >= 'a' && c <= 'f') || (c >= 'A' && c <= 'F')) {
			return false
		}
	}
	return true
}

// FlowKey builds a canonical 5-tuple flow key string.
func FlowKey(srcIP, dstIP string, srcPort, dstPort uint16, proto string) string {
	return fmt.Sprintf("%s:%d→%s:%d/%s", srcIP, srcPort, dstIP, dstPort, proto)
}
