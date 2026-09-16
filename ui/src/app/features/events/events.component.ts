import { Component, inject, signal, DestroyRef, PLATFORM_ID } from '@angular/core';
import { isPlatformBrowser, SlicePipe } from '@angular/common';
import { takeUntilDestroyed } from '@angular/core/rxjs-interop';
import { ReactiveFormsModule, FormBuilder } from '@angular/forms';
import { EventApiService } from '../../core/api/resources/event.api';
import { BreadcrumbsComponent } from '../../layout/breadcrumbs/breadcrumbs.component';
import { SkeletonListComponent } from '../../shared/components/skeleton/skeleton.component';
import { ErrorStateComponent } from '../../shared/components/error-state/error-state.component';
import { StatusBadgeComponent } from '../../shared/components/status-badge/status-badge.component';
import { ContextService } from '../../core/services/context.service';
import { StraitEvent, EventSeverity, EventResourceKind } from '../../core/models/event.model';
import { ApiError } from '../../core/api/api-error';

const SEVERITIES: EventSeverity[] = ['Normal', 'Warning', 'Error'];

@Component({
  selector: 'sg-events',
  imports: [ReactiveFormsModule, SlicePipe, BreadcrumbsComponent, SkeletonListComponent, ErrorStateComponent, StatusBadgeComponent],
  template: `
    <div class="sg-page">
      <sg-breadcrumbs />
      <header class="sg-page-header">
        <h1 class="sg-page-title">Events</h1>
        <div class="sg-header-actions">
          <button class="sg-btn sg-btn-secondary" [class.sg-btn-active]="streaming()" (click)="toggleStream()" type="button">
            {{ streaming() ? '⏸ Stop Live' : '▶ Live' }}
          </button>
          <button class="sg-btn sg-btn-secondary" (click)="load()" type="button">Refresh</button>
        </div>
      </header>

      <!-- Filter bar -->
      <form [formGroup]="filterForm" (ngSubmit)="load()" class="sg-filter-bar">
        <select class="sg-input sg-select" formControlName="severity" aria-label="Filter by severity">
          <option value="">All severities</option>
          @for (s of severities; track s) { <option [value]="s">{{ s }}</option> }
        </select>
        <input class="sg-input" formControlName="resourceKind" placeholder="Resource kind" aria-label="Filter by resource kind" />
        <input class="sg-input" formControlName="search" placeholder="Search message…" aria-label="Search event messages" />
        <button class="sg-btn sg-btn-secondary" type="submit">Filter</button>
      </form>

      @if (loading()) {
        <sg-skeleton-list [count]="8" />
      } @else if (loadError()) {
        <sg-error-state [error]="loadError()" (retry)="load()" />
      } @else {
        <div class="sg-table-wrapper" role="region" aria-label="Events">
          <table class="sg-table" aria-rowcount="{{ events().length }}">
            <thead><tr>
              <th scope="col">Time</th>
              <th scope="col">Severity</th>
              <th scope="col">Resource</th>
              <th scope="col">Reason</th>
              <th scope="col">Message</th>
              <th scope="col">Count</th>
            </tr></thead>
            <tbody>
              @for (ev of events(); track ev.uid; let i = $index) {
                <tr [attr.aria-rowindex]="i + 1">
                  <td class="sg-mono-value" style="white-space:nowrap">{{ ev.timestamp | slice:0:19 }}</td>
                  <td>
                    <sg-status-badge [variant]="ev.severity === 'Error' ? 'failed' : ev.severity === 'Warning' ? 'degraded' : 'healthy'" />
                  </td>
                  <td>{{ ev.resourceKind }}/{{ ev.resourceName }}</td>
                  <td>{{ ev.reason }}</td>
                  <td>{{ ev.message }}</td>
                  <td class="sg-mono-value">{{ ev.count }}</td>
                </tr>
              }
            </tbody>
          </table>
          @if (events().length === 0) {
            <p class="sg-empty-inline">No events match the current filter</p>
          }
        </div>
      }
    </div>
  `,
  styles: [`
    .sg-header-actions{display:flex;gap:8px;}
    .sg-filter-bar{display:flex;gap:8px;flex-wrap:wrap;align-items:center;}
    .sg-filter-bar .sg-input{width:auto;min-width:140px;}
    .sg-btn-active{border-color:var(--sg-accent);color:var(--sg-accent);}
    .sg-empty-inline{color:var(--sg-text-muted);font-size:.8125rem;padding:1rem;}
  `],
})
export class EventsComponent {
  private readonly eventApi   = inject(EventApiService);
  readonly context            = inject(ContextService);
  private readonly destroyRef = inject(DestroyRef);
  private readonly platformId = inject(PLATFORM_ID);
  private readonly fb         = inject(FormBuilder);

  readonly severities = SEVERITIES;
  readonly loading    = signal(false);
  readonly loadError  = signal<ApiError | null>(null);
  readonly events     = signal<StraitEvent[]>([]);
  readonly streaming  = signal(false);
  private eventSource: EventSource | null = null;

  readonly filterForm = this.fb.nonNullable.group({
    severity:     [''],
    resourceKind: [''],
    search:       [''],
  });

  constructor() { this.load(); }

  load(): void {
    this.loading.set(true);
    this.loadError.set(null);
    const { severity, resourceKind, search } = this.filterForm.getRawValue();
    this.eventApi.list({
      namespace:    this.context.selectedNamespace() || undefined,
      cluster:      this.context.selectedCluster()   ?? undefined,
      severity:     (severity as EventSeverity)      || undefined,
      resourceKind: (resourceKind as EventResourceKind) || undefined,
      search:       search                           || undefined,
      pageSize:     100,
    }).pipe(takeUntilDestroyed(this.destroyRef)).subscribe({
      next:  (r)   => { this.events.set(r.items); this.loading.set(false); },
      error: (err) => { this.loadError.set(err);  this.loading.set(false); },
    });
  }

  toggleStream(): void {
    if (this.streaming()) { this.closeStream(); return; }
    if (!isPlatformBrowser(this.platformId)) return;
    this.streaming.set(true);
    this.eventSource = new EventSource(this.eventApi.streamUrl());
    this.eventSource.onmessage = (e) => {
      const ev: StraitEvent = JSON.parse(e.data);
      this.events.update((list) => [ev, ...list].slice(0, 500));
    };
    this.eventSource.onerror = () => this.closeStream();
  }

  private closeStream(): void {
    this.eventSource?.close();
    this.eventSource = null;
    this.streaming.set(false);
  }
}
