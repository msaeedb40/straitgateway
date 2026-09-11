// Copyright 2026 straitgateway Authors
// SPDX-License-Identifier: Apache-2.0

import { Component, input, output, HostListener } from '@angular/core';
import { CommonModule } from '@angular/common';

@Component({
  selector: 'sg-drawer',
  standalone: true,
  imports: [CommonModule],
  template: `
    <div class="sg-drawer-backdrop" (click)="onBackdropClick($event)">
      <div class="sg-drawer-panel" [style.width]="width()" role="dialog" aria-modal="true">
        <div class="sg-drawer-header">
          <div>
            <h3 class="sg-drawer-title">{{ title() }}</h3>
            @if (subtitle()) {
              <p class="sg-drawer-subtitle">{{ subtitle() }}</p>
            }
          </div>
          <button class="sg-drawer-close" (click)="close.emit()" aria-label="Close drawer">✕</button>
        </div>
        <div class="sg-drawer-body">
          <ng-content />
        </div>
      </div>
    </div>
  `,
  styles: [`
    .sg-drawer-backdrop {
      position: fixed;
      inset: 0;
      background: rgba(4, 6, 12, 0.65);
      backdrop-filter: blur(4px);
      z-index: 1040;
      display: flex;
      justify-content: flex-end;
    }
    .sg-drawer-panel {
      height: 100%;
      background: rgba(15, 18, 28, 0.98);
      border-left: 1px solid var(--sg-border-hover);
      box-shadow: -10px 0 30px rgba(0, 0, 0, 0.5);
      display: flex;
      flex-direction: column;
      animation: slideInRight 200ms cubic-bezier(0.4, 0, 0.2, 1);
    }
    .sg-drawer-header {
      padding: 18px 24px;
      display: flex;
      align-items: center;
      justify-content: space-between;
      border-bottom: 1px solid var(--sg-border);
    }
    .sg-drawer-title {
      font-size: 16px;
      font-weight: 600;
      color: var(--sg-text);
      margin: 0;
    }
    .sg-drawer-subtitle {
      font-size: 12px;
      color: var(--sg-text-2);
      margin-top: 2px;
    }
    .sg-drawer-close {
      font-size: 16px;
      color: var(--sg-text-3);
      padding: 4px 8px;
      cursor: pointer;
    }
    .sg-drawer-close:hover {
      color: var(--sg-text);
    }
    .sg-drawer-body {
      padding: 20px 24px;
      overflow-y: auto;
      flex: 1;
    }
    @keyframes slideInRight {
      from { transform: translateX(100%); }
      to { transform: translateX(0); }
    }
  `],
})
export class DrawerComponent {
  title = input.required<string>();
  subtitle = input<string>('');
  width = input<string>('440px');
  close = output<void>();

  onBackdropClick(event: MouseEvent) {
    if ((event.target as HTMLElement).classList.contains('sg-drawer-backdrop')) {
      this.close.emit();
    }
  }

  @HostListener('window:keydown.escape')
  onEscape() {
    this.close.emit();
  }
}
