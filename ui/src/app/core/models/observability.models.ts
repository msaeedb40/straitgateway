// Copyright 2026 straitgateway Authors
// SPDX-License-Identifier: Apache-2.0

export interface FlowEvent {
  srcIP: string;
  dstIP: string;
  srcPort: number;
  dstPort: number;
  srcIdentity: number;
  dstIdentity: number;
  protocol: string;
  direction: 'ingress' | 'egress';
  action: 'allow' | 'deny' | 'reject';
  bytes: number;
  timestampNs: number;
  dropReason?: string;
}

export interface PacketCapture {
  id: string;
  timestamp: string;
  srcIP: string;
  dstIP: string;
  srcPort: number;
  dstPort: number;
  protocol: string;
  length: number;
  action: 'allowed' | 'dropped' | 'forwarded';
  reason?: string;
  hexPreview?: string;
}

export interface MetricSample {
  timestamp: number;
  value: number;
  label?: string;
}

export interface LogEntry {
  timestamp: string;
  level: 'DEBUG' | 'INFO' | 'WARN' | 'ERROR';
  component: string;
  message: string;
  subsystem?: string;
}

export interface TraceSpan {
  traceId: string;
  spanId: string;
  operationName: string;
  serviceName: string;
  durationMs: number;
  startTime: string;
  status: 'ok' | 'error';
  tags?: Record<string, string>;
}
