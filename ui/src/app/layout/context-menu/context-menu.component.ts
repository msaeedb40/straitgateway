// Copyright 2026 straitgateway Authors
// SPDX-License-Identifier: Apache-2.0

import { Component, inject, HostListener } from '@angular/core';
import { CommonModule } from '@angular/common';
import { ContextMenuItem, ContextMenuService } from './context-menu.service';

@Component({
  selector: 'sg-context-menu',
  standalone: true,
  imports: [CommonModule],
  template: `
    @if (menu.state().isOpen) {
      <div
        class="sg-context-backdrop"
        (click)="menu.close()"
        (contextmenu)="$event.preventDefault(); menu.close()"
      >
        <div
          class="sg-context-menu"
          [style.left.px]="menu.state().x"
          [style.top.px]="menu.state().y"
          (click)="$event.stopPropagation()"
        >
          @for (item of menu.state().items; track item.label) {
            @if (item.divider) {
              <div class="sg-context-divider"></div>
            } @else {
              <button
                class="sg-context-item"
                [class.danger]="item.danger"
                [disabled]="item.disabled"
                (click)="trigger(item)"
              >
                @if (item.icon) {
                  <span class="sg-context-icon" [innerHTML]="item.icon"></span>
                }
                <span>{{ item.label }}</span>
              </button>
            }
          }
        </div>
      </div>
    }
  `,
  styles: [`
    .sg-context-backdrop {
      position: fixed;
      inset: 0;
      z-index: 1200;
      background: transparent;
    }
    .sg-context-menu {
      position: absolute;
      min-width: 180px;
      background: rgba(18, 22, 33, 0.96);
      backdrop-filter: blur(14px);
      -webkit-backdrop-filter: blur(14px);
      border: 1px solid var(--sg-border-hover);
      border-radius: var(--sg-radius-sm);
      box-shadow: 0 10px 30px rgba(0, 0, 0, 0.5), 0 0 20px rgba(99, 102, 241, 0.12);
      padding: 6px;
      display: flex;
      flex-direction: column;
      gap: 2px;
      animation: fadeIn 120ms ease;
    }
    .sg-context-item {
      display: flex;
      align-items: center;
      gap: 8px;
      width: 100%;
      padding: 7px 10px;
      font-size: 13px;
      font-weight: 500;
      color: var(--sg-text);
      border-radius: 4px;
      text-align: left;
      transition: background 120ms ease, color 120ms ease;
    }
    .sg-context-item:hover:not(:disabled) {
      background: var(--sg-surface-3);
      color: var(--sg-accent-light);
    }
    .sg-context-item.danger:hover:not(:disabled) {
      background: rgba(239, 68, 68, 0.15);
      color: var(--sg-danger);
    }
    .sg-context-item:disabled {
      opacity: 0.4;
      cursor: not-allowed;
    }
    .sg-context-divider {
      height: 1px;
      background: var(--sg-border);
      margin: 4px 0;
    }
    .sg-context-icon {
      display: inline-flex;
      align-items: center;
      justify-content: center;
      width: 16px;
      height: 16px;
    }
    @keyframes fadeIn {
      from { opacity: 0; transform: scale(0.97); }
      to { opacity: 1; transform: scale(1); }
    }
  `],
})
export class ContextMenuComponent {
  readonly menu = inject(ContextMenuService);

  trigger(item: ContextMenuItem): void {
    if (item.disabled) return;
    const data = this.menu.state().data;
    this.menu.close();
    item.action(data);
  }

  @HostListener('window:keydown.escape')
  onEscape(): void {
    this.menu.close();
  }
}
