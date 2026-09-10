// Copyright 2026 straitgateway Authors — SPDX-License-Identifier: Apache-2.0
import { Component, inject, signal, OnInit } from '@angular/core';
import { CommonModule } from '@angular/common';
import { ApiClient } from '../../core/api/api-client';
import { catchError, of } from 'rxjs';

interface MapStat { name: string; type: string; pinPath: string; entries: number; maxEntries: number; fillPct: number; }

@Component({
  selector: 'app-ebpf', standalone: true, imports: [CommonModule],
  template: `
<div class="sg-fade-in">
  <div class="sg-page-header">
    <div><h1 class="sg-page-title">eBPF Maps</h1><p class="sg-page-subtitle">BPF map statistics from the straitgateway dataplane compiler</p></div>
    <button class="sg-btn sg-btn-secondary" (click)="load()">↻ Refresh</button>
  </div>
  <div class="sg-card">
    <div class="sg-card-header"><span class="sg-card-title">BPF Map Utilization</span></div>
    <div class="sg-card-body p-0">
      <table class="sg-table">
        <thead><tr><th>Map Name</th><th>Type</th><th>Pin Path</th><th>Entries</th><th>Max</th><th style="width:180px">Utilization</th></tr></thead>
        <tbody>
          @if (loading()) {
            @for(i of [1,2,3,4,5,6,7]; track i){<tr><td colspan="6"><div class="sg-skeleton" style="height:13px"></div></td></tr>}
          } @else {
            @for (m of maps(); track m.name) {
              <tr>
                <td class="mono text-sm" style="color:var(--sg-accent)">{{ m.name }}</td>
                <td><span class="sg-badge info">{{ m.type }}</span></td>
                <td class="mono text-sm text-muted">{{ m.pinPath }}</td>
                <td class="mono text-sm">{{ m.entries | number }}</td>
                <td class="mono text-sm text-muted">{{ m.maxEntries | number }}</td>
                <td>
                  <div style="display:flex;align-items:center;gap:8px">
                    <div style="flex:1;height:6px;background:var(--sg-surface-3);border-radius:3px;overflow:hidden">
                      <div style="height:100%;border-radius:3px;transition:width .3s"
                           [style.width]="m.fillPct + '%'"
                           [style.background]="m.fillPct > 80 ? 'var(--sg-danger)' : m.fillPct > 60 ? '#f59e0b' : 'var(--sg-green)'">
                      </div>
                    </div>
                    <span class="mono text-sm" style="width:36px;text-align:right">{{ m.fillPct }}%</span>
                  </div>
                </td>
              </tr>
            } @empty {
              <tr><td colspan="6"><div class="sg-empty"><p>No BPF map data available</p><p class="text-muted text-sm">Ensure straitgatewayd is running and BPF maps are pinned at /sys/fs/bpf/straitgateway/</p></div></td></tr>
            }
          }
        </tbody>
      </table>
    </div>
  </div>
</div>`,
})
export class EbpfComponent implements OnInit {
  private api = inject(ApiClient);
  maps = signal<MapStat[]>([]);
  loading = signal(true);
  ngOnInit() { this.load(); }
  load() {
    this.loading.set(true);
    this.api.getBpfMapStats?.().pipe(catchError(() => of([]))).subscribe((d: any[]) => {
      this.maps.set(d.map(m => ({ ...m, fillPct: Math.round((m.entries / m.maxEntries) * 100) })));
      this.loading.set(false);
    });
    setTimeout(() => { if (this.loading()) this.loading.set(false); }, 800);
  }
}
