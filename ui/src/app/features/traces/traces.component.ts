// Copyright 2026 straitgateway Authors — SPDX-License-Identifier: Apache-2.0
import { Component, inject, signal, OnInit } from '@angular/core';
import { CommonModule } from '@angular/common';
import { FormsModule } from '@angular/forms';
import { ApiClient } from '../../core/api/api-client';
import { catchError, of } from 'rxjs';

interface TraceRow { traceId: string; spanId: string; operation: string; service: string; duration: string; status: 'ok'|'error'; timestamp: string; }

@Component({
  selector: 'app-traces', standalone: true, imports: [CommonModule, FormsModule],
  template: `
<div class="sg-fade-in">
  <div class="sg-page-header">
    <div><h1 class="sg-page-title">Traces</h1><p class="sg-page-subtitle">OpenTelemetry distributed traces exported via OTLP</p></div>
    <input class="sg-search-input" style="width:240px" placeholder="Filter by operation or service…" [(ngModel)]="filter">
  </div>
  <div class="sg-card">
    <div class="sg-card-header"><span class="sg-card-title">Recent Spans</span><span class="text-muted text-sm">{{ filtered().length }} spans</span></div>
    <div class="sg-card-body p-0">
      <table class="sg-table">
        <thead><tr><th>Trace ID</th><th>Operation</th><th>Service</th><th>Duration</th><th>Status</th><th>Time</th></tr></thead>
        <tbody>
          @if (loading()) {
            @for(i of [1,2,3,4]; track i){<tr><td colspan="6"><div class="sg-skeleton" style="height:13px"></div></td></tr>}
          } @else {
            @for (t of filtered(); track t.traceId+t.spanId) {
              <tr>
                <td class="mono text-sm" style="color:var(--sg-accent)">{{ t.traceId.substring(0,12) }}…</td>
                <td class="mono text-sm">{{ t.operation }}</td>
                <td><span class="sg-badge info">{{ t.service }}</span></td>
                <td class="mono text-sm">{{ t.duration }}</td>
                <td><span class="sg-badge" [class]="t.status==='ok'?'active':'error'">{{ t.status }}</span></td>
                <td class="mono text-sm text-muted">{{ t.timestamp }}</td>
              </tr>
            } @empty {
              <tr><td colspan="6"><div class="sg-empty"><p>No traces available</p><p class="text-muted text-sm">Configure OTLP endpoint in Settings → Observability to export traces.</p></div></td></tr>
            }
          }
        </tbody>
      </table>
    </div>
  </div>
</div>`,
})
export class TracesComponent implements OnInit {
  private api = inject(ApiClient);
  traces = signal<TraceRow[]>([]);
  loading = signal(true);
  filter = '';
  filtered() { return this.traces().filter(t => !this.filter || t.operation.includes(this.filter) || t.service.includes(this.filter)); }
  ngOnInit() {
    this.api.getTraces?.().pipe(catchError(() => of([]))).subscribe((d: any[]) => { this.traces.set(d); this.loading.set(false); });
    setTimeout(() => { if (this.loading()) this.loading.set(false); }, 800);
  }
}
