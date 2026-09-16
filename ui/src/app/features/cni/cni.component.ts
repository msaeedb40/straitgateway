import { Component, inject, signal, DestroyRef } from '@angular/core';
import { takeUntilDestroyed } from '@angular/core/rxjs-interop';
import { CniApiService } from '../../core/api/resources/cni.api';
import { BreadcrumbsComponent } from '../../layout/breadcrumbs/breadcrumbs.component';
import { SkeletonListComponent } from '../../shared/components/skeleton/skeleton.component';
import { ErrorStateComponent } from '../../shared/components/error-state/error-state.component';
import { CniSummary } from '../../core/models/cni.model';
import { ApiError } from '../../core/api/api-error';

@Component({
  selector: 'sg-cni',
  imports: [BreadcrumbsComponent, SkeletonListComponent, ErrorStateComponent],
  template: `
    <div class="sg-page">
      <sg-breadcrumbs />
      <header class="sg-page-header">
        <h1 class="sg-page-title">CNI</h1>
        <button class="sg-btn sg-btn-secondary" (click)="load()">Refresh</button>
      </header>

      @if (loading()) {
        <sg-skeleton-list [count]="5" />
      } @else if (loadError()) {
        <sg-error-state [error]="loadError()" (retry)="load()" />
      } @else if (summary()) {
        <div class="sg-cni-summary sg-card">
          <dl class="sg-panel-dl">
            <dt>Plugin</dt><dd>{{ summary()!.config?.type ?? summary()!.config?.name ?? '—' }}</dd>
            <dt>Version</dt><dd class="sg-mono-value">{{ summary()!.config?.cniVersion ?? '—' }}</dd>
            <dt>Phase</dt><dd class="sg-mono-value">{{ summary()!.phase }}</dd>
            <dt>Nodes</dt><dd class="sg-mono-value">{{ summary()!.nodeStatuses.length }}</dd>
          </dl>
        </div>

        <section class="sg-card" aria-labelledby="cni-nodes-heading">
          <h2 id="cni-nodes-heading" class="sg-section-title">Node CNI Status</h2>
          <table class="sg-table" aria-label="CNI node status">
            <thead><tr>
              <th scope="col">Node</th>
              <th scope="col">Status</th>
              <th scope="col">Pod CIDR</th>
              <th scope="col">Version</th>
              <th scope="col">Errors</th>
            </tr></thead>
            <tbody>
              @for (n of summary()!.nodeStatuses; track n.nodeName) {
                <tr>
                  <td>{{ n.nodeName }}</td>
                  <td><span class="sg-mono-value">{{ n.phase }}</span></td>
                  <td class="sg-mono-value">{{ n.podCIDR ?? '—' }}</td>
                  <td class="sg-mono-value">{{ n.version ?? '—' }}</td>
                  <td class="sg-mono-value">
                    @if (n.errors?.length) {
                      @for (e of n.errors; track $index) {
                        <div style="color:var(--sg-failed)">{{ e }}</div>
                      }
                    } @else { — }
                  </td>
                </tr>
              }
            </tbody>
          </table>
        </section>
      }
    </div>
  `,
  styles: [`.sg-section-title{font-size:.8125rem;font-weight:600;color:var(--sg-text-secondary);text-transform:uppercase;letter-spacing:.05em;margin-bottom:1rem;} .sg-cni-summary{margin-bottom:1rem;} .sg-panel-dl{display:grid;grid-template-columns:auto 1fr;gap:6px 16px;font-size:.875rem;} .sg-panel-dl dt{color:var(--sg-text-muted);} .sg-panel-dl dd{color:var(--sg-text-primary);}`],
})
export class CniComponent {
  private readonly cniApi    = inject(CniApiService);
  private readonly destroyRef = inject(DestroyRef);
  readonly loading   = signal(false);
  readonly loadError = signal<ApiError | null>(null);
  readonly summary   = signal<CniSummary | null>(null);
  constructor() { this.load(); }
  load(): void {
    this.loading.set(true);
    this.loadError.set(null);
    this.cniApi.getSummary().pipe(takeUntilDestroyed(this.destroyRef)).subscribe({
      next: (s)   => { this.summary.set(s); this.loading.set(false); },
      error: (err) => { this.loadError.set(err); this.loading.set(false); },
    });
  }
}
