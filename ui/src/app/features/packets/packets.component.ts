// Copyright 2026 straitgateway Authors — SPDX-License-Identifier: Apache-2.0
import { Component, inject, signal, OnInit } from '@angular/core';
import { CommonModule } from '@angular/common';
import { FormsModule } from '@angular/forms';
import { ApiClient } from '../../core/api/api-client';
import { catchError, of } from 'rxjs';

interface PacketRow { timestamp: string; srcIP: string; dstIP: string; srcPort: number; dstPort: number; proto: string; bytes: number; action: 'ALLOW'|'DROP'|'SNAT'|'DNAT'; reason: string; }

@Component({
  selector: 'app-packets', standalone: true, imports: [CommonModule, FormsModule],
  template: `
<div class="sg-fade-in">
  <div class="sg-page-header">
    <div><h1 class="sg-page-title">Packets</h1><p class="sg-page-subtitle">Packet-level drop reasons and action events from eBPF ring buffer</p></div>
    <div style="display:flex;gap:8px">
      <select class="sg-ns-select" [(ngModel)]="actionFilter">
        <option value="">All Actions</option><option>ALLOW</option><option>DROP</option><option>SNAT</option><option>DNAT</option>
      </select>
      <button class="sg-btn sg-btn-secondary" (click)="load()">↻ Refresh</button>
    </div>
  </div>
  <div class="sg-stat-grid">
    <div class="sg-stat-card"><div class="sg-stat-label">Total</div><div class="sg-stat-value">{{ packets().length }}</div></div>
    <div class="sg-stat-card"><div class="sg-stat-label">Allowed</div><div class="sg-stat-value text-success">{{ count('ALLOW') }}</div></div>
    <div class="sg-stat-card"><div class="sg-stat-label">Dropped</div><div class="sg-stat-value text-danger">{{ count('DROP') }}</div></div>
    <div class="sg-stat-card"><div class="sg-stat-label">NATed</div><div class="sg-stat-value text-accent">{{ count('SNAT') + count('DNAT') }}</div></div>
  </div>
  <div class="sg-card">
    <div class="sg-card-body p-0">
      <table class="sg-table">
        <thead><tr><th>Time</th><th>Src</th><th>Dst</th><th>Proto</th><th>Bytes</th><th>Action</th><th>Reason</th></tr></thead>
        <tbody>
          @if (loading()) {
            @for(i of [1,2,3,4,5]; track i){<tr><td colspan="7"><div class="sg-skeleton" style="height:13px"></div></td></tr>}
          } @else {
            @for (p of filtered(); track p.timestamp+p.srcIP) {
              <tr>
                <td class="mono text-sm text-muted">{{ p.timestamp }}</td>
                <td class="mono text-sm">{{ p.srcIP }}:{{ p.srcPort }}</td>
                <td class="mono text-sm">{{ p.dstIP }}:{{ p.dstPort }}</td>
                <td><span class="sg-badge info">{{ p.proto }}</span></td>
                <td class="mono text-sm">{{ p.bytes | number }}</td>
                <td><span class="sg-badge" [class]="p.action==='ALLOW'?'active':p.action==='DROP'?'error':'info'">{{ p.action }}</span></td>
                <td class="text-sm text-muted">{{ p.reason || '—' }}</td>
              </tr>
            } @empty {
              <tr><td colspan="7"><div class="sg-empty"><p>No packet events</p></div></td></tr>
            }
          }
        </tbody>
      </table>
    </div>
  </div>
</div>`,
})
export class PacketsComponent implements OnInit {
  private api = inject(ApiClient);
  packets = signal<PacketRow[]>([]);
  loading = signal(true);
  actionFilter = '';
  count(action: string) { return this.packets().filter(p => p.action === action).length; }
  filtered() { return this.packets().filter(p => !this.actionFilter || p.action === this.actionFilter); }
  ngOnInit() { this.load(); }
  load() {
    this.loading.set(true);
    this.api.getPackets?.().pipe(catchError(() => of([]))).subscribe((d: any[]) => { this.packets.set(d); this.loading.set(false); });
    setTimeout(() => { if (this.loading()) this.loading.set(false); }, 800);
  }
}
