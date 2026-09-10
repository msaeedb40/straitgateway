// Copyright 2026 straitgateway Authors — SPDX-License-Identifier: Apache-2.0
import { Component, input, output } from '@angular/core';
import { CommonModule } from '@angular/common';

@Component({
  selector: 'sg-confirm-dialog',
  standalone: true,
  imports: [CommonModule],
  template: `
@if (open()) {
  <div class="sg-dialog-overlay" (click)="cancel.emit()">
    <div class="sg-dialog" (click)="$event.stopPropagation()">
      <div class="sg-dialog-header">{{ title() }}</div>
      <div class="sg-dialog-body">{{ message() }}</div>
      <div class="sg-dialog-actions">
        <button class="sg-btn sg-btn-secondary" (click)="cancel.emit()">Cancel</button>
        <button class="sg-btn" [class]="destructive() ? 'sg-btn-danger' : 'sg-btn-primary'" (click)="confirm.emit()">
          {{ confirmLabel() }}
        </button>
      </div>
    </div>
  </div>
}
`,
  styles: [`
    .sg-dialog-overlay {
      position: fixed; inset: 0; z-index: 1000;
      background: rgba(0,0,0,0.6); backdrop-filter: blur(4px);
      display: flex; align-items: center; justify-content: center;
    }
    .sg-dialog {
      background: var(--sg-surface-2); border: 1px solid var(--sg-border);
      border-radius: var(--sg-radius-lg); padding: 24px; min-width: 360px;
      box-shadow: 0 20px 40px rgba(0,0,0,0.3);
    }
    .sg-dialog-header { font-size: 16px; font-weight: 600; margin-bottom: 12px; color: var(--sg-text-1); }
    .sg-dialog-body { font-size: 13px; color: var(--sg-text-2); margin-bottom: 20px; line-height: 1.5; }
    .sg-dialog-actions { display: flex; gap: 8px; justify-content: flex-end; }
    .sg-btn-danger { background: var(--sg-danger); color: white; }
    .sg-btn-primary { background: var(--sg-accent); color: white; }
  `],
})
export class ConfirmDialogComponent {
  open         = input<boolean>(false);
  title        = input<string>('Confirm');
  message      = input<string>('Are you sure?');
  confirmLabel = input<string>('Confirm');
  destructive  = input<boolean>(false);
  confirm = output<void>();
  cancel  = output<void>();
}
