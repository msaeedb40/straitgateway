// Copyright 2026 straitgateway Authors
// SPDX-License-Identifier: Apache-2.0

import { Component, inject, output } from '@angular/core';
import { CommonModule } from '@angular/common';
import { RouterLink } from '@angular/router';
import { NamespaceService } from '../../core/services/namespace.service';
import { ConnectionService } from '../../core/services/connection.service';
import { RuntimeConfigService } from '../../core/config/runtime-config';

@Component({
  selector: 'sg-header',
  standalone: true,
  imports: [CommonModule, RouterLink],
  template: `
    <header class="sg-header" role="banner">
      <!-- Mobile Menu Button -->
      <button
        class="sg-mobile-menu-btn"
        (click)="toggleSidebar.emit()"
        aria-label="Toggle Navigation Menu"
        style="display:none;align-items:center;justify-content:center;padding:8px;border-radius:var(--sg-radius-sm);color:var(--sg-text)"
      >
        <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
          <line x1="3" y1="12" x2="21" y2="12" /><line x1="3" y1="6" x2="21" y2="6" /><line x1="3" y1="18" x2="21" y2="18" />
        </svg>
      </button>

      <!-- Logo -->
      <a routerLink="/" class="sg-logo" aria-label="straitgateway home">
        <svg viewBox="0 0 32 32" fill="none" width="28" height="28">
          <rect width="32" height="32" rx="8" fill="#6366f1" fill-opacity=".18" />
          <path d="M8 16 L16 8 L24 16 L16 24 Z" stroke="#818cf8" stroke-width="2" fill="none" />
          <circle cx="16" cy="16" r="3" fill="#6366f1" />
        </svg>
        <span class="font-bold tracking-tight">straitgateway</span>
      </a>

      <!-- Search Trigger -->
      <div class="sg-header-search">
        <div class="sg-search-wrap" (click)="openCommandBar.emit()" style="cursor:pointer">
          <svg class="sg-search-icon" width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
            <circle cx="11" cy="11" r="8" /><line x1="21" y1="21" x2="16.65" y2="16.65" />
          </svg>
          <input
            type="text"
            class="sg-search-input"
            placeholder="Search resources, flows, docs... (Ctrl+K)"
            readonly
          />
        </div>
      </div>

      <div class="sg-header-spacer"></div>

      <!-- Namespace Selector -->
      <div class="sg-ns-select-wrap" style="display:flex;align-items:center;gap:8px">
        <label for="sg-ns-select" class="text-xs text-muted" style="color:var(--sg-text-3)">Namespace:</label>
        <select
          id="sg-ns-select"
          class="sg-select"
          style="height:32px;padding:0 12px;font-size:12px"
          [value]="nsService.active()"
          (change)="onNamespaceChange($event)"
        >
          <option value="">All Namespaces</option>
          @for (ns of nsService.namespaces(); track ns) {
            <option [value]="ns">{{ ns }}</option>
          }
        </select>
      </div>

      <!-- Cluster Badge -->
      <div class="sg-cluster-badge" style="padding:5px 10px;background:var(--sg-surface-2);border:1px solid var(--sg-border);border-radius:var(--sg-radius-sm);font-size:11px;font-family:monospace;color:var(--sg-accent-light)">
        {{ config.clusterName() }}
      </div>

      <!-- Live Connection Status -->
      <div
        class="sg-connection-badge"
        [title]="'Latency: ' + conn.latencyMs() + 'ms'"
        aria-live="polite"
      >
        <span
          class="dot"
          [style.background]="conn.isConnected() ? 'var(--sg-success)' : 'var(--sg-danger)'"
        ></span>
        <span>{{ conn.isConnected() ? 'Connected' : 'Disconnected' }}</span>
        <span style="color:var(--sg-text-3);font-size:10px">({{ conn.latencyMs() }}ms)</span>
      </div>
    </header>
  `,
})
export class HeaderComponent {
  readonly nsService = inject(NamespaceService);
  readonly conn = inject(ConnectionService);
  readonly config = inject(RuntimeConfigService);

  openCommandBar = output<void>();
  toggleSidebar = output<void>();

  onNamespaceChange(event: Event) {
    const target = event.target as HTMLSelectElement;
    this.nsService.select(target.value);
  }
}
