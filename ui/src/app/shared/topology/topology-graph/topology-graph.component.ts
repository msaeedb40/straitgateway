import {
  Component, input, output, inject, ElementRef,
  AfterViewInit, OnDestroy, OnChanges, signal
} from '@angular/core';
import { isPlatformBrowser } from '@angular/common';
import { PLATFORM_ID } from '@angular/core';
import { TopologyGraph, TopologyNode, TopologyEdge } from '../../../core/models/topology.model';

@Component({
  selector: 'sg-topology-graph',
  template: `
    <div class="sg-topology-wrapper" role="application" [attr.aria-label]="'Network topology graph. ' + graph().nodes.length + ' nodes, ' + graph().edges.length + ' connections.'">
      <svg #topologySvg class="sg-topology-canvas" aria-hidden="true"></svg>
      <div class="sg-topology-a11y-list sg-sr-only" role="list" aria-label="Topology nodes">
        @for (node of graph().nodes; track node.id) {
          <div role="listitem">{{ node.kind }}: {{ node.name }} — {{ node.health }}</div>
        }
      </div>
    </div>
  `,
  styles: [`
    :host { display: block; width: 100%; height: 100%; }
    .sg-topology-wrapper { position: relative; width: 100%; height: 100%; }
    svg.sg-topology-canvas { width: 100%; height: 100%; }
  `],
})
export class TopologyGraphComponent implements AfterViewInit, OnDestroy, OnChanges {
  readonly graph        = input.required<TopologyGraph>();
  readonly nodeSelected = output<TopologyNode>();
  readonly edgeSelected = output<TopologyEdge>();

  private readonly el         = inject(ElementRef);
  private readonly platformId = inject(PLATFORM_ID);
  private d3Module: typeof import('d3') | null = null;
  private simulation: any = null;

  async ngAfterViewInit(): Promise<void> {
    if (!isPlatformBrowser(this.platformId)) return;
    this.d3Module = await import('d3');
    this.render();
  }

  ngOnChanges(): void {
    if (this.d3Module) this.render();
  }

  ngOnDestroy(): void {
    this.simulation?.stop();
  }

  private render(): void {
    const d3 = this.d3Module!;
    const svgEl = this.el.nativeElement.querySelector('svg') as SVGElement;
    const { width: W, height: H } = svgEl.getBoundingClientRect();
    if (!W || !H) return;

    const g = d3.select(svgEl);
    g.selectAll('*').remove();

    const zoom = d3.zoom<SVGElement, unknown>()
      .scaleExtent([0.2, 4])
      .on('zoom', (event) => {
        inner.attr('transform', event.transform);
      });
    g.call(zoom);

    const inner = g.append('g');

    const nodes: any[] = this.graph().nodes.map((n) => ({ ...n }));
    const edges = this.graph().edges;

    // Links
    const linkSel = inner.append('g').selectAll('line')
      .data(edges)
      .enter().append('line')
      .attr('class', (e) => `link${(e.trafficBytesPerSec ?? 0) > 1_000_000 ? ' traffic-high' : ''}`)
      .attr('stroke', 'var(--sg-border)')
      .attr('stroke-width', (e) => e.trafficBytesPerSec ? Math.min(Math.log10(e.trafficBytesPerSec + 1), 4) : 1.5)
      .attr('stroke-opacity', 0.6)
      .on('click', (_, e) => this.edgeSelected.emit(e));

    // Health color map
    const healthColor: Record<string, string> = {
      Healthy:  'var(--sg-healthy)',
      Degraded: 'var(--sg-degraded)',
      Failed:   'var(--sg-failed)',
      Unknown:  'var(--sg-unknown)',
    };

    // Nodes
    const nodeSel = inner.append('g').selectAll('g')
      .data(nodes)
      .enter().append('g')
      .attr('class', 'node')
      .attr('tabindex', '0')
      .attr('role', 'button')
      .attr('aria-label', (n: any) => `${n.kind} ${n.name}: ${n.health}`)
      .on('click', (_, n: any) => this.nodeSelected.emit(n))
      .on('keydown', (event, n: any) => { if (event.key === 'Enter') this.nodeSelected.emit(n); })
      .call(d3.drag<SVGGElement, any>()
        .on('start', (event, d: any) => { if (!event.active) sim.alphaTarget(0.3).restart(); d.fx = d.x; d.fy = d.y; })
        .on('drag',  (event, d: any) => { d.fx = event.x; d.fy = event.y; })
        .on('end',   (event, d: any) => { if (!event.active) sim.alphaTarget(0); d.fx = null; d.fy = null; })
      );

    nodeSel.append('circle')
      .attr('r', 8)
      .attr('fill', (n) => healthColor[n.health] ?? healthColor['Unknown'])
      .attr('fill-opacity', 0.85)
      .attr('stroke', 'var(--sg-bg-base)')
      .attr('stroke-width', 1.5);

    nodeSel.append('text')
      .attr('dy', 20)
      .attr('text-anchor', 'middle')
      .attr('font-size', 10)
      .attr('fill', 'var(--sg-text-secondary)')
      .text((n) => n.name.length > 16 ? n.name.slice(0, 14) + '…' : n.name);

    const nodeById = new Map(nodes.map((n) => [n.id, n]));

    const sim = d3.forceSimulation(nodes)
      .force('link', d3.forceLink(edges.map((e) => ({
        source: nodeById.get(e.sourceId)!,
        target: nodeById.get(e.targetId)!,
      }))).distance(80).strength(0.5))
      .force('charge', d3.forceManyBody().strength(-200))
      .force('center', d3.forceCenter(W / 2, H / 2))
      .force('collision', d3.forceCollide(20))
      .on('tick', () => {
        linkSel
          .attr('x1', (e) => (nodeById.get(e.sourceId) as { x: number })?.x ?? 0)
          .attr('y1', (e) => (nodeById.get(e.sourceId) as { y: number })?.y ?? 0)
          .attr('x2', (e) => (nodeById.get(e.targetId) as { x: number })?.x ?? 0)
          .attr('y2', (e) => (nodeById.get(e.targetId) as { y: number })?.y ?? 0);
        nodeSel.attr('transform', (n) => `translate(${n.x},${n.y})`);
      });

    this.simulation = sim;
  }
}
