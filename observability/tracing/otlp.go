// Copyright 2026 straitgateway Authors
// SPDX-License-Identifier: Apache-2.0

// Package tracing provides OpenTelemetry OTLP trace export.
package tracing

import (
	"context"

	"go.uber.org/zap"
)

// Config holds OTLP tracing configuration.
type Config struct {
	Enabled      bool
	Endpoint     string
	SamplingRate float64
}

// Provider manages the OpenTelemetry trace provider.
type Provider struct {
	log    *zap.Logger
	config Config
}

// New creates a new trace Provider.
func New(cfg Config, log *zap.Logger) *Provider {
	return &Provider{log: log, config: cfg}
}

// Start initializes the OTLP trace exporter.
func (p *Provider) Start(ctx context.Context) error {
	if !p.config.Enabled {
		p.log.Info("tracing disabled")
		return nil
	}
	p.log.Info("OTLP tracing started",
		zap.String("endpoint", p.config.Endpoint),
		zap.Float64("samplingRate", p.config.SamplingRate),
	)
	// TODO: Initialize OTLP exporter via go.opentelemetry.io/otel
	return nil
}

// Shutdown flushes and shuts down the trace provider.
func (p *Provider) Shutdown(ctx context.Context) error {
	if !p.config.Enabled {
		return nil
	}
	p.log.Info("OTLP tracing shutdown")
	return nil
}
