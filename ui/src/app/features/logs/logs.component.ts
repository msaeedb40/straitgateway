// Copyright 2026 straitgateway Authors — SPDX-License-Identifier: Apache-2.0
import { Component, inject, signal, OnInit, ViewChild, ElementRef } from '@angular/core';
import { CommonModule } from '@angular/common';
import { FormsModule } from '@angular/forms';
import { ApiClient } from '../../core/api/api-client';
import { catchError, of } from 'rxjs';

interface LogLine { timestamp: string; level: 'debug'|'info'|'warn'|'error'; caller: string; msg: string; fields: string; }

@Component({
  selector: 'app-logs', standalone: true, imports: [CommonModule, FormsModule],
  template: `
<div class="sg-fade-in">
  <div class="sg-page-header">
    <div><h1 class="sg-page-title">Logs</h1><p class="sg-page-subtitle">Structured JSON logs from straitgatewayd (via zap)</p></div>
    <div style="display:flex;gap:12px;align-items:center">
      <select class="sg-select" style="height:34px;padding:0 12px" [(ngModel)]="levelFilter">
        <option value="">All Levels</option><option>debug</option><option>info</option><option>warn</option><option>error</option>
      </select>
      <input class="sg-search-input" style="width:200px;height:34px;padding:0 12px" placeholder="Filter message…" [(ngModel)]="msgFilter">
      <button class="sg-btn sg-btn-secondary" (click)="load()">↻ Refresh</button>
    </div>
  </div>
  <div class="sg-card">
    <div class="sg-card-body p-0">
      <div #logContainer class="sg-log-container">
        @if (loading()) {
          @for(i of [1,2,3,4,5,6]; track i){<div class="sg-skeleton" style="height:12px;margin:4px 0"></div>}
        } @else {
          @for (line of filtered(); track line.timestamp+line.msg) {
            <div class="sg-log-line" [class]="'level-'+line.level">
              <span class="sg-log-ts">{{ line.timestamp }}</span>
              <span class="sg-log-level" [class]="'level-'+line.level">{{ line.level.toUpperCase() }}</span>
              <span class="sg-log-caller">{{ line.caller }}</span>
              <span class="sg-log-msg">{{ line.msg }}</span>
              @if (line.fields) { <span class="sg-log-fields">{{ line.fields }}</span> }
            </div>
          } @empty {
            <div class="sg-empty"><p>No log lines match the current filter</p></div>
          }
        }
      </div>
    </div>
  </div>
</div>`,
  styles: [`
    .sg-log-container { font-family:var(--sg-font-mono,monospace);font-size:11px;line-height:1.7;padding:12px;max-height:70vh;overflow-y:auto;background:var(--sg-surface-1); }
    .sg-log-line { display:flex;gap:10px;padding:1px 0; }
    .sg-log-line.level-error { background:rgba(239,68,68,.05); }
    .sg-log-line.level-warn  { background:rgba(245,158,11,.05); }
    .sg-log-ts { color:var(--sg-text-3);flex-shrink:0;width:170px; }
    .sg-log-level { flex-shrink:0;width:44px;font-weight:600; }
    .sg-log-level.level-error { color:var(--sg-danger); }
    .sg-log-level.level-warn  { color:#f59e0b; }
    .sg-log-level.level-info  { color:var(--sg-accent); }
    .sg-log-level.level-debug { color:var(--sg-text-3); }
    .sg-log-caller { color:var(--sg-text-3);flex-shrink:0;width:220px;overflow:hidden;text-overflow:ellipsis; }
    .sg-log-msg { color:var(--sg-text-1);flex:1; }
    .sg-log-fields { color:var(--sg-text-3); }
  `],
})
export class LogsComponent implements OnInit {
  private api = inject(ApiClient);
  @ViewChild('logContainer') logContainer?: ElementRef;
  lines = signal<LogLine[]>([]);
  loading = signal(true);
  levelFilter = '';
  msgFilter = '';
  filtered() { return this.lines().filter(l => (!this.levelFilter || l.level === this.levelFilter) && (!this.msgFilter || l.msg.includes(this.msgFilter))); }
  ngOnInit() { this.load(); }
  load() {
    this.loading.set(true);
    this.api.getLogs?.().pipe(catchError(() => of([]))).subscribe((d: any[]) => { this.lines.set(d); this.loading.set(false); });
    setTimeout(() => { if (this.loading()) this.loading.set(false); }, 800);
  }
}
