import { Component, inject, signal, DestroyRef } from '@angular/core';
import { takeUntilDestroyed } from '@angular/core/rxjs-interop';
import { EbpfApiService } from '../../core/api/resources/ebpf.api';
import { BreadcrumbsComponent } from '../../layout/breadcrumbs/breadcrumbs.component';
import { SkeletonListComponent } from '../../shared/components/skeleton/skeleton.component';
import { ErrorStateComponent } from '../../shared/components/error-state/error-state.component';
import { StatusBadgeComponent } from '../../shared/components/status-badge/status-badge.component';
import { EbpfProgram, EbpfMap, EbpfAttachment } from '../../core/models/ebpf.model';
import { ApiError } from '../../core/api/api-error';

type EbpfTab = 'programs' | 'maps' | 'attachments' | 'lb' | 'errors';

@Component({
  selector: 'sg-ebpf',
  imports: [BreadcrumbsComponent, SkeletonListComponent, ErrorStateComponent, StatusBadgeComponent],
  template: `
    <div class="sg-page">
      <sg-breadcrumbs />
      <header class="sg-page-header">
        <h1 class="sg-page-title">eBPF</h1>
        <button class="sg-btn sg-btn-secondary" (click)="load()">Refresh</button>
      </header>

      <!-- Tab nav -->
      <nav class="sg-tab-nav" aria-label="eBPF sections" role="tablist">
        @for (tab of tabs; track tab.id) {
          <button
            class="sg-tab-btn"
            [class.sg-tab-active]="activeTab() === tab.id"
            role="tab"
            [attr.aria-selected]="activeTab() === tab.id"
            [attr.id]="'ebpf-tab-' + tab.id"
            [attr.aria-controls]="'ebpf-panel-' + tab.id"
            (click)="activeTab.set(tab.id)"
          >{{ tab.label }}</button>
        }
      </nav>

      <div class="sg-tab-content" [attr.id]="'ebpf-panel-' + activeTab()" role="tabpanel" [attr.aria-labelledby]="'ebpf-tab-' + activeTab()">
        @if (loading()) {
          <sg-skeleton-list [count]="6" />
        } @else if (loadError()) {
          <sg-error-state [error]="loadError()" (retry)="load()" />
        } @else {

          @if (activeTab() === 'programs') {
            <table class="sg-table" aria-label="eBPF Programs">
              <thead><tr>
                <th scope="col">Name</th>
                <th scope="col">Type</th>
                <th scope="col">Hook</th>
                <th scope="col">Node</th>
                <th scope="col">State</th>
                <th scope="col">Error</th>
              </tr></thead>
              <tbody>
                @for (p of programs(); track p.id) {
                  <tr>
                    <td class="sg-mono-value">{{ p.name }}</td>
                    <td>{{ p.type }}</td>
                    <td><span class="sg-badge sg-badge-info">{{ p.hook }}</span></td>
                    <td>{{ p.nodeName }}</td>
                    <td><sg-status-badge [variant]="p.state === 'Loaded' ? 'healthy' : p.state === 'Failed' ? 'failed' : 'unknown'" /></td>
                    <td class="sg-ebpf-error">{{ p.error ?? '—' }}</td>
                  </tr>
                }
              </tbody>
            </table>
          }

          @if (activeTab() === 'maps') {
            <table class="sg-table" aria-label="eBPF Maps">
              <thead><tr>
                <th scope="col">Name</th>
                <th scope="col">Type</th>
                <th scope="col">Key Size</th>
                <th scope="col">Value Size</th>
                <th scope="col">Max Entries</th>
                <th scope="col">Node</th>
              </tr></thead>
              <tbody>
                @for (m of maps(); track m.id) {
                  <tr>
                    <td class="sg-mono-value">{{ m.name }}</td>
                    <td>{{ m.type }}</td>
                    <td class="sg-mono-value">{{ m.keySize }}</td>
                    <td class="sg-mono-value">{{ m.valueSize }}</td>
                    <td class="sg-mono-value">{{ m.maxEntries }}</td>
                    <td>{{ m.nodeName }}</td>
                  </tr>
                }
              </tbody>
            </table>
          }

          @if (activeTab() === 'attachments') {
            <table class="sg-table" aria-label="eBPF Attachments">
              <thead><tr>
                <th scope="col">Program</th>
                <th scope="col">Hook</th>
                <th scope="col">Interface</th>
                <th scope="col">Cgroup</th>
                <th scope="col">Node</th>
                <th scope="col">Attached At</th>
              </tr></thead>
              <tbody>
                @for (a of attachments(); track a.programId + a.nodeName) {
                  <tr>
                    <td class="sg-mono-value">{{ a.programName }}</td>
                    <td><span class="sg-badge sg-badge-info">{{ a.hook }}</span></td>
                    <td class="sg-mono-value">{{ a.interfaceName ?? '—' }}</td>
                    <td class="sg-mono-value sg-truncate">{{ a.cgroupPath ?? '—' }}</td>
                    <td>{{ a.nodeName }}</td>
                    <td class="sg-mono-value">{{ a.attachedAt ?? '—' }}</td>
                  </tr>
                }
              </tbody>
            </table>
          }

          @if (activeTab() === 'errors') {
            @for (err of allErrors(); track $index) {
              <div class="sg-ebpf-err-row" role="alert">
                <span class="sg-badge sg-badge-failed">Error</span>
                <span class="sg-mono-value">{{ err }}</span>
              </div>
            }
            @if (allErrors().length === 0) {
              <p class="sg-empty-inline">No eBPF errors detected</p>
            }
          }
        }
      </div>
    </div>
  `,
  styles: [`
    .sg-tab-nav { display:flex;gap:2px;border-bottom:1px solid var(--sg-border);margin-bottom:1rem; }
    .sg-tab-btn { padding:8px 16px;border:none;background:none;color:var(--sg-text-secondary);font-size:.8125rem;font-weight:500;cursor:pointer;border-bottom:2px solid transparent;transition:all var(--sg-transition-fast); }
    .sg-tab-btn:hover { color:var(--sg-text-primary); }
    .sg-tab-active { color:var(--sg-accent) !important;border-bottom-color:var(--sg-accent) !important; }
    .sg-ebpf-error { color:var(--sg-failed);font-size:.75rem;font-family:var(--sg-font-mono);max-width:300px; }
    .sg-ebpf-err-row { display:flex;align-items:center;gap:10px;padding:8px;border-bottom:1px solid var(--sg-border); }
    .sg-empty-inline { color:var(--sg-text-muted);font-size:.8125rem;padding:1rem 0; }
    .sg-badge-failed { background:rgba(248,113,113,.15);color:var(--sg-failed); }
  `],
})
export class EbpfComponent {
  private readonly ebpfApi  = inject(EbpfApiService);
  private readonly destroyRef = inject(DestroyRef);

