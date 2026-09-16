import { Component, input, inject, ElementRef, AfterViewInit, OnChanges } from '@angular/core';
import { DecimalPipe } from '@angular/common';

@Component({
  selector: 'sg-gauge',
  imports: [DecimalPipe],
  template: `
    <figure class="sg-gauge-figure" [attr.aria-label]="label() + ': ' + value() + '%'">
      <svg #gaugeSvg class="sg-gauge-svg" role="img" [attr.aria-label]="label() + ' gauge'"></svg>
      <figcaption class="sg-gauge-caption">
        <span class="sg-gauge-value">{{ value() | number:'1.0-1' }}{{ unit() }}</span>
        <span class="sg-gauge-label">{{ label() }}</span>
      </figcaption>
    </figure>
  `,
  styles: [`
    :host { display: block; }
    .sg-gauge-figure { margin: 0; text-align: center; }
    .sg-gauge-svg { display: block; margin: 0 auto; }
    .sg-gauge-caption { display: flex; flex-direction: column; align-items: center; margin-top: -8px; }
    .sg-gauge-value { font-size: 1.25rem; font-weight: 700; color: var(--sg-text-primary); font-family: var(--sg-font-mono); }
    .sg-gauge-label { font-size: 0.75rem; color: var(--sg-text-muted); }
  `],
})
export class GaugeComponent implements AfterViewInit, OnChanges {
  readonly value    = input<number>(0);   // 0–100
  readonly label    = input<string>('');
  readonly unit     = input<string>('%');
  readonly size     = input<number>(100);
  readonly color    = input<string>('var(--sg-accent)');

  private readonly el = inject(ElementRef);
  private d3Module: typeof import('d3') | null = null;

  async ngAfterViewInit(): Promise<void> {
    this.d3Module = await import('d3');
    this.render();
  }

  ngOnChanges(): void {
    if (this.d3Module) this.render();
  }

  private render(): void {
    const d3 = this.d3Module!;
    const svg = this.el.nativeElement.querySelector('svg');
    const S = this.size();
    const R = S / 2 - 8;
    const τ = 2 * Math.PI;
    const startAngle = -τ * 0.75;
    const endAngle   =  τ * 0.25;

    d3.select(svg).attr('width', S).attr('height', S * 0.65).selectAll('*').remove();
    const g = d3.select(svg).append('g').attr('transform', `translate(${S / 2},${S / 2})`);

    const arc = d3.arc().innerRadius(R - 10).outerRadius(R).startAngle(startAngle);

    // Background track
    g.append('path')
      .datum({ endAngle })
      .attr('d', arc as unknown as string)
      .attr('fill', 'var(--sg-bg-overlay)');

    // Value arc
    const valueAngle = startAngle + (endAngle - startAngle) * (Math.min(this.value(), 100) / 100);
    g.append('path')
      .datum({ endAngle: valueAngle })
      .attr('d', arc as unknown as string)
      .attr('fill', this.color());
  }
}
