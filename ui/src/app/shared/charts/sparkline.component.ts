// Copyright 2026 straitgateway Authors
// SPDX-License-Identifier: Apache-2.0

import { Component, input, computed } from '@angular/core';
import { CommonModule } from '@angular/common';

@Component({
  selector: 'sg-sparkline',
  standalone: true,
  imports: [CommonModule],
  template: `
    <svg [attr.viewBox]="'0 0 ' + width() + ' ' + height()" [style.width]="width() + 'px'" [style.height]="height() + 'px'" class="sg-sparkline">
      <defs>
        <linearGradient [id]="gradId()" x1="0%" y1="0%" x2="0%" y2="100%">
          <stop offset="0%" [attr.stop-color]="color()" stop-opacity="0.35" />
          <stop offset="100%" [attr.stop-color]="color()" stop-opacity="0.0" />
        </linearGradient>
      </defs>
      @if (areaPath()) {
        <path [attr.d]="areaPath()" [attr.fill]="'url(#' + gradId() + ')'" />
      }
      @if (linePath()) {
        <path [attr.d]="linePath()" fill="none" [attr.stroke]="color()" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" />
      }
    </svg>
  `,
  styles: [`
    .sg-sparkline {
      display: block;
      overflow: visible;
    }
  `],
})
export class SparklineComponent {
  data = input<number[]>([]);
  width = input<number>(120);
  height = input<number>(36);
  color = input<string>('var(--sg-accent)');

  gradId = computed(() => 'sg-grad-' + Math.random().toString(36).substring(2, 9));

  linePath = computed(() => {
    const pts = this.data();
    if (!pts || pts.length < 2) return '';
    const w = this.width();
    const h = this.height();
    const min = Math.min(...pts);
    const max = Math.max(...pts);
    const range = max - min || 1;
    const step = w / (pts.length - 1);

    return pts
      .map((val, idx) => {
        const x = idx * step;
        const y = h - ((val - min) / range) * (h - 6) - 3;
        return `${idx === 0 ? 'M' : 'L'} ${x.toFixed(1)} ${y.toFixed(1)}`;
      })
      .join(' ');
  });

  areaPath = computed(() => {
    const lPath = this.linePath();
    if (!lPath) return '';
    const w = this.width();
    const h = this.height();
    return `${lPath} L ${w} ${h} L 0 ${h} Z`;
  });
}
