export interface MetricSample {
  readonly timestamp: number; // Unix epoch seconds
  readonly value: number;
}

export interface MetricSeries {
  readonly labels: Record<string, string>;
  readonly samples: MetricSample[];
}

export interface MetricQueryResult {
  readonly metric: string;
  readonly series: MetricSeries[];
}

export interface MetricQueryParams {
  readonly query: string;
  readonly start: string; // RFC3339
  readonly end: string;   // RFC3339
  readonly step: string;  // e.g. "30s", "1m", "5m"
}

export interface InstantQueryResult {
  readonly metric: string;
  readonly labels: Record<string, string>;
  readonly value: number;
  readonly timestamp: number;
}

export type MetricTarget =
  | 'gateway_traffic_rx_bytes_total'
  | 'gateway_traffic_tx_bytes_total'
  | 'gateway_packet_rate'
  | 'gateway_error_rate'
  | 'tunnel_latency_ms'
  | 'tunnel_active_flows'
  | 'node_cpu_usage'
  | 'node_memory_usage'
  | 'node_network_rx_bytes'
  | 'node_network_tx_bytes'
  | 'active_connections'
  | 'active_flows';
