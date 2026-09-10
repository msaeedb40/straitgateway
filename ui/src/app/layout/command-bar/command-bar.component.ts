// Copyright 2026 straitgateway Authors
// SPDX-License-Identifier: Apache-2.0

import { Component, inject, signal, output } from '@angular/core';
import { CommonModule } from '@angular/common';
import { Router } from '@angular/router';
import { FormsModule } from '@angular/forms';
import { NamespaceService } from '../../core/services/namespace.service';
import { NotificationService } from '../../core/services/notification.service';

interface CommandItem {
  category: string;
  label: string;
  description?: string;
  action: () => void;
}

@Component({
  selector: 'sg-command-bar',
  standalone: true,
  imports: [CommonModule, FormsModule],
  template: `
    <div class="sg-modal-backdrop" (click)="close.emit()" style="position:fixed;inset:0;background:rgba(0,0,0,.65);backdrop-filter:blur(4px);z-index:999;display:flex;align-items:flex-start;justify-content:center;padding-top:12vh">
      <div class="sg-command-dialog" (click)="$event.stopPropagation()" style="width:100%;max-width:580px;background:var(--sg-surface);border:1px solid var(--sg-border);border-radius:var(--sg-radius);box-shadow:0 20px 40px rgba(0,0,0,.6);overflow:hidden">
        <div style="display:flex;align-items:center;padding:12px 16px;border-bottom:1px solid var(--sg-border);gap:10px">
          <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="var(--sg-text-3)" stroke-width="2">
            <circle cx="11" cy="11" r="8"/><line x1="21" y1="21" x2="16.65" y2="16.65"/>
          </svg>
          <input
            #inputElem
            type="text"
            [(ngModel)]="query"
            placeholder="Type a command, page name, or resource..."
            style="flex:1;background:transparent;border:none;outline:none;font-size:15px;color:var(--sg-text);font-family:inherit"
            (keydown.escape)="close.emit()"
          />
          <kbd style="font-size:10px;padding:2px 6px;background:var(--sg-surface-2);border:1px solid var(--sg-border);border-radius:4px;color:var(--sg-text-3)">ESC</kbd>
        </div>

        <div style="max-height:360px;overflow-y:auto;padding:8px">
          @for (item of filteredCommands(); track item.label) {
            <div
              (click)="runCommand(item)"
              style="display:flex;align-items:center;justify-content:space-between;padding:8px 12px;border-radius:var(--sg-radius-sm);cursor:pointer;transition:background 120ms;color:var(--sg-text)"
              class="sg-command-item hover:bg-[var(--sg-surface-2)]"
            >
              <div>
                <div style="font-size:13px;font-weight:500">{{ item.label }}</div>
                @if (item.description) {
                  <div style="font-size:11px;color:var(--sg-text-3)">{{ item.description }}</div>
                }
              </div>
              <span class="sg-badge" style="font-size:10px">{{ item.category }}</span>
            </div>
          } @empty {
            <div style="padding:24px;text-align:center;color:var(--sg-text-3);font-size:13px">
              No matching commands or resources found.
            </div>
          }
        </div>
      </div>
    </div>
  `,
})
export class CommandBarComponent {
  private router = inject(Router);
  private nsService = inject(NamespaceService);
  private notif = inject(NotificationService);

  close = output<void>();
  query = signal<string>('');

  get commands(): CommandItem[] {
    return [
      {
        category: 'Navigation',
        label: 'Dashboard Overview',
        description: 'Cluster health and live topology',
        action: () => this.router.navigate(['/']),
      },
      {
        category: 'Navigation',
        label: 'Gateways',
        description: 'Manage Gateway API gateways and listeners',
        action: () => this.router.navigate(['/gateways']),
      },
      {
        category: 'Navigation',
        label: 'Topology Visualizer',
        description: 'Interactive cluster network topology & traffic flow',
        action: () => this.router.navigate(['/topology']),
      },
      {
        category: 'Navigation',
        label: 'Flows',
        description: 'Live eBPF ring buffer flow inspector',
        action: () => this.router.navigate(['/flows']),
      },
      {
        category: 'Navigation',
        label: 'Nodes',
        description: 'Node agents, CNI status, and eBPF revisions',
        action: () => this.router.navigate(['/nodes']),
      },
      {
        category: 'Navigation',
        label: 'Transit Tunnels',
        description: 'Multi-cluster WireGuard transit gateway tunnels',
        action: () => this.router.navigate(['/tunnels']),
      },
      {
        category: 'Navigation',
        label: 'Services & Load Balancing',
        description: 'Kubernetes services and Maglev lookup tables',
        action: () => this.router.navigate(['/services']),
      },
      {
        category: 'Navigation',
        label: 'Settings',
        description: 'Configuration, eBPF maps, CNI, and Observability',
        action: () => this.router.navigate(['/settings']),
      },
      {
        category: 'Namespace',
        label: 'Switch Namespace to default',
        description: 'Filter resources for default namespace',
        action: () => this.nsService.select('default'),
      },
      {
        category: 'Namespace',
        label: 'Switch Namespace to straitgateway-system',
        description: 'Filter resources for straitgateway-system',
        action: () => this.nsService.select('straitgateway-system'),
      },
      {
        category: 'Namespace',
        label: 'Show All Namespaces',
        description: 'Disable namespace filtering',
        action: () => this.nsService.select(''),
      },
      {
        category: 'Action',
        label: 'Refresh System Health',
        description: 'Trigger health and latency check to controller',
        action: () => this.notif.info('Health Check', 'Triggered background controller ping'),
      },
    ];
  }

  filteredCommands(): CommandItem[] {
    const q = this.query().trim().toLowerCase();
    if (!q) return this.commands;
    return this.commands.filter(
      (c) =>
        c.label.toLowerCase().includes(q) ||
        (c.description && c.description.toLowerCase().includes(q)) ||
        c.category.toLowerCase().includes(q)
    );
  }

  runCommand(item: CommandItem) {
    item.action();
    this.close.emit();
  }
}
