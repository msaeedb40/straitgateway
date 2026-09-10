// Copyright 2026 straitgateway Authors — SPDX-License-Identifier: Apache-2.0
import { Component, input } from '@angular/core';
import { CommonModule } from '@angular/common';

@Component({
  selector: 'sg-badge',
  standalone: true,
  imports: [CommonModule],
  template: `<span class="sg-badge" [class]="variant()"><ng-content/></span>`,
})
export class BadgeComponent {
  variant = input<'active' | 'warn' | 'error' | 'pending' | 'info'>('active');
}
