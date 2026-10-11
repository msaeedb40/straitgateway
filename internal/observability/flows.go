package observability

import (
	"fmt"
	"net"
	"sync"
	"time"

	"github.com/straitgateway/straitgateway/internal/mesh"
	"github.com/straitgateway/straitgateway/pkg/types"
)

// FlowCollector buffers and streams flow logs from the dataplane
type FlowCollector struct {
	mu          sync.RWMutex
	maxEntries  int
	flows       []FlowLogEntry
	subscribers map[chan FlowLogEntry]struct{}
	identEngine *mesh.IdentityEngine
	nodeName    string
}

// NewFlowCollector creates a FlowCollector
func NewFlowCollector(maxEntries int, ie *mesh.IdentityEngine, nodeName string) *FlowCollector {
	if maxEntries <= 0 {
		maxEntries = 1000
	}
	if nodeName == "" {
		nodeName = "localhost"
	}

	return &FlowCollector{
		maxEntries:  maxEntries,
		flows:       make([]FlowLogEntry, 0, maxEntries),
		subscribers: make(map[chan FlowLogEntry]struct{}),
		identEngine: ie,
		nodeName:    nodeName,
	}
}

// IngestRawTelemetry processes a TelemetryEvent from the BPF ring buffer
func (fc *FlowCollector) IngestRawTelemetry(ev types.TelemetryEvent) FlowLogEntry {
	srcIP := intToIP(ev.SrcIP).String()
	dstIP := intToIP(ev.DstIP).String()

	protoStr := "TCP"
	switch ev.Proto {
	case 17:
		protoStr = "UDP"
	case 1:
		protoStr = "ICMP"
	}

	actionStr := "ALLOW"
	switch ev.Action {
	case 0:
		actionStr = "DENY"
	case 2:
		actionStr = "AUDIT"
	}

	entry := FlowLogEntry{
		ID:                  fmt.Sprintf("flow-%d-%d", ev.Timestamp, ev.SrcPort),
		Timestamp:           time.Unix(0, int64(ev.Timestamp)).UTC(),
		SourceIP:            srcIP,
		DestinationIP:       dstIP,
		SourcePort:          ev.SrcPort,
		DestinationPort:     ev.DstPort,
		Protocol:            protoStr,
		SourceIdentity:      ev.SrcID,
		DestinationIdentity: ev.DstID,
		Action:              actionStr,
		Node:                fc.nodeName,
		Interface:           "netkit",
		Bytes:               ev.Bytes,
		Packets:             ev.Packets,
	}

	fc.RecordFlow(entry)
	return entry
}

// RecordFlow appends a flow log to the buffer and broadcasts to subscribers
func (fc *FlowCollector) RecordFlow(entry FlowLogEntry) {
	fc.mu.Lock()
	defer fc.mu.Unlock()

	if len(fc.flows) >= fc.maxEntries {
		fc.flows = fc.flows[1:]
	}
	fc.flows = append(fc.flows, entry)

	// Broadcast non-blocking
	for ch := range fc.subscribers {
		select {
		case ch <- entry:
		default:
		}
	}
}

// GetRecentFlows returns the last N flow records (up to limit)
func (fc *FlowCollector) GetRecentFlows(limit int) []FlowLogEntry {
	fc.mu.RLock()
	defer fc.mu.RUnlock()

	if limit <= 0 || limit > len(fc.flows) {
		limit = len(fc.flows)
	}

	start := len(fc.flows) - limit
	result := make([]FlowLogEntry, limit)
	copy(result, fc.flows[start:])

	// Return latest first
	for i, j := 0, len(result)-1; i < j; i, j = i+1, j-1 {
		result[i], result[j] = result[j], result[i]
	}

	return result
}

// Subscribe attaches a channel for real-time flow streaming
func (fc *FlowCollector) Subscribe(bufferSize int) chan FlowLogEntry {
	fc.mu.Lock()
	defer fc.mu.Unlock()

	ch := make(chan FlowLogEntry, bufferSize)
	fc.subscribers[ch] = struct{}{}
	return ch
}

// Unsubscribe removes a channel from listeners
func (fc *FlowCollector) Unsubscribe(ch chan FlowLogEntry) {
	fc.mu.Lock()
	defer fc.mu.Unlock()

	delete(fc.subscribers, ch)
	close(ch)
}

func intToIP(val uint32) net.IP {
	return net.IPv4(byte(val>>24), byte(val>>16), byte(val>>8), byte(val))
}
