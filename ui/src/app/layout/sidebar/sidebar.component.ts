// Copyright 2026 straitgateway Authors
// SPDX-License-Identifier: Apache-2.0

import { Component } from '@angular/core';
import { CommonModule } from '@angular/common';
import { RouterLink, RouterLinkActive } from '@angular/router';

interface NavSection {
  title: string;
  items: Array<{
    label: string;
    route: string;
    icon: string;
    badge?: string;
  }>;
}

@Component({
  selector: 'sg-sidebar',
  standalone: true,
  imports: [CommonModule, RouterLink, RouterLinkActive],
  template: `
    <aside class="sg-sidebar" role="navigation" aria-label="Main Navigation">
      <nav class="sg-nav">
        @for (section of sections; track section.title) {
          <div class="sg-nav-section">
            <div class="sg-nav-section-label">
              {{ section.title }}
            </div>
            <div class="flex-col" style="display:flex;flex-direction:column;gap:2px">
              @for (item of section.items; track item.route) {
                <a
                  [routerLink]="item.route"
                  routerLinkActive="active"
                  [routerLinkActiveOptions]="{ exact: item.route === '/' }"
                  class="sg-nav-item"
                >
                  <div style="display:flex;align-items:center;gap:10px">
                    <span [innerHTML]="item.icon" style="display:flex;align-items:center;width:16px;height:16px;color:currentColor"></span>
                    <span>{{ item.label }}</span>
                  </div>
                  @if (item.badge) {
                    <span class="sg-nav-badge">{{ item.badge }}</span>
                  }
                </a>
              }
            </div>
          </div>
        }
      </nav>
    </aside>
  `,
})
export class SidebarComponent {
  sections: NavSection[] = [
    {
      title: 'OVERVIEW',
      items: [
        {
          label: 'Overview',
          route: '/',
          icon: '<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><rect width="7" height="9" x="3" y="3" rx="1"/><rect width="7" height="5" x="14" y="3" rx="1"/><rect width="7" height="9" x="14" y="12" rx="1"/><rect width="7" height="5" x="3" y="16" rx="1"/></svg>',
        },
        {
          label: 'Gateways',
          route: '/gateways',
          icon: '<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="m18 8 4 4-4 4"/><path d="M2 12h20"/><path d="m6 16-4-4 4-4"/></svg>',
        },
        {
          label: 'Nodes',
          route: '/nodes',
          icon: '<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><rect width="20" height="8" x="2" y="2" rx="2"/><rect width="20" height="8" x="2" y="14" rx="2"/><line x1="6" x2="6.01" y1="6" y2="6"/><line x1="6" x2="6.01" y1="18" y2="18"/></svg>',
        },
        {
          label: 'Tunnels',
          route: '/tunnels',
          icon: '<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M12 2v20"/><path d="M17 5H9.5a3.5 3.5 0 0 0 0 7h5a3.5 3.5 0 0 1 0 7H6"/></svg>',
        },
        {
          label: 'Flows',
          route: '/flows',
          icon: '<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><polyline points="22 12 18 12 15 21 9 3 6 12 2 12"/></svg>',
        },
        {
          label: 'Topology',
          route: '/topology',
          icon: '<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><circle cx="18" cy="5" r="3"/><circle cx="6" cy="12" r="3"/><circle cx="18" cy="19" r="3"/><line x1="8.59" x2="15.42" y1="13.51" y2="17.49"/><line x1="15.41" x2="8.59" y1="6.51" y2="10.49"/></svg>',
        },
      ],
    },
    {
      title: 'INFRASTRUCTURE',
      items: [
        {
          label: 'Services',
          route: '/services',
          icon: '<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><ellipse cx="12" cy="5" rx="9" ry="3"/><path d="M3 5v14c0 1.66 4 3 9 3s9-1.34 9-3V5"/><path d="M3 12c0 1.66 4 3 9 3s9-1.34 9-3"/></svg>',
        },
        {
          label: 'Endpoints',
          route: '/endpoints',
          icon: '<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><circle cx="12" cy="12" r="10"/><circle cx="12" cy="12" r="2"/></svg>',
        },
      ],
    },
    {
      title: 'NETWORK',
      items: [
        {
          label: 'eBPF',
          route: '/ebpf',
          icon: '<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><polygon points="13 2 3 14 12 14 11 22 21 10 12 10 13 2"/></svg>',
        },
        {
          label: 'CNI',
          route: '/cni',
          icon: '<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><rect width="18" height="18" x="3" y="3" rx="2"/><path d="M3 9h18"/><path d="M9 21V9"/></svg>',
        },
        {
          label: 'Packets',
          route: '/packets',
          icon: '<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><rect width="20" height="14" x="2" y="5" rx="2"/><line x1="2" x2="22" y1="10" y2="10"/></svg>',
        },
      ],
    },
    {
      title: 'TELEMETRY',
      items: [
        {
          label: 'Events',
          route: '/events',
          icon: '<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M18 8A6 6 0 0 0 6 8c0 7-3 9-3 9h18s-3-2-3-9"/><path d="M13.73 21a2 2 0 0 1-3.46 0"/></svg>',
        },
        {
          label: 'Logs',
          route: '/logs',
          icon: '<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><line x1="4" x2="20" y1="6" y2="6"/><line x1="4" x2="20" y1="12" y2="12"/><line x1="4" x2="20" y1="18" y2="18"/></svg>',
        },
        {
          label: 'Metrics',
          route: '/metrics',
          icon: '<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M3 3v18h18"/><path d="m19 9-5 5-4-4-3 3"/></svg>',
        },
        {
          label: 'Traces',
          route: '/traces',
          icon: '<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><circle cx="12" cy="12" r="10"/><path d="M12 6v6l4 2"/></svg>',
        },
      ],
    },
    {
      title: 'SYSTEM',
      items: [
        {
          label: 'Settings',
          route: '/settings',
          icon: '<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><circle cx="12" cy="12" r="3"/><path d="M19.4 15a1.65 1.65 0 0 0 .33 1.82l.06.06a2 2 0 0 1 0 2.83 2 2 0 0 1-2.83 0l-.06-.06a1.65 1.65 0 0 0-1.82-.33 1.65 1.65 0 0 0-1 1.51V21a2 2 0 0 1-2 2 2 2 0 0 1-2-2v-.09A1.65 1.65 0 0 0 9 19.4a1.65 1.65 0 0 0-1.82.33l-.06.06a2 2 0 0 1-2.83 0 2 2 0 0 1 0-2.83l.06-.06a1.65 1.65 0 0 0 .33-1.82 1.65 1.65 0 0 0-1.51-1H3a2 2 0 0 1-2-2 2 2 0 0 1 2-2h.09A1.65 1.65 0 0 0 4.6 9a1.65 1.65 0 0 0-.33-1.82l-.06-.06a2 2 0 0 1 0-2.83 2 2 0 0 1 2.83 0l.06.06a1.65 1.65 0 0 0 1.82.33H9a1.65 1.65 0 0 0 1-1.51V3a2 2 0 0 1 2-2 2 2 0 0 1 2 2v.09a1.65 1.65 0 0 0 1 1.51 1.65 1.65 0 0 0 1.82-.33l.06-.06a2 2 0 0 1 2.83 0 2 2 0 0 1 0 2.83l-.06.06a1.65 1.65 0 0 0-.33 1.82V9a1.65 1.65 0 0 0 1.51 1H21a2 2 0 0 1 2 2 2 2 0 0 1-2 2h-.09a1.65 1.65 0 0 0-1.51 1z"/></svg>',
        },
      ],
    },
  ];
}
