// Copyright 2026 straitgateway Authors — SPDX-License-Identifier: Apache-2.0

// Package logs provides structured log export for straitgateway components.
// All logs are emitted as structured JSON via go.uber.org/zap and can be
// forwarded to external log aggregators (Loki, Elasticsearch, etc.).
package logs

import (
	"encoding/json"
	"time"

	"go.uber.org/zap"
)

// Level represents a log severity level.
type Level string

const (
	LevelDebug Level = "debug"
	LevelInfo  Level = "info"
	LevelWarn  Level = "warn"
	LevelError Level = "error"
)

// LogEntry is the canonical structured log entry emitted by straitgateway.
// Follows the 11-attribute observability model.
type LogEntry struct {
	Timestamp   time.Time         `json:"timestamp"`
	Level       Level             `json:"level"`
	Caller      string            `json:"caller"`
	Message     string            `json:"msg"`
	ClusterName string            `json:"cluster_name,omitempty"`
	NodeName    string            `json:"node_name,omitempty"`
	Namespace   string            `json:"namespace,omitempty"`
	PodName     string            `json:"pod_name,omitempty"`
	Component   string            `json:"component,omitempty"`
	Fields      map[string]any    `json:"fields,omitempty"`
}

// Exporter exposes log lines for the dashboard API and log aggregators.
type Exporter struct {
	log    *zap.Logger
	buffer []*LogEntry
	cap    int
}

// New creates a new log Exporter with a ring-buffer of the given capacity.
func New(log *zap.Logger, capacity int) *Exporter {
	return &Exporter{log: log, cap: capacity}
}

// Append adds a log entry to the in-memory ring buffer.
func (e *Exporter) Append(entry *LogEntry) {
	if len(e.buffer) >= e.cap {
		e.buffer = e.buffer[1:] // evict oldest
	}
	e.buffer = append(e.buffer, entry)
}

// Recent returns the most recent n entries (newest-first).
func (e *Exporter) Recent(n int) []*LogEntry {
	if n <= 0 || len(e.buffer) == 0 {
		return nil
	}
	start := len(e.buffer) - n
	if start < 0 {
		start = 0
	}
	out := make([]*LogEntry, len(e.buffer)-start)
	copy(out, e.buffer[start:])
	// reverse for newest-first
	for i, j := 0, len(out)-1; i < j; i, j = i+1, j-1 {
		out[i], out[j] = out[j], out[i]
	}
	return out
}

// MarshalJSON serialises a LogEntry to JSON bytes.
func (e *LogEntry) MarshalJSON() ([]byte, error) {
	type Alias LogEntry
	return json.Marshal((*Alias)(e))
}
