import {
  Component, computed, inject, input, output, signal, OnInit, DestroyRef
} from '@angular/core';
import { takeUntilDestroyed } from '@angular/core/rxjs-interop';
import { RouterLink } from '@angular/router';
import { ContextService } from '../../core/services/context.service';
import { ClusterApiService } from '../../core/api/resources/cluster.api';
import { NamespaceApiService } from '../../core/api/resources/namespace.api';
import type { Cluster } from '../../core/api/resources/cluster.api';
import type { Namespace } from '../../core/api/resources/namespace.api';

@Component({
  selector: 'sg-header',
  imports: [RouterLink],
  template: `
    <header class="sg-header" role="banner">
      <!-- Logo + sidebar toggle -->
      <div class="sg-header-left">
        <button
          class="sg-sidebar-toggle sg-btn-ghost"
          (click)="toggleSidebar.emit()"
          [attr.aria-expanded]="!sidebarCollapsed()"
          aria-controls="sg-sidebar"
          aria-label="Toggle navigation sidebar"
        >
          <svg width="18" height="18" viewBox="0 0 18 18" fill="none" aria-hidden="true">
            <rect x="2" y="4" width="14" height="1.5" rx="1" fill="currentColor"/>
            <rect x="2" y="8.25" width="14" height="1.5" rx="1" fill="currentColor"/>
            <rect x="2" y="12.5" width="14" height="1.5" rx="1" fill="currentColor"/>
          </svg>
        </button>
        <a routerLink="/dashboard" class="sg-logo" aria-label="StraitGateway — go to dashboard">
          <span class="sg-logo-text">StraitGateway</span>
        </a>
      </div>

      <!-- Context selectors -->
      <div class="sg-header-context" role="group" aria-label="Active cluster and namespace">
        <!-- Cluster selector -->
        <label class="sg-context-label" for="sg-cluster-select">Cluster</label>
        <select
          id="sg-cluster-select"
          class="sg-input sg-select sg-context-select"
          [value]="context.selectedCluster() ?? ''"
          (change)="onClusterChange($event)"
          [attr.aria-label]="'Selected cluster: ' + (context.selectedCluster() ?? 'none')"
        >
          <option value="" disabled>{{ clustersLoading() ? 'Loading…' : 'Select cluster' }}</option>
          @for (c of clusters(); track c.name) {
            <option [value]="c.name">{{ c.name }}</option>
          }
        </select>

        <span class="sg-context-sep" aria-hidden="true">/</span>

        <!-- Namespace selector -->
        <label class="sg-context-label" for="sg-namespace-select">Namespace</label>
        <select
          id="sg-namespace-select"
          class="sg-input sg-select sg-context-select"
          [value]="context.selectedNamespace()"
          (change)="onNamespaceChange($event)"
          [disabled]="!context.selectedCluster()"
          [attr.aria-label]="'Selected namespace: ' + context.selectedNamespace()"
        >
          <option value="" disabled>{{ namespacesLoading() ? 'Loading…' : 'Select namespace' }}</option>
          @for (n of namespaces(); track n.name) {
            <option [value]="n.name">{{ n.name }}</option>
          }
        </select>
      </div>

      <!-- Right actions -->
      <div class="sg-header-right">
        <!-- Alert count -->
        @if (alertCount() > 0) {
          <a routerLink="/events" class="sg-alert-badge" [attr.aria-label]="alertCount() + ' active alerts'">
            {{ alertCount() }}
          </a>
        }

        <!-- User menu placeholder — wired to auth -->
        <div class="sg-user-menu" role="button" tabindex="0" aria-label="User menu" aria-haspopup="menu">
          <div class="sg-user-avatar" aria-hidden="true">U</div>
        </div>
      </div>
    </header>
  `,
  styles: [`
    .sg-header {
      grid-row: 1;
      display: flex;
      align-items: center;
      gap: 1rem;
      padding: 0 1rem;
      height: var(--sg-header-height);
      background: var(--sg-bg-surface);
      border-bottom: 1px solid var(--sg-border);
      position: sticky;
      top: 0;
      z-index: 100;
    }
    .sg-header-left { display: flex; align-items: center; gap: 12px; flex-shrink: 0; }
    .sg-sidebar-toggle {
      display: flex; align-items: center; justify-content: center;
      width: 36px; height: 36px; border-radius: var(--sg-radius);
      cursor: pointer; color: var(--sg-text-secondary);
      transition: all var(--sg-transition-fast);
      background: none; border: none;
    }
    .sg-sidebar-toggle:hover { background: var(--sg-bg-hover); color: var(--sg-text-primary); }

    .sg-logo { display: flex; align-items: center; gap: 8px; text-decoration: none; }
    .sg-logo-text {
      font-size: 0.9375rem; font-weight: 700; color: var(--sg-accent);
      letter-spacing: -0.02em; white-space: nowrap;
    }

    .sg-header-context {
      display: flex; align-items: center; gap: 8px;
      flex: 1; min-width: 0;
    }
    .sg-context-label {
      font-size: 0.75rem; font-weight: 500; color: var(--sg-text-muted);
      white-space: nowrap;
    }
    .sg-context-select { width: auto; min-width: 120px; max-width: 200px; font-size: 0.8125rem; }
    .sg-context-sep { color: var(--sg-text-muted); font-size: 1rem; }

    .sg-header-right { display: flex; align-items: center; gap: 8px; flex-shrink: 0; }
    .sg-alert-badge {
      display: inline-flex; align-items: center; justify-content: center;
      min-width: 22px; height: 22px; padding: 0 6px;
      background: rgba(248, 113, 113, 0.2); color: var(--sg-failed);
      border-radius: 999px; font-size: 0.75rem; font-weight: 700;
      text-decoration: none;
      transition: background var(--sg-transition-fast);
    }
    .sg-alert-badge:hover { background: rgba(248, 113, 113, 0.35); }

    .sg-user-menu { cursor: pointer; }
    .sg-user-avatar {
      width: 32px; height: 32px; border-radius: 50%;
      background: var(--sg-bg-overlay);
      border: 1px solid var(--sg-border);
      display: flex; align-items: center; justify-content: center;
      font-size: 0.75rem; font-weight: 600; color: var(--sg-text-secondary);
    }
  `],
})
export class HeaderComponent implements OnInit {
  readonly toggleSidebar = output<void>();
  readonly sidebarCollapsed = input<boolean>(false);

