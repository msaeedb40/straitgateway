// Package telemetry initializes Prometheus metrics and OpenTelemetry
// tracing/logging for StraitGateway components.
package telemetry

import (
	"context"
	"fmt"
	"net/http"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.uber.org/zap"
)

// Manager manages observability subsystems.
type Manager struct {
	log      *zap.Logger
	registry *prometheus.Registry
	addr     string
}

// NewManager creates a new telemetry Manager.
func NewManager(log *zap.Logger, addr string) *Manager {
	return &Manager{
		log:      log,
		registry: prometheus.NewRegistry(),
		addr:     addr,
	}
}

// Start starts the Prometheus metrics HTTP server.
func (m *Manager) Start(ctx context.Context) error {
	// Register default Go process metrics.
	m.registry.MustRegister(prometheus.NewGoCollector())
	m.registry.MustRegister(prometheus.NewProcessCollector(prometheus.ProcessCollectorOpts{}))

	mux := http.NewServeMux()
	mux.Handle("/metrics", promhttp.HandlerFor(m.registry, promhttp.HandlerOpts{
		EnableOpenMetrics: true,
	}))
	mux.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		fmt.Fprintln(w, "ok")
	})

	srv := &http.Server{Addr: m.addr, Handler: mux}

	go func() {
		m.log.Info("Prometheus metrics server starting", zap.String("addr", m.addr))
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			m.log.Error("metrics server error", zap.Error(err))
		}
	}()

	go func() {
		<-ctx.Done()
		_ = srv.Shutdown(context.Background())
	}()

	return nil
}

// Registry returns the Prometheus registry for registering custom metrics.
func (m *Manager) Registry() *prometheus.Registry {
	return m.registry
}
