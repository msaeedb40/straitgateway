import { Component, inject, OnInit, signal, computed, DestroyRef } from '@angular/core';
import { DecimalPipe, SlicePipe } from '@angular/common';
import { takeUntilDestroyed } from '@angular/core/rxjs-interop';
import { RouterLink } from '@angular/router';
import { forkJoin } from 'rxjs';
import { GatewayApiService } from '../../core/api/resources/gateway.api';
import { NodeApiService } from '../../core/api/resources/node.api';
import { TunnelApiService } from '../../core/api/resources/tunnel.api';
import { EventApiService } from '../../core/api/resources/event.api';
import { FlowApiService } from '../../core/api/resources/flow.api';
import { TopologyApiService } from '../../core/api/resources/topology.api';
import { ContextService } from '../../core/services/context.service';
import { ConnectionService } from '../../core/services/connection.service';
import { RuntimeConfigService } from '../../core/config/runtime-config.service';
import { PrometheusApiService } from '../../core/api/observability/prometheus.api';
import { SkeletonComponent, SkeletonListComponent } from '../../shared/components/skeleton/skeleton.component';
import { ErrorStateComponent } from '../../shared/components/error-state/error-state.component';
import { StatusBadgeComponent } from '../../shared/components/status-badge/status-badge.component';
import { LineChartComponent } from '../../shared/charts/line-chart/line-chart.component';
import { TopologyGraphComponent } from '../../shared/topology/topology-graph/topology-graph.component';
import { BreadcrumbsComponent } from '../../layout/breadcrumbs/breadcrumbs.component';
import { ApiError } from '../../core/api/api-error';
import { Gateway } from '../../core/models/gateway.model';
import { StraitEvent } from '../../core/models/event.model';
import { Flow } from '../../core/models/flow.model';
import { TopologyGraph } from '../../core/models/topology.model';
import { MetricQueryResult } from '../../core/models/metric.model';