  readonly activeTab  = signal<EbpfTab>('programs');
  readonly loading    = signal(false);
  readonly loadError  = signal<ApiError | null>(null);
  readonly programs   = signal<EbpfProgram[]>([]);
  readonly maps       = signal<EbpfMap[]>([]);
  readonly attachments = signal<EbpfAttachment[]>([]);
  readonly allErrors  = signal<string[]>([]);

  readonly tabs: { id: EbpfTab; label: string }[] = [
    { id: 'programs',    label: 'Programs' },
    { id: 'maps',        label: 'Maps' },
    { id: 'attachments', label: 'Attachments (cgroup · tcx · xdp · lsm)' },
    { id: 'errors',      label: 'Errors' },
  ];

  constructor() { this.load(); }

  load(): void {
    this.loading.set(true);
    this.loadError.set(null);

    this.ebpfApi.listPrograms().pipe(takeUntilDestroyed(this.destroyRef)).subscribe({
      next:  (p) => { this.programs.set(p); this.allErrors.set(p.filter(x => x.error).map(x => `${x.name}: ${x.error}`)); },
      error: (e) => { this.loadError.set(e); this.loading.set(false); },
    });
    this.ebpfApi.listMaps().pipe(takeUntilDestroyed(this.destroyRef)).subscribe({
      next: (m) => this.maps.set(m),
    });
    this.ebpfApi.listAttachments().pipe(takeUntilDestroyed(this.destroyRef)).subscribe({
      next: (a) => { this.attachments.set(a); this.loading.set(false); },
    });
  }
}
