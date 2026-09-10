// Copyright 2026 straitgateway Authors
// SPDX-License-Identifier: Apache-2.0

import { Component, inject, signal, OnInit } from '@angular/core';
import { CommonModule } from '@angular/common';
import { ApiClient, FlowEvent } from '../../core/api/api-client';
import { catchError, of, interval } from 'rxjs';
import { takeUntilDestroyed } from '@angular/core/rxjs-interop';

import { FormsModule } from '@angular/forms';

@Component({
  selector: 'app-flows',
  standalone: true,
  imports: [CommonModule, FormsModule],
  template: `
<div class="sg-fade-in">
  <div class="sg-page-header">
    <div>
      <h1 class="sg-page-title">Flow Logs</h1>
      <p class="sg-page-subtitle">eBPF ring buffer network flow events — observability only, not forwarding path</p>
    </div>
    <div class="flex gap-2">
      <select class="sg-select" [(ngModel)]="filter" style="width:120px">
        <option value="">All</option>
        <option value="deny">Denied</option>
        <option value="allow">Allowed</option>
      </select>
      <button class="sg-btn sg-btn-secondary" (click)="load()">Refresh</button>
    </div>
  </div>

  <div class="sg-card">
    <div class="sg-card-header">
      <span class="sg-card-title">Recent Flows</span>
      <span class="text-muted text-sm">Last 100 events from ring buffer</span>
    </div>
    <div class="sg-table-wrap">
      <table class="sg-table">
        <thead>
          <tr>
            <th>Time</th>
            <th>Direction</th>
            <th>Src IP:Port</th>
            <th>Dst IP:Port</th>
            <th>Proto</th>
            <th>Src Identity</th>
            <th>Dst Identity</th>
            <th>Bytes</th>
            <th>Action</th>
            <th>Drop Reason</th>
          </tr>
        </thead>
        <tbody>
          @for (flow of flows(); track flow.timestampNs) {
            <tr>
              <td class="mono text-muted text-sm">{{ flow.timestampNs | date:'HH:mm:ss.SSS' }}</td>
              <td><span class="sg-badge" [class]="flow.direction === 'ingress' ? 'active' : 'pending'">{{ flow.direction }}</span></td>
              <td class="mono">{{ flow.srcIP }}:{{ flow.srcPort }}</td>
              <td class="mono">{{ flow.dstIP }}:{{ flow.dstPort }}</td>
              <td class="mono">{{ flow.protocol }}</td>
              <td class="mono">{{ flow.srcIdentity }}</td>
              <td class="mono">{{ flow.dstIdentity }}</td>
              <td class="mono">{{ flow.bytes | number }}</td>
              <td>
                <span class="sg-badge"
                  [class]="flow.action === 'allow' ? 'active' : flow.action === 'deny' ? 'error' : 'warn'">
                  {{ flow.action }}
                </span>
              </td>
              <td class="mono text-muted text-sm">{{ flow.dropReason ?? '—' }}</td>
            </tr>
          } @empty {
            <tr><td colspan="10">
              <div class="sg-empty">
                <p>No flow events. Ensure straitgatewayd is running and flow logs are enabled.</p>
              </div>
            </td></tr>
          }
        </tbody>
      </table>
    </div>
  </div>
</div>
`,
})
export class FlowsComponent implements OnInit {
  private api = inject(ApiClient);
  flows  = signal<FlowEvent[]>([]);
  filter = '';

  ngOnInit() { this.load(); }
  load() { this.api.getFlows(100).pipe(catchError(() => of([]))).subscribe(f => this.flows.set(f)); }
}
