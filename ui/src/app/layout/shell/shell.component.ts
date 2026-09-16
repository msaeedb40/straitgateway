import { Component, inject, OnInit, OnDestroy, signal, computed } from '@angular/core';
import { RouterOutlet } from '@angular/router';
import { ConnectionService } from '../../core/services/connection.service';
import { HeaderComponent } from '../header/header.component';
import { SidebarComponent } from '../sidebar/sidebar.component';
import { ConnectionStatusComponent } from '../overlays/connection-status.component';
import { PreferencesService } from '../../core/services/preferences.service';

@Component({
  selector: 'sg-shell',
  imports: [RouterOutlet, HeaderComponent, SidebarComponent, ConnectionStatusComponent],
  template: `
    <div class="sg-shell" [class.sg-sidebar-collapsed]="sidebarCollapsed()">
      <!-- Skip to main content for accessibility -->
      <a href="#sg-main-content" class="sg-skip-link">Skip to main content</a>

      <sg-header
        (toggleSidebar)="toggleSidebar()"
        [sidebarCollapsed]="sidebarCollapsed()"
      />

      <div class="sg-shell-body">
        <sg-sidebar [collapsed]="sidebarCollapsed()" />

        <main
          id="sg-main-content"
          class="sg-main-content"
          tabindex="-1"
          role="main"
        >
          <router-outlet />
        </main>
      </div>

      <sg-connection-status />
    </div>
  `,
  styles: [`
    .sg-shell {
      display: grid;
      grid-template-rows: var(--sg-header-height) 1fr;
      grid-template-columns: 1fr;
      height: 100dvh;
      overflow: hidden;
      background: var(--sg-bg-base);
    }

    .sg-shell-body {
      display: grid;
      grid-template-columns: var(--sg-sidebar-width) 1fr;
      grid-row: 2;
      min-height: 0;
      transition: grid-template-columns var(--sg-transition);
    }

    .sg-shell.sg-sidebar-collapsed .sg-shell-body {
      grid-template-columns: var(--sg-sidebar-collapsed) 1fr;
    }

    .sg-main-content {
      overflow-y: auto;
      overflow-x: hidden;
      background: var(--sg-bg-base);
      min-width: 0;
    }

    .sg-skip-link {
      position: absolute;
      top: -100%;
      left: 1rem;
      z-index: 9999;
      background: var(--sg-accent);
      color: #0a0e1a;
      padding: 8px 16px;
      border-radius: 0 0 var(--sg-radius) var(--sg-radius);
      font-weight: 600;
      text-decoration: none;
      transition: top var(--sg-transition-fast);
    }
    .sg-skip-link:focus { top: 0; }

    @media (max-width: 768px) {
      .sg-shell-body {
        grid-template-columns: 0 1fr;
      }
      .sg-shell.sg-sidebar-collapsed .sg-shell-body {
        grid-template-columns: 0 1fr;
      }
    }
  `],
})
export class ShellComponent implements OnInit, OnDestroy {
  private readonly connection = inject(ConnectionService);
  private readonly preferences = inject(PreferencesService);

  readonly sidebarCollapsed = signal<boolean>(
    this.preferences.get<boolean>('sidebarCollapsed', false)
  );

  toggleSidebar(): void {
    this.sidebarCollapsed.update((v) => !v);
    this.preferences.set('sidebarCollapsed', this.sidebarCollapsed());
  }

  ngOnInit(): void {
    this.connection.startPolling();
  }

  ngOnDestroy(): void {
    this.connection.stopPolling();
  }
}
