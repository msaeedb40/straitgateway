// Copyright 2026 straitgateway Authors — SPDX-License-Identifier: Apache-2.0
import { Component, input } from '@angular/core';
import { CommonModule } from '@angular/common';
import { RouterLink, RouterLinkActive } from '@angular/router';

/** Top navigation bar for mobile / responsive layouts. */
@Component({
  selector: 'sg-navbar',
  standalone: true,
  imports: [CommonModule, RouterLink, RouterLinkActive],
  template: `
<nav class="sg-navbar" role="navigation" aria-label="Top navigation">
  <div class="sg-navbar-inner">
    <div class="sg-logo">
      <svg viewBox="0 0 32 32" fill="none" width="22" height="22">
        <rect width="32" height="32" rx="8" fill="#6366f1" fill-opacity=".18"/>
        <path d="M8 16 L16 8 L24 16 L16 24 Z" stroke="#818cf8" stroke-width="1.5" fill="none"/>
        <circle cx="16" cy="16" r="3" fill="#6366f1"/>
      </svg>
      <span>straitgateway</span>
    </div>
    <div class="sg-navbar-links">
      @for (link of links(); track link.route) {
        <a [routerLink]="link.route"
           routerLinkActive="active"
           [routerLinkActiveOptions]="{exact: link.route === '/'}"
           class="sg-navbar-link">
          {{ link.label }}
        </a>
      }
    </div>
  </div>
</nav>
`,
  styles: [`
    .sg-navbar {
      display: none;
      background: var(--sg-surface-1);
      border-bottom: 1px solid var(--sg-border);
      padding: 0 16px;
    }
    @media (max-width: 768px) { .sg-navbar { display: block; } }
    .sg-navbar-inner {
      display: flex; align-items: center; justify-content: space-between; height: 52px;
    }
    .sg-navbar-links { display: flex; gap: 4px; overflow-x: auto; }
    .sg-navbar-link {
      font-size: 12px; padding: 4px 8px; border-radius: 4px;
      color: var(--sg-text-2); text-decoration: none; white-space: nowrap;
    }
    .sg-navbar-link.active { color: var(--sg-accent); background: rgba(99,102,241,.1); }
  `],
})
export class NavbarComponent {
  links = input<{ label: string; route: string }[]>([]);
}
