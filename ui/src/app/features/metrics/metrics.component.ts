// Copyright 2026 straitgateway Authors — SPDX-License-Identifier: Apache-2.0
import { Component, inject, signal, OnInit } from '@angular/core';
import { CommonModule } from '@angular/common';
import { ApiClient } from '../../core/api/api-client';
import { catchError, of } from 'rxjs';

interface MetricRow { name: string; value: string; labels: string; help: string; }
interface MetricGroup { prefix: string; metrics: MetricRow[]; }

@Component({
  selector: 'app-metrics', standalone: true, imports: [CommonModule],
  template: `
<div class="sg-fade-in">
  <div class="sg-page-header">
    <div><h1 class="sg-page-title">Metrics</h1><p class="sg-page-subtitle">Prometheus metrics exposed at :9090/metrics by straitgatewayd</p></div>
    <a href="/metrics" target="_blank" class="sg-btn sg-btn-secondary">Open /metrics ↗</a>
  </div>
  @for (group of groups(); track group.prefix) {
    <div class="sg-card">
      <div class="sg-card-header"><span class="sg-card-title">{{ group.prefix }}</span><span class="text-muted text-sm">{{ group.metrics.length }} series</span></div>
      <div class="sg-card-body p-0">
        <table class="sg-table">
          <thead><tr><th>Metric</th><th>Value</th><th>Labels</th><th>Help</th></tr></thead>
          <tbody>
            @for (m of group.metrics; track m.name) {
              <tr>
                <td class="mono text-sm" style="color:var(--sg-accent-light)">{{ m.name }}</td>
                <td class="mono text-sm">{{ m.value }}</td>
                <td class="mono text-sm text-muted">{{ m.labels }}</td>
                <td class="text-sm text-muted">{{ m.help }}</td>
              </tr>
            }
          </tbody>
        </table>
      </div>
    </div>
  } @empty {
    @if (loading()) {
      @for(i of [1,2,3]; track i){<div class="sg-card"><div class="sg-card-body"><div class="sg-skeleton" style="height:14px"></div></div></div>}
    } @else {
      <div class="sg-empty"><p>No metrics available</p><p class="text-muted text-sm">Ensure straitgatewayd is running and exposing :9090/metrics.</p></div>
    }
  }
</div>`,
})
export class MetricsComponent implements OnInit {
  private api = inject(ApiClient);
  groups = signal<MetricGroup[]>([]);
  loading = signal(true);
  ngOnInit() {
    this.api.getMetrics?.().pipe(catchError(() => of([]))).subscribe((d: MetricRow[]) => {
      const m = new Map<string, MetricRow[]>();
      for (const row of d) {
        const prefix = row.name.split('_').slice(0,2).join('_');
        if (!m.has(prefix)) m.set(prefix, []);
        m.get(prefix)!.push(row);
      }
      this.groups.set(Array.from(m.entries()).map(([prefix, metrics]) => ({ prefix, metrics })));
      this.loading.set(false);
    });
    setTimeout(() => { if (this.loading()) this.loading.set(false); }, 800);
  }
}
