// Copyright 2026 straitgateway Authors
// SPDX-License-Identifier: Apache-2.0

import { Component, input } from '@angular/core';
import { CommonModule } from '@angular/common';

export type StatusVariant = 'healthy' | 'active' | 'degraded' | 'warn' | 'inactive' | 'pending' | 'critical';

@Component({
  selector: 'sg-status-badge',
  standalone: true,
  imports: [CommonModule],
  template: `
    <span class="sg-badge" [ngClass]="variantClass()">
      <span class="sg-dot" [ngClass]="variantClass()"></span>
      <span>{{ label() }}</span>
    </span>
  `,
  styles: [`
    .sg-dot {
      display: inline-block;
      width: 6px;
      height: 6px;
      border-radius: 50%;
      margin-right: 5px;
    }
    .sg-dot.active, .sg-dot.healthy {
      background: var(--sg-success);
      box-shadow: 0 0 6px var(--sg-success);
    }
    .sg-dot.warn, .sg-dot.degraded {
      background: var(--sg-warn);
      box-shadow: 0 0 6px var(--sg-warn);
    }
    .sg-dot.critical, .sg-dot.inactive {
      background: var(--sg-danger);
      box-shadow: 0 0 6px var(--sg-danger);
    }
    .sg-dot.pending {
      background: var(--sg-accent-light);
      box-shadow: 0 0 6px var(--sg-accent-light);
    }
  `],
})
export class StatusBadgeComponent {
  label = input.required<string>();
  variant = input<StatusVariant>('active');

  variantClass(): string {
    const v = this.variant();
    switch (v) {
      case 'healthy':
      case 'active':
        return 'active';
      case 'degraded':
      case 'warn':
        return 'warn';
      case 'critical':
      case 'inactive':
        return 'inactive';
      case 'pending':
        return 'pending';
      default:
        return 'active';
    }
  }
}
