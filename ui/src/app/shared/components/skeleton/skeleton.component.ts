import { Component, input } from '@angular/core';

export type SkeletonVariant = 'line' | 'block' | 'circle';

@Component({
  selector: 'sg-skeleton',
  template: `
    <div
      class="sg-skeleton"
      [class]="variantClass()"
      [style.width]="width()"
      [style.height]="height()"
      role="status"
      aria-busy="true"
      [attr.aria-label]="label()"
    ></div>
  `,
  styles: [`
    :host { display: block; }
    .sg-skeleton-line   { height: 14px; border-radius: 4px; }
    .sg-skeleton-block  { border-radius: var(--sg-radius); }
    .sg-skeleton-circle { border-radius: 50%; }
  `],
})
export class SkeletonComponent {
  readonly variant = input<SkeletonVariant>('line');
  readonly width    = input<string>('100%');
  readonly height   = input<string | undefined>(undefined);
  readonly label    = input<string>('Loading…');

  variantClass(): string {
    return `sg-skeleton-${this.variant()}`;
  }
}

/** Renders N skeleton lines to fill a loading content area */
@Component({
  selector: 'sg-skeleton-list',
  imports: [SkeletonComponent],
  template: `
    <div class="sg-skeleton-list" [attr.aria-label]="label()" role="status" aria-busy="true">
      @for (_ of rows(); track $index) {
        <sg-skeleton variant="line" [width]="$index % 3 === 2 ? '70%' : '100%'" />
      }
    </div>
  `,
  styles: [`.sg-skeleton-list { display: flex; flex-direction: column; gap: 10px; padding: 1rem 0; }`],
})
export class SkeletonListComponent {
  readonly count = input<number>(5);
  readonly label = input<string>('Loading content…');
  rows(): number[] { return Array.from({ length: this.count() }, (_, i) => i); }
}