@Component({
  selector: 'sg-dashboard',
  imports: [
    RouterLink, BreadcrumbsComponent, DecimalPipe, SlicePipe,
    SkeletonComponent, SkeletonListComponent, ErrorStateComponent,
    StatusBadgeComponent, LineChartComponent, TopologyGraphComponent,
  ],
  template: `
    <div class="sg-page">
      <sg-breadcrumbs />

      <header class="sg-page-header">
        <h1 class="sg-page-title">Dashboard</h1>
        <span class="sg-context-chip" role="status" aria-live="polite">
          {{ context.selectedCluster() ?? '—' }} / {{ context.selectedNamespace() || '—' }}
        </span>
      </header>

      <!-- Summary counters -->
      <section aria-labelledby="dashboard-summary-heading">
        <h2 id="dashboard-summary-heading" class="sg-sr-only">Resource summary</h2>
        <div class="sg-summary-grid">
          @for (card of summaryCards(); track card.label) {
            <a [routerLink]="card.route" class="sg-summary-card sg-card" [attr.aria-label]="card.label + ': ' + card.count">
              @if (loading()) {
                <sg-skeleton variant="line" width="40px" height="28px" />
                <sg-skeleton variant="line" width="80px" />
              } @else {
                <span class="sg-summary-count">{{ card.count }}</span>
                <span class="sg-summary-label">{{ card.label }}</span>
              }
            </a>
          }
        </div>
      </section>

      <div class="sg-dashboard-grid">
        <!-- Mini topology -->
        <section class="sg-card sg-dashboard-topology" aria-labelledby="dashboard-topology-heading">
          <h2 id="dashboard-topology-heading" class="sg-section-title">Network Topology</h2>
          @if (topologyLoading()) {
            <sg-skeleton variant="block" height="260px" />
          } @else if (topology()) {
            <div class="sg-mini-topology-canvas">
              <sg-topology-graph [graph]="topology()!" />
            </div>
          }
          <a routerLink="/topology" class="sg-card-link">View full topology →</a>
        </section>

        <!-- Traffic chart -->
        <section class="sg-card sg-dashboard-traffic" aria-labelledby="dashboard-traffic-heading">
          <h2 id="dashboard-traffic-heading" class="sg-section-title">Traffic Overview</h2>
          @if (trafficLoading()) {
            <sg-skeleton variant="block" height="180px" />
          } @else {
            <sg-line-chart
              [series]="trafficSeries()"
              label="Network traffic RX/TX bytes per second"
              unit="B/s"
            />
          }
          <a routerLink="/metrics" class="sg-card-link">View metrics →</a>
        </section>

        <!-- Active flows -->
        <section class="sg-card sg-dashboard-flows" aria-labelledby="dashboard-flows-heading">
          <h2 id="dashboard-flows-heading" class="sg-section-title">Active Flows</h2>
          @if (loading()) {
            <sg-skeleton-list [count]="5" />
          } @else if (flows().length === 0) {
            <p class="sg-empty-inline">No active flows</p>
          } @else {
            <table class="sg-table" aria-label="Active flows">
              <thead><tr>
                <th scope="col">Source</th>
                <th scope="col">Destination</th>
                <th scope="col">Protocol</th>
                <th scope="col">Rate</th>
                <th scope="col">State</th>
              </tr></thead>
              <tbody>
                @for (f of flows(); track f.id) {
                  <tr>
                    <td class="sg-mono-value">{{ f.sourceIP }}:{{ f.sourcePort }}</td>
                    <td class="sg-mono-value">{{ f.destinationIP }}:{{ f.destinationPort }}</td>
                    <td>{{ f.protocol }}</td>
                    <td class="sg-mono-value">{{ f.bytesPerSecond | number:'1.0-0' }} B/s</td>
                    <td><sg-status-badge [variant]="f.state === 'Active' ? 'healthy' : 'degraded'" /></td>
                  </tr>
                }
              </tbody>
            </table>
          }
          <a routerLink="/flows" class="sg-card-link">View all flows →</a>
        </section>

        <!-- Gateway health -->
        <section class="sg-card sg-dashboard-gateways" aria-labelledby="dashboard-gw-heading">
          <h2 id="dashboard-gw-heading" class="sg-section-title">Gateway Health</h2>
          @if (loading()) {
            <sg-skeleton-list [count]="4" />
          } @else {
            @for (gw of gateways(); track gw.metadata.name) {
              <div class="sg-gw-row">
                <a [routerLink]="['/gateways', gw.metadata.name]" class="sg-gw-name">{{ gw.metadata.name }}</a>
                <sg-status-badge [phase]="gw.status.phase" />
              </div>
            }
          }
          <a routerLink="/gateways" class="sg-card-link">View all gateways →</a>
        </section>

        <!-- Events feed -->
        <section class="sg-card sg-dashboard-events" aria-labelledby="dashboard-events-heading">
          <h2 id="dashboard-events-heading" class="sg-section-title">Recent Events</h2>
          @if (loading()) {
            <sg-skeleton-list [count]="5" />
          } @else {
            @for (ev of events(); track ev.uid) {
              <div class="sg-event-row">
                <time class="sg-event-time" [attr.datetime]="ev.timestamp">{{ ev.timestamp | slice:11:19 }}</time>
                <span class="sg-event-kind">{{ ev.resourceKind }}</span>
                <span class="sg-event-name">{{ ev.resourceName }}</span>
                <span class="sg-event-msg sg-truncate">{{ ev.message }}</span>
              </div>
            }
          }
          <a routerLink="/events" class="sg-card-link">View all events →</a>
        </section>
      </div>

      @if (loadError()) {
        <sg-error-state [error]="loadError()" (retry)="load()" />
      }
    </div>
  `,
  styles: [`
    .sg-context-chip {
      font-size: 0.75rem; color: var(--sg-text-muted);
      background: var(--sg-bg-elevated);
      border: 1px solid var(--sg-border);
      border-radius: 999px;
      padding: 3px 10px;
    }
    .sg-summary-grid {
      display: grid;
      grid-template-columns: repeat(auto-fill, minmax(130px, 1fr));
      gap: 12px;
    }
    .sg-summary-card {
      display: flex; flex-direction: column; gap: 4px;
      text-decoration: none;
      transition: transform var(--sg-transition-fast), border-color var(--sg-transition-fast);
    }
    .sg-summary-card:hover { transform: translateY(-2px); }
    .sg-summary-count { font-size: 1.75rem; font-weight: 700; color: var(--sg-accent); font-family: var(--sg-font-mono); }
    .sg-summary-label { font-size: 0.75rem; color: var(--sg-text-secondary); font-weight: 500; }

    .sg-dashboard-grid {
      display: grid;
      grid-template-columns: repeat(2, 1fr);
      grid-template-rows: auto;
      gap: 16px;
    }
    @media (max-width: 900px) { .sg-dashboard-grid { grid-template-columns: 1fr; } }

    .sg-dashboard-topology { grid-column: 1; }
    .sg-dashboard-traffic  { grid-column: 2; }
    .sg-dashboard-flows    { grid-column: 1; }
    .sg-dashboard-gateways { grid-column: 2; }
    .sg-dashboard-events   { grid-column: 1 / -1; }

    .sg-section-title { font-size: 0.8125rem; font-weight: 600; color: var(--sg-text-secondary); margin-bottom: 12px; text-transform: uppercase; letter-spacing: 0.05em; }
    .sg-mini-topology-canvas { height: 260px; }
    .sg-card-link { display: block; margin-top: 12px; font-size: 0.75rem; color: var(--sg-accent); text-decoration: none; }
    .sg-card-link:hover { text-decoration: underline; }
    .sg-gw-row { display: flex; align-items: center; justify-content: space-between; padding: 6px 0; border-bottom: 1px solid var(--sg-border); }
    .sg-gw-name { font-size: 0.8125rem; color: var(--sg-text-primary); text-decoration: none; }
    .sg-gw-name:hover { color: var(--sg-accent); }
    .sg-event-row { display: grid; grid-template-columns: 60px 90px 1fr 2fr; gap: 8px; padding: 5px 0; border-bottom: 1px solid var(--sg-border); font-size: 0.75rem; align-items: center; }
    .sg-event-time { color: var(--sg-text-muted); font-family: var(--sg-font-mono); }
    .sg-event-kind { color: var(--sg-accent); font-weight: 500; }
    .sg-event-name { color: var(--sg-text-secondary); }
    .sg-event-msg  { color: var(--sg-text-muted); }
    .sg-empty-inline { color: var(--sg-text-muted); font-size: 0.8125rem; padding: 1rem 0; }
  `],
})
export class DashboardComponent implements OnInit {
  readonly context    = inject(ContextService);
  readonly connection = inject(ConnectionService);
  private readonly runtimeConfig  = inject(RuntimeConfigService);
  private readonly gatewayApi     = inject(GatewayApiService);
  private readonly nodeApi        = inject(NodeApiService);
  private readonly tunnelApi      = inject(TunnelApiService);
  private readonly flowApi        = inject(FlowApiService);
  private readonly eventApi       = inject(EventApiService);
  private readonly topologyApi    = inject(TopologyApiService);
  private readonly prometheusApi  = inject(PrometheusApiService);
  private readonly destroyRef     = inject(DestroyRef);

