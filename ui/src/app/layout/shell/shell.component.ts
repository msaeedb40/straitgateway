// Copyright 2026 straitgateway Authors
// SPDX-License-Identifier: Apache-2.0

import { Component, HostListener, inject, signal } from '@angular/core';
import { CommonModule } from '@angular/common';
import { RouterOutlet } from '@angular/router';
import { HeaderComponent } from '../header/header.component';
import { SidebarComponent } from '../sidebar/sidebar.component';
import { BreadcrumbsComponent } from '../breadcrumbs/breadcrumbs.component';
import { CommandBarComponent } from '../command-bar/command-bar.component';
import { ContextMenuComponent } from '../context-menu/context-menu.component';
import { NotificationService } from '../../core/services/notification.service';

@Component({
  selector: 'sg-shell',
  standalone: true,
  imports: [
    CommonModule,
    RouterOutlet,
    HeaderComponent,
    SidebarComponent,
    BreadcrumbsComponent,
    CommandBarComponent,
    ContextMenuComponent,
  ],
  template: `
    <div class="sg-shell">
      <!-- Header -->
      <sg-header
        (openCommandBar)="showCommandBar.set(true)"
        (toggleSidebar)="sidebarOpen.update(v => !v)"
      />

      <!-- Sidebar -->
      <sg-sidebar [class.open]="sidebarOpen()" />

      <!-- Backdrop overlay on mobile when sidebar open -->
      @if (sidebarOpen()) {
        <div
          (click)="sidebarOpen.set(false)"
          style="position:fixed;inset:var(--sg-header-h) 0 0 0;background:rgba(0,0,0,0.6);z-index:998"
        ></div>
      }

      <!-- Main Workspace -->
      <main
        class="sg-main"
        (click)="sidebarOpen.set(false)"
        style="overflow-y:auto;height:calc(100vh - var(--sg-header-h))"
      >
        <div class="sg-main-inner">
          <sg-breadcrumbs />
          <router-outlet />
        </div>
      </main>

      <!-- Command Bar Modal -->
      @if (showCommandBar()) {
        <sg-command-bar (close)="showCommandBar.set(false)" />
      }

      <!-- Global Context Menu -->
      <sg-context-menu />

      <!-- Toast Notifications Container -->
      <aside aria-label="Notifications" class="sg-toast-container">
        @for (item of notif.notifications(); track item.id) {
          <div
            class="sg-toast"
            [style.border-left-color]="item.type === 'danger' ? 'var(--sg-danger)' : item.type === 'warn' ? 'var(--sg-warn)' : item.type === 'success' ? 'var(--sg-success)' : 'var(--sg-accent)'"
          >
            <div>
              <div style="font-size:13px;font-weight:600;color:var(--sg-text)">{{ item.title }}</div>
              <div style="font-size:12px;color:var(--sg-text-2);margin-top:4px">{{ item.message }}</div>
            </div>
            <button
              (click)="notif.dismiss(item.id)"
              style="color:var(--sg-text-3);font-size:16px;line-height:1;padding:4px;cursor:pointer"
            >×</button>
          </div>
        }
      </aside>
    </div>
  `,
})
export class ShellComponent {
  readonly notif = inject(NotificationService);
  showCommandBar = signal<boolean>(false);
  sidebarOpen = signal<boolean>(false);

  @HostListener('window:keydown', ['$event'])
  handleKeyboardEvent(event: KeyboardEvent) {
    if ((event.ctrlKey || event.metaKey) && event.key.toLowerCase() === 'k') {
      event.preventDefault();
      this.showCommandBar.update((open) => !open);
    }
  }
}
