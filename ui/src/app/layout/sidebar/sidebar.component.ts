import { Component, input, computed } from '@angular/core';
import { RouterLink, RouterLinkActive } from '@angular/router';

interface NavItem {
  readonly path: string;
  readonly label: string;
  readonly icon: string;
  readonly ariaLabel: string;
}

const NAV_ITEMS: NavItem[] = [
  { path: '/dashboard', label: 'Dashboard',  icon: 'grid',      ariaLabel: 'Dashboard' },
  { path: '/gateways',  label: 'Gateways',   icon: 'gateway',   ariaLabel: 'Gateways' },
  { path: '/nodes',     label: 'Nodes',      icon: 'server',    ariaLabel: 'Nodes' },
  { path: '/tunnels',   label: 'Tunnels',    icon: 'tunnel',    ariaLabel: 'Tunnels' },
  { path: '/flows',     label: 'Flows',      icon: 'flow',      ariaLabel: 'Traffic Flows' },
  { path: '/packets',   label: 'Packets',    icon: 'packet',    ariaLabel: 'Packet Capture' },
  { path: '/topology',  label: 'Topology',   icon: 'topology',  ariaLabel: 'Network Topology' },
  { path: '/services',  label: 'Services',   icon: 'service',   ariaLabel: 'Kubernetes Services' },
  { path: '/endpoints', label: 'Endpoints',  icon: 'endpoint',  ariaLabel: 'Endpoints' },
  { path: '/ebpf',      label: 'eBPF',       icon: 'ebpf',      ariaLabel: 'eBPF Programs' },
  { path: '/cni',       label: 'CNI',        icon: 'cni',       ariaLabel: 'CNI Configuration' },
  { path: '/events',    label: 'Events',     icon: 'event',     ariaLabel: 'Events' },
  { path: '/logs',      label: 'Logs',       icon: 'log',       ariaLabel: 'Log Viewer' },
  { path: '/metrics',   label: 'Metrics',    icon: 'metric',    ariaLabel: 'Metrics' },
  { path: '/traces',    label: 'Traces',     icon: 'trace',     ariaLabel: 'Distributed Traces' },
  { path: '/settings',  label: 'Settings',   icon: 'settings',  ariaLabel: 'Settings' },
];

@Component({
  selector: 'sg-sidebar',
  imports: [RouterLink, RouterLinkActive],
  template: `
    <nav
      id="sg-sidebar"
      class="sg-sidebar"
      [class.sg-sidebar-collapsed]="collapsed()"
      aria-label="Main navigation"
      role="navigation"
    >
      <ul class="sg-nav-list" role="list">
        @for (item of navItems; track item.path) {
          <li role="listitem">
            <a
              [routerLink]="item.path"
              routerLinkActive="sg-nav-active"
              class="sg-nav-item"
              [attr.aria-label]="collapsed() ? item.ariaLabel : null"
              [title]="collapsed() ? item.label : ''"
            >
              <span class="sg-nav-icon" aria-hidden="true">
                @switch (item.icon) {
                  @case ('grid') {
                    <svg width="18" height="18" viewBox="0 0 18 18" fill="none"><rect x="2" y="2" width="6" height="6" rx="1" fill="currentColor" opacity=".7"/><rect x="10" y="2" width="6" height="6" rx="1" fill="currentColor" opacity=".7"/><rect x="2" y="10" width="6" height="6" rx="1" fill="currentColor" opacity=".7"/><rect x="10" y="10" width="6" height="6" rx="1" fill="currentColor" opacity=".7"/></svg>
                  }
                  @default {
                    <svg width="18" height="18" viewBox="0 0 18 18" fill="none"><circle cx="9" cy="9" r="6" stroke="currentColor" stroke-width="1.5" opacity=".7"/></svg>
                  }
                }
              </span>
              @if (!collapsed()) {
                <span class="sg-nav-label">{{ item.label }}</span>
              }
            </a>
          </li>
        }
      </ul>
    </nav>
  `,
  styles: [`
    .sg-sidebar {
      grid-column: 1;
      background: var(--sg-bg-surface);
      border-right: 1px solid var(--sg-border);
      overflow-y: auto;
      overflow-x: hidden;
      display: flex;
      flex-direction: column;
      transition: width var(--sg-transition);
    }

    .sg-nav-list {
      list-style: none;
      padding: 8px 0;
      margin: 0;
      display: flex;
      flex-direction: column;
      gap: 2px;
    }

    .sg-nav-item {
      display: flex;
      align-items: center;
      gap: 10px;
      padding: 8px 12px;
      margin: 0 6px;
      border-radius: var(--sg-radius);
      text-decoration: none;
      color: var(--sg-text-secondary);
      font-size: 0.8125rem;
      font-weight: 500;
      transition: all var(--sg-transition-fast);
      white-space: nowrap;
    }

    .sg-nav-item:hover {
      background: var(--sg-bg-hover);
      color: var(--sg-text-primary);
    }

    .sg-nav-item:focus-visible {
      outline: 2px solid var(--sg-accent);
      outline-offset: -2px;
    }

    .sg-nav-active {
      background: var(--sg-accent-glow);
      color: var(--sg-accent) !important;
      border: 1px solid rgba(56, 189, 248, 0.2);
    }

    .sg-nav-icon {
      display: flex;
      align-items: center;
      justify-content: center;
      flex-shrink: 0;
      width: 20px;
    }

    .sg-nav-label {
      overflow: hidden;
      text-overflow: ellipsis;
    }

    .sg-sidebar-collapsed .sg-nav-item {
      justify-content: center;
      padding: 10px;
      margin: 0 8px;
    }
  `],
})
export class SidebarComponent {
  readonly collapsed = input<boolean>(false);
  readonly navItems = NAV_ITEMS;
}
