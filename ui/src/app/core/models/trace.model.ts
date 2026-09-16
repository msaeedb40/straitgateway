export type TraceStatus = 'OK' | 'Error' | 'Unset';

export interface TraceReference {
  readonly traceId: string;
  readonly spanId: string;
  readonly operationName: string;
  readonly serviceName: string;
  readonly startTime: string;
  readonly durationMs: number;
  readonly status: TraceStatus;
  readonly spanCount: number;
  readonly errorCount: number;
}

export interface Span {
  readonly spanId: string;
  readonly parentSpanId: string | null;
  readonly traceId: string;
  readonly operationName: string;
  readonly serviceName: string;
  readonly startTime: string;
  readonly durationMs: number;
  readonly status: TraceStatus;
  readonly tags: Record<string, string>;
  readonly logs: SpanLog[];
}

export interface SpanLog {
  readonly timestamp: string;
  readonly fields: Record<string, string>;
}

export interface Trace {
  readonly traceId: string;
  readonly spans: Span[];
  readonly durationMs: number;
  readonly status: TraceStatus;
  readonly services: string[];
  readonly rootSpan: Span;
}

export interface TraceFilter {
  service?: string;
  operation?: string;
  tags?: Record<string, string>;
  minDurationMs?: number;
  maxDurationMs?: number;
  since?: string;
  until?: string;
  limit?: number;
}
