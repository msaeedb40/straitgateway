export interface RuntimeApiConfig {
  readonly controller: string;
}

export interface RuntimeObservabilityConfig {
  readonly prometheus: string;
  readonly grafana: string;
  readonly jaeger: string;
  readonly logs: string;
}

export interface RuntimeFeaturesConfig {
  readonly packetCapture: boolean;
  readonly ebpf: boolean;
  readonly topology: boolean;
}

export interface RuntimeUiConfig {
  readonly refreshInterval: number;
}

export interface RuntimeConfig {
  readonly api: RuntimeApiConfig;
  readonly observability: RuntimeObservabilityConfig;
  readonly features: RuntimeFeaturesConfig;
  readonly ui: RuntimeUiConfig;
}
