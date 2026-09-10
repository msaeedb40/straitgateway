// Copyright 2026 straitgateway Authors — SPDX-License-Identifier: Apache-2.0
import { Component, input } from '@angular/core';
import { CommonModule } from '@angular/common';

@Component({
  selector: 'sg-empty-state',
  standalone: true,
  imports: [CommonModule],
  template: `
<div class="sg-empty">
  <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5" width="40" height="40">
    <circle cx="12" cy="12" r="10"/>
    <line x1="12" y1="8" x2="12" y2="12"/>
    <line x1="12" y1="16" x2="12.01" y2="16"/>
  </svg>
  <p>{{ message() }}</p>
  @if (hint()) {
    <p class="text-muted text-sm">{{ hint() }}</p>
  }
</div>
`,
})
export class EmptyStateComponent {
  message = input<string>('No data available');
  hint    = input<string>('');
}
