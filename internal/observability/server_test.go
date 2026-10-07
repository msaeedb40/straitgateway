package observability

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/straitgateway/straitgateway/internal/mesh"
)

func TestObservabilityServerEndpoints(t *testing.T) {
	ie := mesh.NewIdentityEngine()
	metricsCol := NewMetricsCollector(nil, nil, nil, nil, "test-node")
	flowCol := NewFlowCollector(100, ie, "test-node")
	ebpfCol := NewEbpfStatusCollector()

	// Seed one sample flow
	flowCol.RecordFlow(FlowLogEntry{
		ID:            "flow-test-1",
		Timestamp:     time.Now().UTC(),
		SourceIP:      "10.244.0.5",
		DestinationIP: "10.244.0.10",
		Protocol:      "TCP",
		Action:        "ALLOW",
		Bytes:         512,
		Packets:       1,
	})

	server := NewServer(":0", metricsCol, flowCol, ebpfCol)
	ts := httptest.NewServer(server.withCORS(server.httpServer.Handler))
	defer ts.Close()

	client := ts.Client()

	// 1. Test /healthz
	resp, err := client.Get(ts.URL + "/healthz")
	if err != nil {
		t.Fatalf("healthz request failed: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected 200 from healthz, got %d", resp.StatusCode)
	}
	_ = resp.Body.Close()

	// 2. Test /metrics (Prometheus text)
	resp, err = client.Get(ts.URL + "/metrics")
	if err != nil {
		t.Fatalf("metrics request failed: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected 200 from metrics, got %d", resp.StatusCode)
	}
	body, _ := io.ReadAll(resp.Body)
	_ = resp.Body.Close()
	if len(body) == 0 {
		t.Error("expected non-empty prometheus metrics output")
	}

	// 3. Test /observability/metrics (JSON)
	resp, err = client.Get(ts.URL + "/observability/metrics")
	if err != nil {
		t.Fatalf("JSON metrics request failed: %v", err)
	}
	var metricsData DataplaneMetrics
	if err := json.NewDecoder(resp.Body).Decode(&metricsData); err != nil {
		t.Fatalf("failed to decode JSON metrics: %v", err)
	}
	_ = resp.Body.Close()

	// 4. Test /observability/flows (JSON array)
	resp, err = client.Get(ts.URL + "/observability/flows?limit=10")
	if err != nil {
		t.Fatalf("flows request failed: %v", err)
	}
	var flows []FlowLogEntry
	if err := json.NewDecoder(resp.Body).Decode(&flows); err != nil {
		t.Fatalf("failed to decode flows array: %v", err)
	}
	_ = resp.Body.Close()
	if len(flows) != 1 || flows[0].ID != "flow-test-1" {
		t.Errorf("expected 1 flow with ID flow-test-1, got %+v", flows)
	}

	// 5. Test /observability/ebpf/programs (JSON array)
	resp, err = client.Get(ts.URL + "/observability/ebpf/programs")
	if err != nil {
		t.Fatalf("ebpf programs request failed: %v", err)
	}
	var programs []EbpfProgramStatus
	if err := json.NewDecoder(resp.Body).Decode(&programs); err != nil {
		t.Fatalf("failed to decode ebpf programs array: %v", err)
	}
	_ = resp.Body.Close()
	if len(programs) < 5 {
		t.Errorf("expected at least 5 registered eBPF programs, got %d", len(programs))
	}

	// 6. Test CORS headers
	req, _ := http.NewRequest("OPTIONS", ts.URL+"/observability/metrics", nil)
	optionsResp, err := client.Do(req)
	if err != nil {
		t.Fatalf("CORS options request failed: %v", err)
	}
	if optionsResp.Header.Get("Access-Control-Allow-Origin") != "*" {
		t.Errorf("expected CORS header *, got %s", optionsResp.Header.Get("Access-Control-Allow-Origin"))
	}
	_ = optionsResp.Body.Close()

	// 7. Test UI REST Endpoints: Cluster Overview
	resp, err = client.Get(ts.URL + "/api/v1/cluster/overview")
	if err != nil {
		t.Fatalf("cluster overview failed: %v", err)
	}
	var overview ClusterOverview
	if err := json.NewDecoder(resp.Body).Decode(&overview); err != nil {
		t.Fatalf("failed to decode cluster overview: %v", err)
	}
	_ = resp.Body.Close()
	if overview.NodesCount != 8 || overview.Status != "Healthy" {
		t.Errorf("unexpected overview data: %+v", overview)
	}

	// 8. Test UI REST Endpoints: Nodes
	resp, err = client.Get(ts.URL + "/api/v1/nodes")
	if err != nil {
		t.Fatalf("nodes request failed: %v", err)
	}
	var nodes []NodeInfo
	if err := json.NewDecoder(resp.Body).Decode(&nodes); err != nil {
		t.Fatalf("failed to decode nodes: %v", err)
	}
	_ = resp.Body.Close()
	if len(nodes) == 0 {
		t.Error("expected non-empty nodes list")
	}

	// 9. Test UI REST Endpoints: Routes
	resp, err = client.Get(ts.URL + "/api/v1/routing/routes")
	if err != nil {
		t.Fatalf("routes request failed: %v", err)
	}
	var routes []RouteEntry
	if err := json.NewDecoder(resp.Body).Decode(&routes); err != nil {
		t.Fatalf("failed to decode routes: %v", err)
	}
	_ = resp.Body.Close()
	if len(routes) == 0 {
		t.Error("expected non-empty routes list")
	}
}
