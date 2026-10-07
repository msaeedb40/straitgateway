package observability

import (
	"testing"
	"time"

	"github.com/straitgateway/straitgateway/internal/mesh"
	"github.com/straitgateway/straitgateway/pkg/types"
)

func TestFlowCollector(t *testing.T) {
	ie := mesh.NewIdentityEngine()
	fc := NewFlowCollector(5, ie, "node-1")

	// 1. Ingest raw TelemetryEvent
	rawEv := types.TelemetryEvent{
		Timestamp: uint64(time.Now().UnixNano()),
		SrcIP:     0x0af4010a, // 10.244.1.10
		DstIP:     0x0af40214, // 10.244.2.20
		SrcPort:   45120,
		DstPort:   8080,
		Proto:     6, // TCP
		SrcID:     1001,
		DstID:     1002,
		Action:    1, // ALLOW
		Bytes:     1024,
		Packets:   1,
	}

	entry := fc.IngestRawTelemetry(rawEv)
	if entry.SourceIP != "10.244.1.10" {
		t.Errorf("expected source IP 10.244.1.10, got %s", entry.SourceIP)
	}
	if entry.DestinationIP != "10.244.2.20" {
		t.Errorf("expected destination IP 10.244.2.20, got %s", entry.DestinationIP)
	}
	if entry.Action != "ALLOW" {
		t.Errorf("expected action ALLOW, got %s", entry.Action)
	}

	// 2. Pub/Sub streaming
	subCh := fc.Subscribe(10)
	defer fc.Unsubscribe(subCh)

	flow2 := FlowLogEntry{
		ID:        "flow-2",
		SourceIP:  "192.168.1.5",
		Action:    "DENY",
		Protocol:  "UDP",
		Timestamp: time.Now(),
	}
	fc.RecordFlow(flow2)

	select {
	case received := <-subCh:
		if received.ID != "flow-2" || received.Action != "DENY" {
			t.Errorf("unexpected flow received on subscription channel: %+v", received)
		}
	case <-time.After(100 * time.Millisecond):
		t.Fatal("timed out waiting for flow on subscription channel")
	}

	// 3. Sliding window buffer cap
	for i := 0; i < 10; i++ {
		fc.RecordFlow(FlowLogEntry{ID: "overflow"})
	}

	recent := fc.GetRecentFlows(10)
	if len(recent) != 5 { // maxEntries = 5
		t.Errorf("expected buffer capped at 5 entries, got %d", len(recent))
	}
}
