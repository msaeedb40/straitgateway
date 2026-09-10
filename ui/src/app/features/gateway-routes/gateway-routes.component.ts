// Copyright 2026 straitgateway Authors — SPDX-License-Identifier: Apache-2.0
import { Component, inject, signal, OnInit } from '@angular/core';
import { CommonModule } from '@angular/common';
import { FormsModule } from '@angular/forms';
import { NamespaceService } from '../../core/services/namespace.service';
import { ApiClient } from '../../core/api/api-client';
import { catchError, of } from 'rxjs';

interface RouteRow { name: string; namespace: string; kind: string; parent: string; hostnames: string; rules: number; status: string; }

@Component({
  selector: 'app-gateway-routes', standalone: true, imports: [CommonModule, FormsModule],
  template: `
<div class="sg-fade-in">
  <div class="sg-page-header">
    <div><h1 class="sg-page-title">Routes</h1><p class="sg-page-subtitle">HTTPRoute · GRPCRoute · TCPRoute · TLSRoute · UDPRoute</p></div>
  </div>
  <div style="display:flex;gap:10px;margin-bottom:20px;flex-wrap:wrap">
    @for (kind of kinds; track kind) {
      <button class="sg-btn" [class.sg-btn-secondary]="kindFilter()!==kind" (click)="kindFilter.set(kind)">{{ kind }}</button>
    }
  </div>
  <div class="sg-card">
    <div class="sg-card-header">
      <span class="sg-card-title">{{ kindFilter() === 'All' ? 'All Routes' : kindFilter() }}</span>
      <span class="text-muted text-sm">{{ filtered().length }} routes</span>
    </div>
    <div class="sg-card-body p-0">
      <table class="sg-table">
        <thead><tr><th>Name</th><th>Namespace</th><th>Kind</th><th>Parent Gateway</th><th>Hostnames</th><th>Rules</th><th>Status</th></tr></thead>
        <tbody>
          @if (loading()) {
            @for(i of [1,2,3]; track i){<tr><td colspan="7"><div class="sg-skeleton" style="height:13px"></div></td></tr>}
          } @else {
            @for (r of filtered(); track r.name+r.namespace) {
              <tr>
                <td class="mono text-sm">{{ r.name }}</td>
                <td><span class="sg-badge pending">{{ r.namespace }}</span></td>
                <td><span class="sg-badge info">{{ r.kind }}</span></td>
                <td class="mono text-sm">{{ r.parent }}</td>
                <td class="text-sm">{{ r.hostnames || '—' }}</td>
                <td>{{ r.rules }}</td>
                <td><span class="sg-badge" [class]="r.status==='Accepted'?'active':'warn'">{{ r.status }}</span></td>
              </tr>
            } @empty {
              <tr><td colspan="7"><div class="sg-empty"><p>No routes found</p></div></td></tr>
            }
          }
        </tbody>
      </table>
    </div>
  </div>
</div>`,
})
export class GatewayRoutesComponent implements OnInit {
  private api = inject(ApiClient);
  protected ns = inject(NamespaceService);
  routes = signal<RouteRow[]>([]);
  loading = signal(true);
  kindFilter = signal('All');
  kinds = ['All', 'HTTPRoute', 'GRPCRoute', 'TCPRoute', 'TLSRoute', 'UDPRoute'];

  filtered() {
    return this.routes().filter(r => {
      const nsOk = !this.ns.active() || r.namespace === this.ns.active();
      const kindOk = this.kindFilter() === 'All' || r.kind === this.kindFilter();
      return nsOk && kindOk;
    });
  }

  ngOnInit() {
    this.api.getGatewayRoutes?.().pipe(catchError(() => of([]))).subscribe((data: any[]) => {
      this.routes.set(data);
      this.loading.set(false);
    });
    // fallback if API not ready
    setTimeout(() => { if (this.loading()) this.loading.set(false); }, 1000);
  }
}
