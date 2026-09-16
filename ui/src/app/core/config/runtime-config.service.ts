import { HttpClient } from '@angular/common/http';
import { inject, Injectable, signal } from '@angular/core';
import { firstValueFrom } from 'rxjs';
import { RuntimeConfig } from './runtime-config.model';

const DEFAULT_CONFIG: RuntimeConfig = {
  api: {
    controller: '/api',
  },
  observability: {
    prometheus: '/prometheus',
    grafana: '/grafana',
    jaeger: '/jaeger',
    logs: '/logs',
  },
  features: {
    packetCapture: true,
    ebpf: true,
    topology: true,
  },
  ui: {
    refreshInterval: 5000,
  },
};

@Injectable({ providedIn: 'root' })
export class RuntimeConfigService {
  private readonly http = inject(HttpClient);
  private config: RuntimeConfig = DEFAULT_CONFIG;

  async load(): Promise<void> {
    try {
      const loaded = await firstValueFrom(
        this.http.get<RuntimeConfig>('/runtime-config.json')
      );
      if (loaded) {
        this.config = loaded;
      }
    } catch {
      // Fall back to DEFAULT_CONFIG during SSR, route extraction, or network error
    }
  }

  get controllerBase(): string {
    return this.config.api.controller;
  }

  get prometheusBase(): string {
    return this.config.observability.prometheus;
  }

  get grafanaBase(): string {
    return this.config.observability.grafana;
  }

  get jaegerBase(): string {
    return this.config.observability.jaeger;
  }

  get logsBase(): string {
    return this.config.observability.logs;
  }

  get features(): RuntimeConfig['features'] {
    return this.config.features;
  }

  get refreshInterval(): number {
    return this.config.ui.refreshInterval;
  }
}
