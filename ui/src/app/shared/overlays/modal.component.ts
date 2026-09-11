// Copyright 2026 straitgateway Authors
// SPDX-License-Identifier: Apache-2.0

import { Component, input, output, HostListener } from '@angular/core';
import { CommonModule } from '@angular/common';

@Component({
  selector: 'sg-modal',
  standalone: true,
  imports: [CommonModule],
  template: `
    <div class="sg-modal-backdrop" (click)="onBackdropClick($event)">
      <div class="sg-modal-panel" [style.max-width]="maxWidth()" role="dialog" aria-modal="true" [attr.aria-label]="title()">
        <!-- Header -->
        <div class="sg-modal-header">
          <div>
            <h2 class="sg-modal-title">{{ title() }}</h2>
            @if (subtitle()) {
              <p class="sg-modal-subtitle">{{ subtitle() }}</p>
            }
          </div>
          <button class="sg-modal-close" (click)="close.emit()" aria-label="Close modal">
            <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
              <line x1="18" y1="6" x2="6" y2="18" />
              <line x1="6" y1="6" x2="18" y2="18" />
            </svg>
          </button>
        </div>

        <!-- Body -->
        <div class="sg-modal-body">
          <ng-content />
        </div>

        <!-- Optional Footer slot -->
        <div class="sg-modal-footer">
          <ng-content select="[slot=footer]" />
        </div>
      </div>
    </div>
  `,
  styles: [`
    .sg-modal-backdrop {
      position: fixed;
      inset: 0;
      background: rgba(4, 6, 12, 0.75);
      backdrop-filter: blur(8px);
      -webkit-backdrop-filter: blur(8px);
      display: flex;
      align-items: center;
      justify-content: center;
      z-index: 1050;
      padding: 20px;
      animation: fadeIn 150ms cubic-bezier(0.4, 0, 0.2, 1);
    }
    .sg-modal-panel {
      width: 100%;
      background: rgba(18, 22, 33, 0.95);
      border: 1px solid var(--sg-border-hover);
      border-radius: var(--sg-radius);
      box-shadow: 0 20px 50px rgba(0, 0, 0, 0.6), 0 0 30px rgba(99, 102, 241, 0.15);
      display: flex;
      flex-direction: column;
      max-height: 90vh;
      overflow: hidden;
      animation: scaleIn 180ms cubic-bezier(0.4, 0, 0.2, 1);
    }
    .sg-modal-header {
      padding: 18px 24px;
      display: flex;
      align-items: flex-start;
      justify-content: space-between;
      border-bottom: 1px solid var(--sg-border);
    }
    .sg-modal-title {
      font-size: 17px;
      font-weight: 600;
      color: var(--sg-text);
      margin: 0;
    }
    .sg-modal-subtitle {
      font-size: 12px;
      color: var(--sg-text-2);
      margin-top: 3px;
    }
    .sg-modal-close {
      color: var(--sg-text-3);
      padding: 4px;
      border-radius: var(--sg-radius-sm);
      display: flex;
      align-items: center;
      justify-content: center;
      transition: color 150ms ease, background 150ms ease;
    }
    .sg-modal-close:hover {
      color: var(--sg-text);
      background: var(--sg-surface-3);
    }
    .sg-modal-body {
      padding: 20px 24px;
      overflow-y: auto;
      flex: 1;
    }
    .sg-modal-footer {
      padding: 14px 24px;
      border-top: 1px solid var(--sg-border);
      display: flex;
      justify-content: flex-end;
      gap: 12px;
      background: rgba(13, 16, 24, 0.5);
    }
    @keyframes fadeIn {
      from { opacity: 0; }
      to { opacity: 1; }
    }
    @keyframes scaleIn {
      from { opacity: 0; transform: scale(0.96) translateY(6px); }
      to { opacity: 1; transform: scale(1) translateY(0); }
    }
  `],
})
export class ModalComponent {
  title = input.required<string>();
  subtitle = input<string>('');
  maxWidth = input<string>('580px');
  close = output<void>();

  onBackdropClick(event: MouseEvent) {
    if ((event.target as HTMLElement).classList.contains('sg-modal-backdrop')) {
      this.close.emit();
    }
  }

  @HostListener('window:keydown.escape')
  onEscape() {
    this.close.emit();
  }
}
