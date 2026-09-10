// Copyright 2026 straitgateway Authors — SPDX-License-Identifier: Apache-2.0
import { Component, input } from '@angular/core';
import { CommonModule } from '@angular/common';

@Component({
  selector: 'sg-loading-skeleton',
  standalone: true,
  imports: [CommonModule],
  template: `
<div class="sg-skeleton-grid">
  @for (i of rows(); track i) {
    <div class="sg-skeleton" [style.height]="height()" [style.width]="width()"></div>
  }
</div>
`,
  styles: [`
    .sg-skeleton-grid { display: flex; flex-direction: column; gap: 8px; }
  `],
})
export class LoadingSkeletonComponent {
  rows   = input<number[]>([1, 2, 3]);
  height = input<string>('14px');
  width  = input<string>('100%');
}
