import { inject, Injectable, signal } from '@angular/core';
import { interval, Subscription } from 'rxjs';
import { BackendHealth, ConnectionStatus } from '../api/api.types';
import { PrometheusApiService } from '../api/observability/prometheus.api';
import { GrafanaApiService } from '../api/observability/grafana.api';
import { JaegerApiService } from '../api/observability/jaeger.api';
import { LogsApiService } from '../api/observability/logs.api';
import { ApiClient } from '../api/api-client';
import { RuntimeConfigService } from '../config/runtime-config.service';

@Injectable({ providedIn: 'root' })
export class ConnectionService {
  private readonly apiClient = inject(ApiClient);
  private readonly runtimeConfig = inject(RuntimeConfigService);
  private readonly prometheus = inject(PrometheusApiService);
  private readonly grafana = inject(GrafanaApiService);
  private readonly jaeger = inject(JaegerApiService);
  private readonly logs = inject(LogsApiService);

  readonly controller = signal<ConnectionStatus>('checking');
  readonly prometheusStatus = signal<ConnectionStatus>('checking');
  readonly grafanaStatus = signal<ConnectionStatus>('checking');
  readonly jaegerStatus = signal<ConnectionStatus>('checking');
  readonly logsStatus = signal<ConnectionStatus>('checking');

  private pollSub: Subscription | null = null;

  startPolling(): void {
    this.checkAll();
    this.pollSub = interval(this.runtimeConfig.refreshInterval).subscribe(() =>
      this.checkAll()
    );
  }

  stopPolling(): void {
    this.pollSub?.unsubscribe();
    this.pollSub = null;
  }

  get snapshot(): BackendHealth {
    return {
      controller: this.controller(),
      prometheus: this.prometheusStatus(),
      grafana: this.grafanaStatus(),
      jaeger: this.jaegerStatus(),
      logs: this.logsStatus(),
    };
  }

  private checkAll(): void {
    this.checkController();
    this.prometheus.healthCheck().subscribe({
      next: (ok) => this.prometheusStatus.set(ok ? 'connected' : 'unavailable'),
      error: () => this.prometheusStatus.set('unavailable'),
    });
    this.grafana.healthCheck().subscribe({
      next: (ok) => this.grafanaStatus.set(ok ? 'connected' : 'unavailable'),
      error: () => this.grafanaStatus.set('unavailable'),
    });
    this.jaeger.healthCheck().subscribe({
      next: (ok) => this.jaegerStatus.set(ok ? 'connected' : 'unavailable'),
      error: () => this.jaegerStatus.set('unavailable'),
    });
    this.logs.healthCheck().subscribe({
      next: (ok) => this.logsStatus.set(ok ? 'connected' : 'unavailable'),
      error: () => this.logsStatus.set('unavailable'),
    });
  }

  private checkController(): void {
    this.apiClient.get<unknown>('/healthz').subscribe({
      next: () => this.controller.set('connected'),
      error: () => this.controller.set('unavailable'),
    });
  }
}
