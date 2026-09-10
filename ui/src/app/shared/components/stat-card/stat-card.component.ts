// Copyright 2026 straitgateway Authors — SPDX-License-Identifier: Apache-2.0
import { Component, input } from '@angular/core';
import { CommonModule } from '@angular/common';

@Component({
  selector: 'sg-stat-card',
  standalone: true,
  imports: [CommonModule],
  template: `
<div class="sg-stat-card">
  <div class="sg-stat-label">{{ label() }}</div>
  <div class="sg-stat-value" [style.color]="color()">{{ value() }}</div>
  @if (meta()) {
    <div class="sg-stat-meta">{{ meta() }}</div>
  }
</div>
`,
})
export class StatCardComponent {
  label = input.required<string>();
  value = input.required<string | number>();
  meta  = input<string>('');
  color = input<string>('var(--sg-text-1)');
}
