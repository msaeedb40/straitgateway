// Copyright 2026 straitgateway Authors
// SPDX-License-Identifier: Apache-2.0

import { Component, inject } from '@angular/core';
import { CommonModule } from '@angular/common';
import { Router, NavigationEnd, RouterLink } from '@angular/router';
import { filter } from 'rxjs/operators';
import { toSignal } from '@angular/core/rxjs-interop';

@Component({
  selector: 'sg-breadcrumbs',
  standalone: true,
  imports: [CommonModule, RouterLink],
  template: `
    <nav aria-label="Breadcrumb" class="sg-breadcrumbs">
      <a routerLink="/" style="color:var(--sg-text-2);display:flex;align-items:center;gap:4px">
        <svg width="12" height="12" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
          <path d="m3 9 9-7 9 7v11a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2z"/>
          <polyline points="9 22 9 12 15 12 15 22"/>
        </svg>
        <span>Home</span>
      </a>
      @if (currentRoute()?.urlAfterRedirects && currentRoute()?.urlAfterRedirects !== '/') {
        <span>/</span>
        <span style="color:var(--sg-text);text-transform:capitalize;font-weight:500">
          {{ formatSegment(currentRoute()) }}
        </span>
      }
    </nav>
  `,
})
export class BreadcrumbsComponent {
  private router = inject(Router);

  readonly currentRoute = toSignal(
    this.router.events.pipe(
      filter((e): e is NavigationEnd => e instanceof NavigationEnd),
      filter(e => !!e.urlAfterRedirects)
    ),
    { initialValue: { urlAfterRedirects: this.router.url } as NavigationEnd }
  );

  formatSegment(nav?: NavigationEnd | null): string {
    const raw = (nav?.urlAfterRedirects || this.router.url || '').split('?')[0].replace(/^\//, '');
    return raw ? raw.replace(/-/g, ' ') : 'Dashboard';
  }
}
