import { Component, inject, signal, DestroyRef } from '@angular/core';
import { takeUntilDestroyed } from '@angular/core/rxjs-interop';
import { ReactiveFormsModule, FormBuilder, Validators } from '@angular/forms';
import { PrometheusApiService } from '../../core/api/observability/prometheus.api';
import { GrafanaApiService } from '../../core/api/observability/grafana.api';
import { BreadcrumbsComponent } from '../../layout/breadcrumbs/breadcrumbs.component';
import { LineChartComponent } from '../../shared/charts/line-chart/line-chart.component';
import { SkeletonComponent } from '../../shared/components/skeleton/skeleton.component';
import { ErrorStateComponent } from '../../shared/components/error-state/error-state.component';
import { MetricQueryResult } from '../../core/models/metric.model';
import { ApiError } from '../../core/api/api-error';

const PRESET_RANGES = [
  { label: '15m', minutes: 15,  step: '30s' },
  { label: '1h',  minutes: 60,  step: '1m'  },
  { label: '3h',  minutes: 180, step: '3m'  },
  { label: '6h',  minutes: 360, step: '5m'  },
  { label: '24h', minutes: 1440,step: '15m' },
];

@Component({
  selector: 'sg-metrics',
  imports: [ReactiveFormsModule, BreadcrumbsComponent, LineChartComponent, SkeletonComponent, ErrorStateComponent],
  template: `
    <div class="sg-page">
      <sg-breadcrumbs />
      <header class="sg-page-header">
        <h1 class="sg-page-title">Metrics</h1>
        <div class="sg-header-actions">
          @for (preset of presets; track preset.label) {
            <button
              class="sg-btn sg-btn-secondary"
              [class.sg-btn-active]="activePreset() === preset.label"
              (click)="applyPreset(preset)"
              [attr.aria-pressed]="activePreset() === preset.label"
            >{{ preset.label }}</button>
          }
          <a [href]="grafanaUrl()" target="_blank" rel="noopener noreferrer" class="sg-btn sg-btn-secondary">
            Open in Grafana ↗
          </a>
        </div>
      </header>

      <!-- Custom PromQL -->
      <section class="sg-card sg-custom-query" aria-labelledby="custom-query-heading">
        <h2 id="custom-query-heading" class="sg-section-title">Custom Query</h2>
        <form [formGroup]="queryForm" (ngSubmit)="runCustomQuery()" class="sg-query-form">
          <input
            class="sg-input sg-font-mono"
            formControlName="query"
            placeholder="Enter PromQL query…"
            aria-label="PromQL expression"
          />
          <button class="sg-btn sg-btn-primary" type="submit" [disabled]="queryForm.invalid || querying()">
            {{ querying() ? 'Running…' : 'Run Query' }}
          </button>
        </form>
        @if (queryError()) {
          <sg-error-state [error]="queryError()" />
        }
        @if (customResult()) {
          <sg-line-chart [series]="customResult()!.series" [label]="queryForm.value.query ?? 'Custom'" />
        }
      </section>

      <!-- Standard metric charts -->
      <div class="sg-metrics-grid">
        @for (chart of metricCharts(); track chart.label) {
          <section class="sg-card" [attr.aria-labelledby]="'chart-' + chart.id">
            <h2 [id]="'chart-' + chart.id" class="sg-section-title">{{ chart.label }}</h2>
            @if (chart.loading) {
              <sg-skeleton variant="block" height="180px" />
            } @else if (chart.error) {
              <p class="sg-chart-error" role="alert">Metrics unavailable: {{ chart.error }}</p>
            } @else {
              <sg-line-chart [series]="chart.series" [label]="chart.label" [unit]="chart.unit" />
            }
          </section>
        }
      </div>
    </div>
  `,
  styles: [`
    .sg-section-title{font-size:.8125rem;font-weight:600;color:var(--sg-text-secondary);text-transform:uppercase;letter-spacing:.05em;margin-bottom:.75rem;}
    .sg-header-actions{display:flex;gap:6px;flex-wrap:wrap;align-items:center;}
    .sg-btn-active{border-color:var(--sg-accent);color:var(--sg-accent);}
    .sg-query-form{display:flex;gap:8px;margin-bottom:.5rem;}
    .sg-query-form .sg-input{flex:1;}
    .sg-metrics-grid{display:grid;grid-template-columns:repeat(auto-fill,minmax(360px,1fr));gap:16px;}
    .sg-chart-error{color:var(--sg-failed);font-size:.8125rem;}
    .sg-font-mono{font-family:var(--sg-font-mono);}
  `],
})
export class MetricsComponent {
  private readonly prometheusApi = inject(PrometheusApiService);
  private readonly grafanaApi    = inject(GrafanaApiService);
  private readonly destroyRef    = inject(DestroyRef);
  private readonly fb            = inject(FormBuilder);

