// Copyright 2026 straitgateway Authors — SPDX-License-Identifier: Apache-2.0
import { Component, inject, signal, OnInit, computed } from '@angular/core';
import { CommonModule } from '@angular/common';
import { FormsModule } from '@angular/forms';
import { ApiClient } from '../../core/api/api-client';
import { NamespaceService } from '../../core/services/namespace.service';
import { catchError, of } from 'rxjs';

import { CountByPipe } from '../../shared';

interface ServiceRow { name: string; namespace: string; clusterIP: string; type: string; ports: string; backends: number; algorithm: string; ready: boolean; }

@Component({
  selector: 'app-services', standalone: true, imports: [CommonModule, FormsModule, CountByPipe],
  template: `
<div class="sg-fade-in">
  <div class="sg-page-header">
    <div>
      <h1 class="sg-page-title">Services</h1>
      <p class="sg-page-subtitle">Kubernetes Services managed by the straitgateway Maglev LB</p>
    </div>
  </div>
  <div class="sg-stat-grid">
    <div class="sg-stat-card"><div class="sg-stat-label">Total Services</div><div class="sg-stat-value">{{ services().length }}</div></div>
    <div class="sg-stat-card"><div class="sg-stat-label">ClusterIP</div><div class="sg-stat-value">{{ services() | countBy:'type':'ClusterIP' }}</div></div>
    <div class="sg-stat-card"><div class="sg-stat-label">NodePort</div><div class="sg-stat-value">{{ services() | countBy:'type':'NodePort' }}</div></div>
    <div class="sg-stat-card"><div class="sg-stat-label">LoadBalancer</div><div class="sg-stat-value">{{ services() | countBy:'type':'LoadBalancer' }}</div></div>
  </div>
  <div class="sg-card">
    <div class="sg-card-header">
      <span class="sg-card-title">Service Catalog</span>
      <div style="display:flex;gap:12px;align-items:center">
        <input class="sg-search-input" style="width:220px;height:34px;padding:0 12px" placeholder="Filter by name…" [(ngModel)]="filter">
        <select class="sg-select" style="height:34px;padding:0 12px" [(ngModel)]="typeFilter">
          <option value="">All Types</option><option>ClusterIP</option><option>NodePort</option><option>LoadBalancer</option>
        </select>
      </div>
    </div>
    <div class="sg-card-body p-0">
      <table class="sg-table">
        <thead><tr><th>Name</th><th>Namespace</th><th>ClusterIP</th><th>Type</th><th>Ports</th><th>Backends</th><th>Algorithm</th><th>Status</th></tr></thead>
        <tbody>
          @if (loading()) {
            @for(i of [1,2,3,4,5]; track i){<tr><td colspan="8"><div class="sg-skeleton" style="height:14px;margin:2px 0"></div></td></tr>}
          } @else {
            @for (svc of filtered(); track svc.name+svc.namespace) {
              <tr>
                <td><span class="mono text-sm">{{ svc.name }}</span></td>
                <td><span class="sg-badge pending">{{ svc.namespace }}</span></td>
                <td class="mono text-sm">{{ svc.clusterIP }}</td>
                <td><span class="sg-badge" [class]="svc.type==='LoadBalancer'?'active':svc.type==='NodePort'?'info':'pending'">{{ svc.type }}</span></td>
                <td class="mono text-sm">{{ svc.ports }}</td>
                <td>{{ svc.backends }}</td>
                <td><span class="sg-badge info">{{ svc.algorithm }}</span></td>
                <td><span class="sg-badge" [class]="svc.ready?'active':'warn'">{{ svc.ready?'Ready':'Pending' }}</span></td>
              </tr>
            } @empty {
              <tr><td colspan="8"><div class="sg-empty"><p>No services match the current filter</p></div></td></tr>
            }
          }
        </tbody>
      </table>
    </div>
  </div>
</div>`,
})
export class ServicesComponent implements OnInit {
  private api = inject(ApiClient);
  protected ns = inject(NamespaceService);
  services = signal<ServiceRow[]>([]);
  loading = signal(true);
  filter = '';
  typeFilter = '';

  filtered() {
    return this.services().filter(s => {
      const nsOk = !this.ns.active() || s.namespace === this.ns.active();
      const nameOk = !this.filter || s.name.includes(this.filter);
      const typeOk = !this.typeFilter || s.type === this.typeFilter;
      return nsOk && nameOk && typeOk;
    });
  }

  ngOnInit() {
    this.api.getServices().pipe(catchError(() => of([]))).subscribe((data: any[]) => {
      this.services.set(data.map(d => ({
        name: d.name, namespace: d.namespace, clusterIP: d.clusterIP ?? '—',
        type: d.type ?? 'ClusterIP', ports: d.ports ?? '—',
        backends: d.backends ?? 0, algorithm: d.algorithm ?? 'Maglev',
        ready: d.ready ?? true,
      })));
      this.loading.set(false);
    });
  }
}
