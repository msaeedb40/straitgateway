import { Component, inject, signal, DestroyRef, PLATFORM_ID } from '@angular/core';
import { isPlatformBrowser, SlicePipe } from '@angular/common';
import { takeUntilDestroyed } from '@angular/core/rxjs-interop';
import { ReactiveFormsModule, FormBuilder } from '@angular/forms';
import { LogsApiService } from '../../core/api/observability/logs.api';
import { BreadcrumbsComponent } from '../../layout/breadcrumbs/breadcrumbs.component';
import { SkeletonListComponent } from '../../shared/components/skeleton/skeleton.component';
import { ErrorStateComponent } from '../../shared/components/error-state/error-state.component';
import { ContextService } from '../../core/services/context.service';
import { LogEntry, LogSeverity } from '../../core/models/log.model';
import { ApiError } from '../../core/api/api-error';

const SEVERITIES: LogSeverity[] = ['DEBUG', 'INFO', 'WARN', 'ERROR', 'FATAL'];

@Component({
  selector: 'sg-logs',
  imports: [ReactiveFormsModule, SlicePipe, BreadcrumbsComponent, SkeletonListComponent, ErrorStateComponent],
  template: `
    <div class="sg-logs-page">
      <div class="sg-logs-toolbar">
        <sg-breadcrumbs />
        <h1 class="sg-page-title">Logs</h1>
        <form [formGroup]="filterForm" (ngSubmit)="query()" class="sg-filter-bar" role="search" aria-label="Log filters">
          <select class="sg-input sg-select" formControlName="severity" aria-label="Minimum severity">
            <option value="">All severities</option>
            @for (s of severities; track s) { <option [value]="s">{{ s }}</option> }
          </select>
          <input class="sg-input" formControlName="component" placeholder="Component" aria-label="Filter by component" />
          <input class="sg-input" formControlName="nodeName" placeholder="Node" aria-label="Filter by node" />
          <input class="sg-input" formControlName="podName" placeholder="Pod" aria-label="Filter by pod" />
          <input class="sg-input" formControlName="search" placeholder="Search…" aria-label="Search log messages" />
          <button class="sg-btn sg-btn-secondary" type="submit">Apply</button>
          <button class="sg-btn sg-btn-secondary" [class.sg-btn-active]="tailing()" type="button" (click)="toggleTail()">
            {{ tailing() ? '⏸ Stop Tail' : '▶ Tail' }}
          </button>
        </form>
      </div>

      @if (loading()) {
        <sg-skeleton-list [count]="10" label="Loading logs…" />
      } @else if (loadError()) {
        <sg-error-state [error]="loadError()" (retry)="query()" />
      } @else {
        <div class="sg-log-stream" role="log" aria-live="off" aria-label="Log output" aria-atomic="false">
          @for (entry of entries(); track entry.id) {
            <div class="sg-log-line sg-log-{{ entry.severity.toLowerCase() }}" [attr.aria-label]="entry.severity + ': ' + entry.message">
              <span class="sg-log-time">{{ entry.timestamp | slice:0:23 }}</span>
              <span class="sg-log-sev">{{ entry.severity }}</span>
              <span class="sg-log-comp">{{ entry.component }}</span>
              @if (entry.nodeName) { <span class="sg-log-node">{{ entry.nodeName }}</span> }
              <span class="sg-log-msg">{{ entry.message }}</span>
              @if (entry.traceId) { <span class="sg-log-trace sg-mono-value">{{ entry.traceId | slice:0:8 }}</span> }
            </div>
          }
          @if (entries().length === 0) {
            <p class="sg-empty-inline">No log entries found</p>
          }
        </div>
      }
    </div>
  `,
  styles: [`
    .sg-logs-page { display:flex;flex-direction:column;height:100%;overflow:hidden; }
    .sg-logs-toolbar { padding:.75rem 1.5rem;border-bottom:1px solid var(--sg-border);flex-shrink:0;display:flex;flex-direction:column;gap:8px; }
    .sg-filter-bar { display:flex;gap:6px;flex-wrap:wrap;align-items:center; }
    .sg-filter-bar .sg-input { width:auto;min-width:100px; }
    .sg-log-stream { flex:1;overflow-y:auto;padding:.5rem 1.5rem;font-family:var(--sg-font-mono);font-size:.75rem;background:var(--sg-bg-base); }
    .sg-log-line { display:flex;gap:10px;padding:2px 0;align-items:baseline;border-bottom:1px solid rgba(148,163,184,.05); }
    .sg-log-time { color:var(--sg-text-muted);flex-shrink:0;white-space:nowrap; }
    .sg-log-sev  { width:42px;flex-shrink:0;font-weight:700; }
    .sg-log-comp { color:var(--sg-indigo);flex-shrink:0;max-width:140px;overflow:hidden;text-overflow:ellipsis;white-space:nowrap; }
    .sg-log-node { color:var(--sg-teal);flex-shrink:0; }
    .sg-log-msg  { color:var(--sg-text-primary);word-break:break-word; }
    .sg-log-trace { color:var(--sg-text-muted);margin-left:auto; }
    .sg-log-debug { opacity:.6; }
    .sg-log-info  { }
    .sg-log-warn  .sg-log-sev { color:var(--sg-degraded); }
    .sg-log-error .sg-log-sev,.sg-log-fatal .sg-log-sev { color:var(--sg-failed); }
    .sg-btn-active { border-color:var(--sg-accent);color:var(--sg-accent); }
    .sg-empty-inline { color:var(--sg-text-muted);padding:1rem 0; }
  `],
})
export class LogsComponent {
  private readonly logsApi    = inject(LogsApiService);
  readonly context            = inject(ContextService);
  private readonly destroyRef = inject(DestroyRef);
  private readonly platformId = inject(PLATFORM_ID);
  private readonly fb         = inject(FormBuilder);

