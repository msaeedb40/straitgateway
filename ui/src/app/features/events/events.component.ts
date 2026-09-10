// Copyright 2026 straitgateway Authors — SPDX-License-Identifier: Apache-2.0
import { Component, inject, signal, OnInit } from '@angular/core';
import { CommonModule } from '@angular/common';
import { FormsModule } from '@angular/forms';
import { NamespaceService } from '../../core/services/namespace.service';
import { ApiClient } from '../../core/api/api-client';
import { catchError, of } from 'rxjs';

interface EventRow { timestamp: string; namespace: string; name: string; kind: string; reason: string; message: string; type: 'Normal'|'Warning'; count: number; }

@Component({
  selector: 'app-events', standalone: true, imports: [CommonModule, FormsModule],
  template: `
<div class="sg-fade-in">
  <div class="sg-page-header">
    <div><h1 class="sg-page-title">Events</h1><p class="sg-page-subtitle">Kubernetes events from straitgateway controllers and the dataplane</p></div>
    <div style="display:flex;gap:12px;align-items:center">
      <select class="sg-select" style="height:34px;padding:0 12px" [(ngModel)]="typeFilter"><option value="">All Types</option><option>Normal</option><option>Warning</option></select>
      <button class="sg-btn sg-btn-secondary" (click)="load()">↻ Refresh</button>
    </div>
  </div>
  <div class="sg-card">
    <div class="sg-card-header">
      <span class="sg-card-title">Event Stream</span>
      <span class="text-muted text-sm">{{ filtered().length }} events</span>
    </div>
    <div class="sg-card-body p-0">
      <table class="sg-table">
        <thead><tr><th style="width:160px">Time</th><th>Namespace</th><th>Object</th><th>Reason</th><th>Message</th><th>Count</th></tr></thead>
        <tbody>
          @if (loading()) {
            @for(i of [1,2,3,4,5]; track i){<tr><td colspan="6"><div class="sg-skeleton" style="height:13px"></div></td></tr>}
          } @else {
            @for (e of filtered(); track e.timestamp+e.name) {
              <tr [style.border-left]="e.type==='Warning'?'2px solid var(--sg-danger)':''">
                <td class="mono text-sm text-muted">{{ e.timestamp }}</td>
                <td><span class="sg-badge pending">{{ e.namespace }}</span></td>
                <td class="mono text-sm">{{ e.kind }}/{{ e.name }}</td>
                <td><span class="sg-badge" [class]="e.type==='Warning'?'error':'info'">{{ e.reason }}</span></td>
                <td class="text-sm">{{ e.message }}</td>
                <td>{{ e.count }}</td>
              </tr>
            } @empty {
              <tr><td colspan="6"><div class="sg-empty"><p>No events</p></div></td></tr>
            }
          }
        </tbody>
      </table>
    </div>
  </div>
</div>`,
})
export class EventsComponent implements OnInit {
  private api = inject(ApiClient);
  protected ns = inject(NamespaceService);
  events = signal<EventRow[]>([]);
  loading = signal(true);
  typeFilter = '';
  filtered() { return this.events().filter(e => (!this.ns.active() || e.namespace === this.ns.active()) && (!this.typeFilter || e.type === this.typeFilter)); }
  ngOnInit() { this.load(); }
  load() {
    this.loading.set(true);
    this.api.getEvents?.().pipe(catchError(() => of([]))).subscribe((d: any[]) => { this.events.set(d); this.loading.set(false); });
    setTimeout(() => { if (this.loading()) this.loading.set(false); }, 800);
  }
}
