import { Component, input, output, inject, ElementRef, AfterViewInit, OnDestroy, OnChanges } from '@angular/core';
import { MetricSeries } from '../../../core/models/metric.model';

@Component({
  selector: 'sg-line-chart',
  template: `
    <figure class="sg-chart-figure" [attr.aria-label]="label()">
      <figcaption class="sg-sr-only">{{ label() }}</figcaption>
      <svg #chartSvg class="sg-chart-svg" role="img" [attr.aria-label]="label()"></svg>
    </figure>
  `,
  styles: [`:host { display: block; width: 100%; } .sg-chart-figure { margin: 0; } .sg-chart-svg { width: 100%; display: block; }`],
})
export class LineChartComponent implements AfterViewInit, OnDestroy, OnChanges {
  readonly series   = input<MetricSeries[]>([]);
  readonly label    = input<string>('Line chart');
  readonly height   = input<number>(180);
  readonly unit     = input<string>('');
  readonly colors   = input<string[]>(['#38bdf8', '#2dd4bf', '#818cf8', '#34d399', '#fbbf24']);

  private readonly el = inject(ElementRef);
  private d3Module: typeof import('d3') | null = null;

  async ngAfterViewInit(): Promise<void> {
    this.d3Module = await import('d3');
    this.render();
  }

  ngOnChanges(): void {
    if (this.d3Module) this.render();
  }

  ngOnDestroy(): void {
    const svg = this.el.nativeElement.querySelector('svg');
    if (svg && this.d3Module) this.d3Module.select(svg).selectAll('*').remove();
  }

  private render(): void {
    const d3 = this.d3Module!;
    const svgEl = this.el.nativeElement.querySelector('svg') as SVGElement;
    const container = svgEl.parentElement!;
    const W = container.clientWidth || 400;
    const H = this.height();
    const margin = { top: 12, right: 20, bottom: 28, left: 48 };
    const innerW = W - margin.left - margin.right;
    const innerH = H - margin.top - margin.bottom;

    const root = d3.select(svgEl)
      .attr('width', W)
      .attr('height', H)
      .attr('viewBox', `0 0 ${W} ${H}`);
    root.selectAll('*').remove();

    const g = root.append('g').attr('transform', `translate(${margin.left},${margin.top})`);

    const allSamples = this.series().flatMap((s) => s.samples);
    if (!allSamples.length) {
      g.append('text')
        .attr('x', innerW / 2).attr('y', innerH / 2)
        .attr('text-anchor', 'middle')
        .attr('fill', 'var(--sg-text-muted)')
        .attr('font-size', '12')
        .text('No data');
      return;
    }

    const xScale = d3.scaleTime()
      .domain(d3.extent(allSamples, (d) => new Date(d.timestamp * 1000)) as [Date, Date])
      .range([0, innerW]);

    const yScale = d3.scaleLinear()
      .domain([0, d3.max(allSamples, (d) => d.value) ?? 1])
      .nice()
      .range([innerH, 0]);

    // Axes
    g.append('g').attr('transform', `translate(0,${innerH})`)
      .call(d3.axisBottom(xScale).ticks(4).tickSize(0))
      .call((ax) => ax.select('.domain').remove())
      .selectAll('text').attr('fill', 'var(--sg-text-muted)').attr('font-size', '10');

    g.append('g')
      .call(d3.axisLeft(yScale).ticks(4).tickSize(-innerW))
      .call((ax) => {
        ax.select('.domain').remove();
        ax.selectAll('.tick line').attr('stroke', 'var(--sg-border)');
        ax.selectAll('text').attr('fill', 'var(--sg-text-muted)').attr('font-size', '10');
      });

    // Lines
    const lineGen = d3.line<{ timestamp: number; value: number }>()
      .x((d) => xScale(new Date(d.timestamp * 1000)))
      .y((d) => yScale(d.value))
      .curve(d3.curveMonotoneX);

    this.series().forEach((s, i) => {
      const color = this.colors()[i % this.colors().length];
      g.append('path')
        .datum(s.samples)
        .attr('fill', 'none')
        .attr('stroke', color)
        .attr('stroke-width', 1.5)
        .attr('d', lineGen);
    });
  }
}
