// Copyright 2026 straitgateway Authors
// SPDX-License-Identifier: Apache-2.0

import { Component, input, computed } from '@angular/core';
import { CommonModule } from '@angular/common';

@Component({
  selector: 'sg-gauge',
  standalone: true,
  imports: [CommonModule],
  template: `
    <div class="sg-gauge-container" [style.width]="size() + 'px'" [style.height]="size() + 'px'">
      <svg [attr.viewBox]="'0 0 ' + size() + ' ' + size()" class="sg-gauge-svg">
        <!-- Background Track -->
        <circle
          [attr.cx]="center()"
          [attr.cy]="center()"
          [attr.r]="radius()"
          fill="none"
          stroke="var(--sg-surface-3)"
          [attr.stroke-width]="strokeWidth()"
        />
        <!-- Progress Arc -->
        <circle
          [attr.cx]="center()"
          [attr.cy]="center()"
          [attr.r]="radius()"
          fill="none"
          [attr.stroke]="gaugeColor()"
          [attr.stroke-width]="strokeWidth()"
          stroke-linecap="round"
          [attr.stroke-dasharray]="circumference()"
          [attr.stroke-dashoffset]="dashOffset()"
          transform-origin="center"
          style="transform: rotate(-90deg); transition: stroke-dashoffset 0.6s cubic-bezier(0.4, 0, 0.2, 1), stroke 0.3s ease;"
        />
      </svg>
      <div class="sg-gauge-content">
        <span class="sg-gauge-value" [style.color]="gaugeColor()">
          {{ value() | number:'1.0-1' }}%
        </span>
        @if (label()) {
          <span class="sg-gauge-label">{{ label() }}</span>
        }
      </div>
    </div>
  `,
  styles: [`
    .sg-gauge-container {
      position: relative;
      display: inline-flex;
      align-items: center;
      justify-content: center;
    }
    .sg-gauge-svg {
      width: 100%;
      height: 100%;
    }
    .sg-gauge-content {
      position: absolute;
      display: flex;
      flex-direction: column;
      align-items: center;
      justify-content: center;
      text-align: center;
      pointer-events: none;
    }
    .sg-gauge-value {
      font-size: 16px;
      font-weight: 700;
      line-height: 1.2;
    }
    .sg-gauge-label {
      font-size: 10px;
      color: var(--sg-text-3);
      text-transform: uppercase;
      letter-spacing: 0.05em;
    }
  `],
})
export class GaugeComponent {
  value = input<number>(0);
  label = input<string>('');
  size = input<number>(90);
  strokeWidth = input<number>(7);

  center = computed(() => this.size() / 2);
  radius = computed(() => (this.size() - this.strokeWidth()) / 2);
  circumference = computed(() => 2 * Math.PI * this.radius());

  dashOffset = computed(() => {
    const val = Math.min(100, Math.max(0, this.value()));
    const progress = val / 100;
    return this.circumference() * (1 - progress);
  });

  gaugeColor = computed(() => {
    const val = this.value();
    if (val >= 90) return 'var(--sg-success)';
    if (val >= 70) return 'var(--sg-accent)';
    if (val >= 50) return 'var(--sg-warn)';
    return 'var(--sg-danger)';
  });
}