  readonly severities = SEVERITIES;
  readonly loading    = signal(false);
  readonly loadError  = signal<ApiError | null>(null);
  readonly entries    = signal<LogEntry[]>([]);
  readonly tailing    = signal(false);
  private eventSource: EventSource | null = null;

  readonly filterForm = this.fb.nonNullable.group({
    severity:  [''],
    component: [''],
    nodeName:  [''],
    podName:   [''],
    search:    [''],
  });

  constructor() { this.query(); }

  query(): void {
    this.loading.set(true);
    this.loadError.set(null);
    const f = this.filterForm.getRawValue();
    this.logsApi.query({
      namespace:  this.context.selectedNamespace() || undefined,
      severity:   (f.severity as LogSeverity) || undefined,
      component:  f.component || undefined,
      nodeName:   f.nodeName  || undefined,
      podName:    f.podName   || undefined,
      search:     f.search    || undefined,
      tailLines:  500,
    }).pipe(takeUntilDestroyed(this.destroyRef)).subscribe({
      next:  (r)   => { this.entries.set(r.items); this.loading.set(false); },
      error: (err) => { this.loadError.set(err);   this.loading.set(false); },
    });
  }

  toggleTail(): void {
    if (this.tailing()) { this.closeTail(); return; }
    if (!isPlatformBrowser(this.platformId)) return;
    this.tailing.set(true);
    const f = this.filterForm.getRawValue();
    const url = this.logsApi.tailUrl({
      namespace: this.context.selectedNamespace() || undefined,
      severity:  (f.severity as LogSeverity) || undefined,
      component: f.component || undefined,
    });
    this.eventSource = new EventSource(url);
    this.eventSource.onmessage = (e) => {
      const entry: LogEntry = JSON.parse(e.data);
      this.entries.update((list) => [...list, entry].slice(-2000));
    };
    this.eventSource.onerror = () => this.closeTail();
  }

  private closeTail(): void {
    this.eventSource?.close();
    this.eventSource = null;
    this.tailing.set(false);
  }
}
