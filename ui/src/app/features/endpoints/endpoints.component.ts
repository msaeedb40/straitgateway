// Copyright 2026 straitgateway Authors — SPDX-License-Identifier: Apache-2.0
import { Component, inject, signal, OnInit } from '@angular/core';
import { CommonModule } from '@angular/common';
import { NamespaceService } from '../../core/services/namespace.service';
import { ApiClient } from '../../core/api/api-client';
import { catchError, of } from 'rxjs';

interface EndpointRow { service: string; namespace: string; address: string; port: number; protocol: string; identity: number; state: string; }

@Component({
  selector: 'app-endpoints', standalone: true, imports: [CommonModule],
  template: `
<div class="sg-fade-in">
  <div class="sg-page-header">
    <div><h1 class="sg-page-title">Endpoints</h1><p class="sg-page-subtitle">Active backend endpoints tracked by the straitgateway dataplane</p></div>
  </div>
  <div class="sg-card">
    <div class="sg-card-header">
      <span class="sg-card-title">Endpoint Slices</span>
      <span class="text-muted text-sm">{{ filtered().length }} endpoints · ns: {{ ns.active()||'all' }}</span>
    </div>
    <div class="sg-card-body p-0">
      <table class="sg-table">
        <thead><tr><th>Service</th><th>Namespace</th><th>Address</th><th>Port</th><th>Protocol</th><th>Identity</th><th>State</th></tr></thead>
        <tbody>
          @if (loading()) {
            @for(i of [1,2,3,4]; track i){<tr><td colspan="7"><div class="sg-skeleton" style="height:13px"></div></td></tr>}
          } @else {
            @for (ep of filtered(); track ep.address+ep.port) {
              <tr>
                <td class="mono text-sm">{{ ep.service }}</td>
                <td><span class="sg-badge pending">{{ ep.namespace }}</span></td>
                <td class="mono text-sm">{{ ep.address }}</td>
                <td class="mono text-sm">{{ ep.port }}</td>
                <td><span class="sg-badge info">{{ ep.protocol }}</span></td>
                <td class="mono text-sm" style="color:var(--sg-accent)">{{ ep.identity }}</td>
                <td><span class="sg-badge" [class]="ep.state==='Ready'?'active':'warn'">{{ ep.state }}</span></td>
              </tr>
            } @empty {
              <tr><td colspan="7"><div class="sg-empty"><p>No endpoints found</p></div></td></tr>
            }
          }
        </tbody>
      </table>
    </div>
  </div>
</div>`,
})
export class EndpointsComponent implements OnInit {
  private api = inject(ApiClient);
  protected ns = inject(NamespaceService);
  endpoints = signal<EndpointRow[]>([]);
  loading = signal(true);
  filtered() { const n = this.ns.active(); return n ? this.endpoints().filter(e => e.namespace === n) : this.endpoints(); }
  ngOnInit() {
    this.api.getEndpoints?.().pipe(catchError(() => of([]))).subscribe((d: any[]) => { this.endpoints.set(d); this.loading.set(false); });
    setTimeout(() => { if (this.loading()) this.loading.set(false); }, 800);
  }
}
