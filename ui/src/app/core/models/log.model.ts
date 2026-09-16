export type LogSeverity = 'DEBUG' | 'INFO' | 'WARN' | 'ERROR' | 'FATAL';

export interface LogEntry {
  readonly id: string;
  readonly timestamp: string;
  readonly severity: LogSeverity;
  readonly component: string;
  readonly nodeName?: string;
  readonly podName?: string;
  readonly namespace?: string;
  readonly message: string;
  /** Raw structured fields when log is JSON — null for plain text logs */
  readonly structured: Record<string, unknown> | null;
  readonly traceId?: string;
  readonly spanId?: string;
}

export interface LogFilter {
  namespace?: string;
  cluster?: string;
  severity?: LogSeverity;
  component?: string;
  nodeName?: string;
  podName?: string;
  search?: string;
  since?: string;
  until?: string;
  tailLines?: number;
}

export interface LogStreamConfig {
  readonly filter: LogFilter;
  readonly follow: boolean;
}
