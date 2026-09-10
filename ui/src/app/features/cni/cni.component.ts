// Copyright 2026 straitgateway Authors — SPDX-License-Identifier: Apache-2.0
import { Component, inject, signal, OnInit } from '@angular/core';
import { CommonModule } from '@angular/common';
import { ApiClient } from '../../core/api/api-client';
import { catchError, of } from 'rxjs';

interface CniNode { node: string; podCIDR: string; allocated: number; available: number; tunnelEndpoint: string; mtu: number; ready: boolean; }

@Component({
  selector: 'app-cni', standalone: true, imports: [CommonModule],
  template: `
<div class="sg-fade-in">
  <div class="sg-page-header">
    <div><h1 class="sg-page-title">CNI</h1><p class="sg-page-subtitle">NetKit-based pod networking — per-node IPAM and tunnel endpoints</p></div>
  </div>
  <div class="sg-stat-grid">
    <div class="sg-stat-card"><div class="sg-stat-label">Nodes</div><div class="sg-stat-value">{{ nodes().length }}</div></div>
    <div class="sg-stat-card"><div class="sg-stat-label">IPs Allocated</div><div class="sg-stat-value text-accent">{{ totalAllocated() }}</div></div>
    <div class="sg-stat-card"><div class="sg-stat-label">IPs Available</div><div class="sg-stat-value text-success">{{ totalAvail() }}</div></div>
    <div class="sg-stat-card"><div class="sg-stat-label">Ready Nodes</div><div class="sg-stat-value text-success">{{ readyCount() }}</div></div>
  </div>
  <div class="sg-card">
    <div class="sg-card-header"><span class="sg-card-title">Node CNI Status</span></div>
    <div class="sg-card-body p-0">
      <table class="sg-table">
        <thead><tr><th>Node</th><th>Pod CIDR</th><th>Allocated</th><th>Available</th><th>Tunnel Endpoint</th><th>MTU</th><th>CNI</th></tr></thead>
        <tbody>
          @if (loading()) {
            @for(i of [1,2,3]; track i){<tr><td colspan="7"><div class="sg-skeleton" style="height:13px"></div></td></tr>}
          } @else {
            @for (n of nodes(); track n.node) {
              <tr>
                <td class="mono text-sm">{{ n.node }}</td>
                <td class="mono text-sm">{{ n.podCIDR }}</td>
                <td class="mono text-sm" style="color:var(--sg-accent)">{{ n.allocated }}</td>
                <td class="mono text-sm" style="color:var(--sg-green)">{{ n.available }}</td>
                <td class="mono text-sm">{{ n.tunnelEndpoint }}</td>
                <td class="mono text-sm">{{ n.mtu }}</td>
                <td><span class="sg-badge" [class]="n.ready?'active':'error'">{{ n.ready?'Ready':'Error' }}</span></td>
              </tr>
            } @empty {
              <tr><td colspan="7"><div class="sg-empty"><p>No CNI node data</p></div></td></tr>
            }
          }
        </tbody>
      </table>
    </div>
  </div>
</div>`,
})
export class CniComponent implements OnInit {
  private api = inject(ApiClient);
  nodes = signal<CniNode[]>([]);
  loading = signal(true);
  totalAllocated() { return this.nodes().reduce((s, n) => s + n.allocated, 0); }
  totalAvail()     { return this.nodes().reduce((s, n) => s + n.available, 0); }
  readyCount()     { return this.nodes().filter(n => n.ready).length; }
  ngOnInit() {
    this.api.getCniStatus?.().pipe(catchError(() => of([]))).subscribe((d: any[]) => { this.nodes.set(d); this.loading.set(false); });
    setTimeout(() => { if (this.loading()) this.loading.set(false); }, 800);
  }
}
