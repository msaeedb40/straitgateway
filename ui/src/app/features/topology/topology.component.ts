import { Component, inject, signal, DestroyRef, computed, effect } from '@angular/core';
import { DecimalPipe } from '@angular/common';
import { takeUntilDestroyed } from '@angular/core/rxjs-interop';
import { TopologyApiService } from '../../core/api/resources/topology.api';
import { ContextService } from '../../core/services/context.service';
import { PrometheusApiService } from '../../core/api/observability/prometheus.api';
import { BreadcrumbsComponent } from '../../layout/breadcrumbs/breadcrumbs.component';
import { TopologyGraphComponent } from '../../shared/topology/topology-graph/topology-graph.component';
import { SkeletonComponent } from '../../shared/components/skeleton/skeleton.component';
import { ErrorStateComponent } from '../../shared/components/error-state/error-state.component';
import { StatusBadgeComponent } from '../../shared/components/status-badge/status-badge.component';
import { ApiError } from '../../core/api/api-error';
import { TopologyGraph, TopologyNode, TopologyEdge } from '../../core/models/topology.model';

@Component({
  selector: 'sg-topology',
  imports: [DecimalPipe, BreadcrumbsComponent, TopologyGraphComponent, SkeletonComponent, ErrorStateComponent, StatusBadgeComponent],
  template: `
    <div class="sg-topology-page">
      <div class="sg-topology-toolbar">
        <sg-breadcrumbs />
        <h1 class="sg-page-title">Network Topology</h1>
        <div class="sg-toolbar-actions">
          <button class="sg-btn sg-btn-secondary" (click)="load()" [disabled]="loading()">Refresh</button>
        </div>
      </div>

      @if (loading()) {
        <sg-skeleton variant="block" width="100%" height="calc(100vh - 160px)" />
      } @else if (loadError()) {
        <sg-error-state [error]="loadError()" (retry)="load()" />
      } @else if (graph()) {
        <div class="sg-topology-canvas-wrapper">
          <sg-topology-graph
            [graph]="graph()!"
            (nodeSelected)="onNodeSelected($event)"
            (edgeSelected)="onEdgeSelected($event)"
          />
        </div>

        <!-- Side panel for selected node -->
        @if (selectedNode()) {
          <aside class="sg-topology-detail-panel sg-glass" aria-label="Selected node details" role="complementary">
            <div class="sg-panel-header">
              <h2 class="sg-panel-title">{{ selectedNode()!.kind }}: {{ selectedNode()!.name }}</h2>
              <button class="sg-btn sg-btn-ghost" (click)="selectedNode.set(null)" aria-label="Close panel">✕</button>
            </div>
            <dl class="sg-panel-dl">
              <dt>Namespace</dt><dd>{{ selectedNode()!.namespace }}</dd>
              <dt>Cluster</dt><dd>{{ selectedNode()!.cluster }}</dd>
              <dt>Health</dt><dd><sg-status-badge [health]="selectedNode()!.health" /></dd>
            </dl>
          </aside>
        }

        <!-- Side panel for selected edge -->
        @if (selectedEdge()) {
          <aside class="sg-topology-detail-panel sg-glass" aria-label="Selected connection details" role="complementary">
            <div class="sg-panel-header">
              <h2 class="sg-panel-title">Connection</h2>
              <button class="sg-btn sg-btn-ghost" (click)="selectedEdge.set(null)" aria-label="Close panel">✕</button>
            </div>
            <dl class="sg-panel-dl">
              <dt>Type</dt><dd>{{ selectedEdge()!.kind }}</dd>
              <dt>Traffic</dt>
              <dd class="sg-mono-value">
                {{ selectedEdge()!.trafficBytesPerSec !== null ? (selectedEdge()!.trafficBytesPerSec! | number:'1.0-0') + ' B/s' : 'Metrics unavailable' }}
              </dd>
              <dt>Latency</dt>
              <dd class="sg-mono-value">{{ selectedEdge()!.latencyMs !== null ? selectedEdge()!.latencyMs + ' ms' : '—' }}</dd>
              <dt>Error rate</dt>
              <dd class="sg-mono-value">{{ selectedEdge()!.errorRate !== null ? (selectedEdge()!.errorRate! * 100 | number:'1.2-2') + '%' : '—' }}</dd>
            </dl>
          </aside>
        }
      }
    </div>
  `,
  styles: [`
    .sg-topology-page { display:flex;flex-direction:column;height:100%;position:relative;overflow:hidden; }
    .sg-topology-toolbar { display:flex;align-items:center;gap:12px;padding:.75rem 1.5rem;border-bottom:1px solid var(--sg-border);flex-shrink:0; }
    .sg-toolbar-actions { margin-left:auto; }
    .sg-topology-canvas-wrapper { flex:1;min-height:0;position:relative; }
    .sg-topology-detail-panel {
      position:absolute;top:12px;right:12px;width:280px;
      border-radius:var(--sg-radius-lg);padding:1rem;
      display:flex;flex-direction:column;gap:.75rem;
    }
    .sg-panel-header { display:flex;align-items:center;justify-content:space-between; }
    .sg-panel-title { font-size:.875rem;font-weight:600;color:var(--sg-text-primary); }
    .sg-panel-dl { display:grid;grid-template-columns:auto 1fr;gap:6px 12px;font-size:.8125rem; }
    .sg-panel-dl dt { color:var(--sg-text-muted); }
    .sg-panel-dl dd { color:var(--sg-text-primary); }
  `],
})
export class TopologyComponent {
  private readonly topologyApi  = inject(TopologyApiService);
  readonly context              = inject(ContextService);
  private readonly destroyRef   = inject(DestroyRef);

  readonly loading     = signal(false);
  readonly loadError   = signal<ApiError | null>(null);
  readonly graph       = signal<TopologyGraph | null>(null);
  readonly selectedNode = signal<TopologyNode | null>(null);
  readonly selectedEdge = signal<TopologyEdge | null>(null);

  constructor() {
    effect(() => {
      const cl = this.context.selectedCluster();
      const ns = this.context.selectedNamespace();
      if (cl && ns) this.load();
    });
  }

  load(): void {
    const cl = this.context.selectedCluster();
    const ns = this.context.selectedNamespace();
    if (!cl || !ns) return;
    this.loading.set(true);
    this.loadError.set(null);
    this.topologyApi.getGraph(cl, ns)
      .pipe(takeUntilDestroyed(this.destroyRef))
      .subscribe({
        next:  (g)   => { this.graph.set(g); this.loading.set(false); },
        error: (err) => { this.loadError.set(err); this.loading.set(false); },
      });
  }

  onNodeSelected(node: TopologyNode): void { this.selectedEdge.set(null); this.selectedNode.set(node); }
  onEdgeSelected(edge: TopologyEdge): void { this.selectedNode.set(null); this.selectedEdge.set(edge); }
}