  readonly presets     = PRESET_RANGES;
  readonly activePreset = signal('1h');
  readonly querying    = signal(false);
  readonly queryError  = signal<ApiError | null>(null);
  readonly customResult = signal<MetricQueryResult | null>(null);

  readonly queryForm = this.fb.nonNullable.group({
    query: ['', Validators.required],
  });

  grafanaUrl = signal(this.grafanaApi.dashboardUrl('straitgateway-overview'));

  readonly metricCharts = signal<Array<{
    id: string; label: string; unit: string;
    loading: boolean; error: string | null;
    series: MetricQueryResult['series'];
  }>>([
    { id: 'rx', label: 'RX Bytes/s', unit: 'B/s', loading: true, error: null, series: [] },
    { id: 'tx', label: 'TX Bytes/s', unit: 'B/s', loading: true, error: null, series: [] },
    { id: 'pps', label: 'Packets/s', unit: 'pps', loading: true, error: null, series: [] },
    { id: 'err', label: 'Error Rate', unit: '%',  loading: true, error: null, series: [] },
    { id: 'flows', label: 'Active Flows', unit: '',  loading: true, error: null, series: [] },
    { id: 'lat', label: 'Tunnel Latency', unit: 'ms', loading: true, error: null, series: [] },
  ]);

  private readonly CHART_QUERIES: Record<string, string> = {
    rx:    'sum(rate(straitgateway_traffic_rx_bytes_total[1m]))',
    tx:    'sum(rate(straitgateway_traffic_tx_bytes_total[1m]))',
    pps:   'sum(rate(straitgateway_packet_rate[1m]))',
    err:   'sum(rate(straitgateway_error_rate[1m])) * 100',
    flows: 'straitgateway_active_flows',
    lat:   'histogram_quantile(0.99, straitgateway_tunnel_latency_ms_bucket)',
  };

  constructor() { this.applyPreset(PRESET_RANGES[1]); }

  applyPreset(preset: typeof PRESET_RANGES[0]): void {
    this.activePreset.set(preset.label);
    const now   = new Date();
    const start = new Date(now.getTime() - preset.minutes * 60_000).toISOString();
    const end   = now.toISOString();
    this.metricCharts.update((charts) => charts.map((c) => ({ ...c, loading: true, error: null })));
    this.metricCharts().forEach((c) => {
      const query = this.CHART_QUERIES[c.id];
      if (!query) return;
      this.prometheusApi.queryRange({ query, start, end, step: preset.step })
        .pipe(takeUntilDestroyed(this.destroyRef))
        .subscribe({
          next:  (r)   => this.metricCharts.update((charts) => charts.map((ch) => ch.id === c.id ? { ...ch, series: r.series, loading: false } : ch)),
          error: (err) => this.metricCharts.update((charts) => charts.map((ch) => ch.id === c.id ? { ...ch, error: err?.message ?? 'Failed', loading: false } : ch)),
        });
    });
  }

  runCustomQuery(): void {
    const q = this.queryForm.getRawValue().query;
    if (!q) return;
    this.querying.set(true);
    this.queryError.set(null);
    const now   = new Date();
    const start = new Date(now.getTime() - 3600_000).toISOString();
    this.prometheusApi.queryRange({ query: q, start, end: now.toISOString(), step: '1m' })
      .pipe(takeUntilDestroyed(this.destroyRef))
      .subscribe({
        next:  (r)   => { this.customResult.set(r); this.querying.set(false); },
        error: (err) => { this.queryError.set(err); this.querying.set(false); },
      });
  }
}