  readonly context = inject(ContextService);
  private readonly clusterApi = inject(ClusterApiService);
  private readonly namespaceApi = inject(NamespaceApiService);
  private readonly destroyRef = inject(DestroyRef);

  readonly clusters = signal<Cluster[]>([]);
  readonly namespaces = signal<Namespace[]>([]);
  readonly clustersLoading = signal(false);
  readonly namespacesLoading = signal(false);
  readonly alertCount = signal(0);

  ngOnInit(): void {
    this.loadClusters();
  }

  private loadClusters(): void {
    this.clustersLoading.set(true);
    this.clusterApi.list().pipe(takeUntilDestroyed(this.destroyRef)).subscribe({
      next: (list) => { this.clusters.set(list); this.clustersLoading.set(false); },
      error: () => this.clustersLoading.set(false),
    });
  }

  onClusterChange(event: Event): void {
    const value = (event.target as HTMLSelectElement).value;
    this.context.setCluster(value || null);
    this.loadNamespaces(value);
  }

  private loadNamespaces(cluster: string): void {
    if (!cluster) { this.namespaces.set([]); return; }
    this.namespacesLoading.set(true);
    this.namespaceApi.list(cluster).pipe(takeUntilDestroyed(this.destroyRef)).subscribe({
      next: (list) => { this.namespaces.set(list); this.namespacesLoading.set(false); },
      error: () => this.namespacesLoading.set(false),
    });
  }

  onNamespaceChange(event: Event): void {
    const value = (event.target as HTMLSelectElement).value;
    this.context.setNamespace(value);
  }
}