  readonly loading        = signal(false);
  readonly topologyLoading = signal(false);
  readonly trafficLoading = signal(false);
  readonly loadError      = signal<ApiError | null>(null);

  readonly gateways  = signal<Gateway[]>([]);
  readonly flows     = signal<Flow[]>([]);
  readonly events    = signal<StraitEvent[]>([]);
  readonly topology  = signal<TopologyGraph | null>(null);
  readonly trafficSeries = signal<MetricQueryResult['series']>([]);

  readonly summaryCards = computed(() => [
    { label: 'Gateways', count: this.gateways().length,    route: '/gateways' },
    { label: 'Flows',    count: this.flows().length,       route: '/flows' },
    { label: 'Events',   count: this.events().length,      route: '/events' },
  ]);

  ngOnInit(): void { this.load(); }

  load(): void {
    const ns  = this.context.selectedNamespace();
    const cl  = this.context.selectedCluster() ?? undefined;
    this.loading.set(true);
    this.loadError.set(null);

    forkJoin({
      gateways: this.gatewayApi.list({ namespace: ns, cluster: cl, pageSize: 20 }),
      flows:    this.flowApi.list({ namespace: ns, cluster: cl, state: 'Active', pageSize: 10 }),
      events:   this.eventApi.list({ namespace: ns, cluster: cl, pageSize: 10 }),
    }).pipe(takeUntilDestroyed(this.destroyRef)).subscribe({
      next: (res) => {
        this.gateways.set(res.gateways.items);
        this.flows.set(res.flows.items);
        this.events.set(res.events.items);
        this.loading.set(false);
        this.loadTopology(ns, cl);
        this.loadTraffic();
      },
      error: (err) => { this.loadError.set(err); this.loading.set(false); },
    });
  }

  private loadTopology(ns: string, cl?: string): void {
    if (!cl) return;
    this.topologyLoading.set(true);
    this.topologyApi.getGraph(cl, ns)
      .pipe(takeUntilDestroyed(this.destroyRef))
      .subscribe({
        next: (g)  => { this.topology.set(g); this.topologyLoading.set(false); },
        error: ()  => this.topologyLoading.set(false),
      });
  }

  private loadTraffic(): void {
    this.trafficLoading.set(true);
    const now   = new Date();
    const start = new Date(now.getTime() - 30 * 60 * 1000).toISOString();
    const end   = now.toISOString();
    this.prometheusApi.queryRange({
      query: 'sum(rate(straitgateway_traffic_bytes_total[1m]))',
      start, end, step: '60s',
    }).pipe(takeUntilDestroyed(this.destroyRef)).subscribe({
      next: (r)  => { this.trafficSeries.set(r.series); this.trafficLoading.set(false); },
      error: ()  => this.trafficLoading.set(false),
    });
  }
}
