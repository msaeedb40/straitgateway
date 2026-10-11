package observability

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"
)

// Server exposes metrics, flow logs, and program status via HTTP
type Server struct {
	bindAddr    string
	metrics     *MetricsCollector
	flows       *FlowCollector
	ebpfStatus  *EbpfStatusCollector
	httpServer  *http.Server
	corsEnabled bool
}

// NewServer creates an Observability HTTP server
func NewServer(
	bindAddr string,
	metrics *MetricsCollector,
	flows *FlowCollector,
	ebpfStatus *EbpfStatusCollector,
) *Server {
	if bindAddr == "" {
		bindAddr = ":9090"
	}
	if ebpfStatus == nil {
		ebpfStatus = NewEbpfStatusCollector()
	}

	s := &Server{
		bindAddr:    bindAddr,
		metrics:     metrics,
		flows:       flows,
		ebpfStatus:  ebpfStatus,
		corsEnabled: true,
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/metrics", s.handlePrometheusMetrics)
	mux.HandleFunc("/observability/metrics", s.handleJSONMetrics)
	mux.HandleFunc("/api/v1/observability/metrics", s.handleJSONMetrics)
	mux.HandleFunc("/observability/flows", s.handleFlows)
	mux.HandleFunc("/api/v1/observability/flows", s.handleFlows)
	mux.HandleFunc("/observability/ebpf/programs", s.handleEbpfPrograms)
	mux.HandleFunc("/api/v1/observability/ebpf/programs", s.handleEbpfPrograms)
	mux.HandleFunc("/healthz", s.handleHealthz)

	// UI REST API Endpoints
	mux.HandleFunc("/api/v1/cluster/overview", s.handleClusterOverview)
	mux.HandleFunc("/cluster/overview", s.handleClusterOverview)
	mux.HandleFunc("/api/v1/cluster/health", s.handleClusterHealth)
	mux.HandleFunc("/cluster/health", s.handleClusterHealth)
	mux.HandleFunc("/api/v1/nodes", s.handleNodes)
	mux.HandleFunc("/nodes", s.handleNodes)
	mux.HandleFunc("/api/v1/pods", s.handlePods)
	mux.HandleFunc("/pods", s.handlePods)
	mux.HandleFunc("/api/v1/services", s.handleServices)
	mux.HandleFunc("/services", s.handleServices)
	mux.HandleFunc("/api/v1/gateways", s.handleGateways)
	mux.HandleFunc("/gateways", s.handleGateways)
	mux.HandleFunc("/api/v1/transitgateways", s.handleTransitGateways)
	mux.HandleFunc("/transitgateways", s.handleTransitGateways)
	mux.HandleFunc("/api/v1/httproutes", s.handleHttpRoutes)
	mux.HandleFunc("/httproutes", s.handleHttpRoutes)
	mux.HandleFunc("/api/v1/routing/routes", s.handleRoutes)
	mux.HandleFunc("/routing/routes", s.handleRoutes)
	mux.HandleFunc("/api/v1/routing/bgp/peers", s.handleBgpPeers)
	mux.HandleFunc("/routing/bgp/peers", s.handleBgpPeers)
	mux.HandleFunc("/api/v1/routing/bfd/sessions", s.handleBfdSessions)
	mux.HandleFunc("/routing/bfd/sessions", s.handleBfdSessions)
	mux.HandleFunc("/api/v1/wireguard/peers", s.handleWireguardPeers)
	mux.HandleFunc("/wireguard/peers", s.handleWireguardPeers)
	mux.HandleFunc("/api/v1/security/policies", s.handleSecurityPolicies)
	mux.HandleFunc("/security/policies", s.handleSecurityPolicies)

	s.httpServer = &http.Server{
		Addr:         bindAddr,
		Handler:      s.withCORS(mux),
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	return s
}

// Start launches the HTTP server
func (s *Server) Start() error {
	return s.httpServer.ListenAndServe()
}

// Stop gracefully shuts down the server
func (s *Server) Stop(ctx context.Context) error {
	return s.httpServer.Shutdown(ctx)
}

func (s *Server) withCORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func (s *Server) handlePrometheusMetrics(w http.ResponseWriter, r *http.Request) {
	if s.metrics == nil {
		http.Error(w, "metrics collector unavailable", http.StatusServiceUnavailable)
		return
	}
	w.Header().Set("Content-Type", "text/plain; version=0.0.4")
	_, _ = w.Write([]byte(s.metrics.PrometheusExport()))
}

func (s *Server) handleJSONMetrics(w http.ResponseWriter, r *http.Request) {
	if s.metrics == nil {
		http.Error(w, "metrics collector unavailable", http.StatusServiceUnavailable)
		return
	}
	data := s.metrics.Collect()
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(data)
}

func (s *Server) handleFlows(w http.ResponseWriter, r *http.Request) {
	if s.flows == nil {
		http.Error(w, "flow collector unavailable", http.StatusServiceUnavailable)
		return
	}

	limit := 100
	if l := r.URL.Query().Get("limit"); l != "" {
		if val, err := strconv.Atoi(l); err == nil && val > 0 {
			limit = val
		}
	}

	recent := s.flows.GetRecentFlows(limit)
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(recent)
}

func (s *Server) handleEbpfPrograms(w http.ResponseWriter, r *http.Request) {
	if s.ebpfStatus == nil {
		http.Error(w, "ebpf status collector unavailable", http.StatusServiceUnavailable)
		return
	}

	programs := s.ebpfStatus.GetProgramStatuses()
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(programs)
}

func (s *Server) handleHealthz(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	_, _ = fmt.Fprintf(w, `{"status":"ok","timestamp":"%s"}`+"\n", time.Now().UTC().Format(time.RFC3339))
}
