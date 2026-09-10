// Copyright 2026 straitgateway Authors
// SPDX-License-Identifier: Apache-2.0

export interface RuntimeConfig {
  /** Base URL for the straitgateway controller API. */
  apiBase: string;
  /** Base URL for Prometheus. */
  prometheusBase: string;
  /** Base URL for Grafana. */
  grafanaBase: string;
  /** Base URL for Jaeger. */
  jaegerBase: string;
  /** Cluster name shown in the UI header. */
  clusterName: string;
  /** Refresh interval for auto-refreshing views (ms). */
  refreshIntervalMs: number;
  /** Standalone mock mode when backend is offline. */
  mockMode?: boolean;
}

export const DEFAULT_RUNTIME_CONFIG: RuntimeConfig = {
  apiBase: 'http://localhost:8080',
  prometheusBase: 'http://localhost:9090',
  grafanaBase: 'http://localhost:3000',
  jaegerBase: 'http://localhost:16686',
  clusterName: 'dev',
  refreshIntervalMs: 15000,
  mockMode: false,
};
