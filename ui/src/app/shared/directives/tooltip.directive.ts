// Copyright 2026 straitgateway Authors — SPDX-License-Identifier: Apache-2.0
import { Directive, ElementRef, HostListener, inject, input, Renderer2 } from '@angular/core';

/** Simple CSS tooltip directive using ::after pseudo-element. */
@Directive({ selector: '[sgTooltip]', standalone: true })
export class TooltipDirective {
  sgTooltip = input.required<string>();
  private el = inject(ElementRef);
  private renderer = inject(Renderer2);
  private tooltipEl: HTMLElement | null = null;

  @HostListener('mouseenter')
  onMouseEnter() {
    if (this.tooltipEl) return;
    this.tooltipEl = this.renderer.createElement('div');
    this.renderer.setStyle(this.tooltipEl, 'position', 'absolute');
    this.renderer.setStyle(this.tooltipEl, 'background', 'var(--sg-surface-3, #333)');
    this.renderer.setStyle(this.tooltipEl, 'color', 'var(--sg-text-1, #fff)');
    this.renderer.setStyle(this.tooltipEl, 'padding', '4px 8px');
    this.renderer.setStyle(this.tooltipEl, 'border-radius', '4px');
    this.renderer.setStyle(this.tooltipEl, 'font-size', '11px');
    this.renderer.setStyle(this.tooltipEl, 'white-space', 'nowrap');
    this.renderer.setStyle(this.tooltipEl, 'z-index', '9999');
    this.renderer.setStyle(this.tooltipEl, 'pointer-events', 'none');
    this.renderer.setStyle(this.tooltipEl, 'transform', 'translateX(-50%)');
    this.tooltipEl!.textContent = this.sgTooltip();

    const rect = this.el.nativeElement.getBoundingClientRect();
    this.renderer.setStyle(this.tooltipEl, 'top', `${rect.bottom + 6}px`);
    this.renderer.setStyle(this.tooltipEl, 'left', `${rect.left + rect.width / 2}px`);
    this.renderer.appendChild(document.body, this.tooltipEl);
  }

  @HostListener('mouseleave')
  onMouseLeave() {
    if (this.tooltipEl) {
      this.renderer.removeChild(document.body, this.tooltipEl);
      this.tooltipEl = null;
    }
  }
}
