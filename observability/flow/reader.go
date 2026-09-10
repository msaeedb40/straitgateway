// Copyright 2026 straitgateway Authors
// SPDX-License-Identifier: Apache-2.0

// Package flow reads flow events from the eBPF ring buffer.
// The ring buffer is populated by bpf/observability/trace.c.
//
// Architectural invariant: flow observation is strictly READ-ONLY.
// This package never writes to BPF maps or affects forwarding decisions.
package flow

import (
	"encoding/binary"
	"fmt"

	"github.com/cilium/ebpf/ringbuf"
	"go.uber.org/zap"
)

// Event represents a decoded flow event from the eBPF ring buffer.
type Event struct {
	SrcIP       uint32
	DstIP       uint32
	SrcPort     uint16
	DstPort     uint16
	Protocol    uint8
	Direction   uint8 // 0=ingress, 1=egress
	Action      uint8 // 0=allow, 1=deny, 2=reject
	SrcIdentity uint32
	DstIdentity uint32
	Bytes       uint64
	TimestampNs uint64
	DropReason  uint8
}

// Reader reads flow events from the eBPF ring buffer.
type Reader struct {
	log    *zap.Logger
	reader *ringbuf.Reader
}

// New creates a new flow event Reader.
func New(rb *ringbuf.Reader, log *zap.Logger) *Reader {
	return &Reader{reader: rb, log: log}
}

// Read blocks until a flow event is available and returns it.
func (r *Reader) Read() (*Event, error) {
	record, err := r.reader.Read()
	if err != nil {
		return nil, fmt.Errorf("reading ring buffer: %w", err)
	}

	if len(record.RawSample) < 40 {
		return nil, fmt.Errorf("flow event too small: %d bytes", len(record.RawSample))
	}

	ev := &Event{
		SrcIP:       binary.LittleEndian.Uint32(record.RawSample[0:4]),
		DstIP:       binary.LittleEndian.Uint32(record.RawSample[4:8]),
		SrcPort:     binary.LittleEndian.Uint16(record.RawSample[8:10]),
		DstPort:     binary.LittleEndian.Uint16(record.RawSample[10:12]),
		Protocol:    record.RawSample[12],
		Direction:   record.RawSample[13],
		Action:      record.RawSample[14],
		DropReason:  record.RawSample[15],
		SrcIdentity: binary.LittleEndian.Uint32(record.RawSample[16:20]),
		DstIdentity: binary.LittleEndian.Uint32(record.RawSample[20:24]),
		Bytes:       binary.LittleEndian.Uint64(record.RawSample[24:32]),
		TimestampNs: binary.LittleEndian.Uint64(record.RawSample[32:40]),
	}

	return ev, nil
}

// Close closes the ring buffer reader.
func (r *Reader) Close() error {
	return r.reader.Close()
}
