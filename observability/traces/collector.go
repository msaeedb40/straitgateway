// Copyright 2026 straitgateway Authors — SPDX-License-Identifier: Apache-2.0

// Package traces provides an in-process trace span collector for the dashboard.
// Completed spans are buffered in a ring buffer and served via the metrics API.
// For production export, use observability/tracing (OTLP exporter).
package traces

import (
	"sync"
	"time"

	"go.opentelemetry.io/otel/codes"
	"go.uber.org/zap"
)

// SpanRecord is a completed trace span stored in the collector ring buffer.
type SpanRecord struct {
	TraceID   string        `json:"traceId"`
	SpanID    string        `json:"spanId"`
	ParentID  string        `json:"parentId,omitempty"`
	Operation string        `json:"operation"`
	Service   string        `json:"service"`
	StartTime time.Time     `json:"startTime"`
	EndTime   time.Time     `json:"endTime"`
	Duration  time.Duration `json:"durationNs"`
	Status    codes.Code    `json:"status"`
	Error     string        `json:"error,omitempty"`
	Attrs     map[string]string `json:"attrs,omitempty"`
}

// Collector buffers completed spans for the dashboard.
type Collector struct {
	log    *zap.Logger
	mu     sync.RWMutex
	spans  []*SpanRecord
	cap    int
}

// New creates a new span Collector with the given buffer capacity.
func New(log *zap.Logger, capacity int) *Collector {
	return &Collector{log: log, cap: capacity}
}

// Record adds a completed span to the buffer.
func (c *Collector) Record(span *SpanRecord) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if len(c.spans) >= c.cap {
		c.spans = c.spans[1:]
	}
	c.spans = append(c.spans, span)
}

// Recent returns the n most recently recorded spans (newest-first).
func (c *Collector) Recent(n int) []*SpanRecord {
	c.mu.RLock()
	defer c.mu.RUnlock()
	if n <= 0 || len(c.spans) == 0 {
		return nil
	}
	start := len(c.spans) - n
	if start < 0 {
		start = 0
	}
	out := make([]*SpanRecord, len(c.spans)-start)
	copy(out, c.spans[start:])
	for i, j := 0, len(out)-1; i < j; i, j = i+1, j-1 {
		out[i], out[j] = out[j], out[i]
	}
	return out
}

// Len returns the number of buffered spans.
func (c *Collector) Len() int {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return len(c.spans)
}
