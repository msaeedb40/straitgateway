// Copyright 2026 straitgateway Authors
// SPDX-License-Identifier: Apache-2.0

import {
  Component,
  ElementRef,
  OnDestroy,
  OnInit,
  ViewChild,
  inject,
  input,
  output,
  signal,
} from '@angular/core';
import { CommonModule } from '@angular/common';
import * as d3 from 'd3';

import { PLATFORM_ID } from '@angular/core';
import { isPlatformBrowser } from '@angular/common';
import { Router } from '@angular/router';
import { ContextMenuService } from '../../layout/context-menu/context-menu.service';

export interface TopologyNode extends d3.SimulationNodeDatum {
  id: string;
  name: string;
  type: 'gateway' | 'node' | 'tunnel' | 'service' | 'client';
  namespace?: string;
  status: 'healthy' | 'warning' | 'error';
  details?: Record<string, any>;
}

export interface TopologyLink extends d3.SimulationLinkDatum<TopologyNode> {
  source: string | TopologyNode;
  target: string | TopologyNode;
  trafficRate?: string;
  active?: boolean;
}

interface Particle {
  sourceNode: TopologyNode;
  targetNode: TopologyNode;
  progress: number;
  speed: number;
  color: string;
}

@Component({
  selector: 'sg-topology-graph',
  standalone: true,
  imports: [CommonModule],
  template: `
    <div class="sg-topology-container" style="position:relative;width:100%;height:100%;min-height:340px;background:radial-gradient(ellipse at 50% 50%, rgba(99,102,241,0.12) 0%, transparent 70%), radial-gradient(rgba(255,255,255,0.07) 1px, transparent 1px), linear-gradient(180deg, rgba(16,19,29,0.9) 0%, rgba(10,12,18,0.9) 100%);background-size:100% 100%, 24px 24px, 100% 100%;border:1px solid var(--sg-border);border-radius:var(--sg-radius);overflow:hidden">
      <!-- Toolbar controls -->
      <div style="position:absolute;top:12px;left:12px;z-index:10;display:flex;align-items:center;gap:8px">
        <span class="sg-badge active" style="display:flex;align-items:center;gap:5px">
          <span style="display:inline-block;width:6px;height:6px;border-radius:50%;background:var(--sg-success);animation:pulse 2s infinite"></span>
          Live Traffic Active
        </span>
        <span style="font-size:11px;color:var(--sg-text-3)">eBPF NetKit Real-time</span>
      </div>

      <div style="position:absolute;top:12px;right:12px;z-index:10;display:flex;align-items:center;gap:6px">
        <button
          (click)="zoomIn()"
          class="sg-btn sg-btn-secondary"
          style="padding:4px 8px;font-size:12px"
          title="Zoom In"
        >+</button>
        <button
          (click)="zoomOut()"
          class="sg-btn sg-btn-secondary"
          style="padding:4px 8px;font-size:12px"
          title="Zoom Out"
        >−</button>
        <button
          (click)="resetZoom()"
          class="sg-btn sg-btn-secondary"
          style="padding:4px 8px;font-size:11px"
          title="Reset View"
        >Reset</button>
      </div>

      <!-- Graph SVG and Overlay Canvas -->
      <svg #svgRef style="width:100%;height:100%;display:block;cursor:grab"></svg>

      <!-- Selected Node Card -->
      @if (selectedNode()) {
        <div
          class="sg-card"
          style="position:absolute;bottom:12px;right:12px;z-index:20;width:260px;background:var(--sg-surface-2);box-shadow:0 12px 30px rgba(0,0,0,.6);border:1px solid var(--sg-border-hover);padding:12px"
        >
          <div style="display:flex;align-items:center;justify-content:space-between;margin-bottom:6px">
            <span style="font-weight:600;font-size:13px;color:var(--sg-text)">{{ selectedNode()?.name }}</span>
            <button (click)="selectedNode.set(null)" style="color:var(--sg-text-3);font-size:12px">✕</button>
          </div>
          <div style="font-size:11px;color:var(--sg-text-2);display:flex;flex-direction:column;gap:4px">
            <div>Type: <span class="mono" style="color:var(--sg-accent-light)">{{ selectedNode()?.type }}</span></div>
            <div>Status: <span [style.color]="selectedNode()?.status === 'healthy' ? 'var(--sg-success)' : 'var(--sg-warn)'">{{ selectedNode()?.status }}</span></div>
            @if (selectedNode()?.namespace) {
              <div>Namespace: <span class="mono">{{ selectedNode()?.namespace }}</span></div>
            }
          </div>
        </div>
      }
    </div>
  `,
  styles: [`
    @keyframes pulse {
      0% { opacity: 0.4; }
      50% { opacity: 1; }
      100% { opacity: 0.4; }
    }
  `],
})
export class TopologyGraphComponent implements OnInit, OnDestroy {
  @ViewChild('svgRef', { static: true }) svgRef!: ElementRef<SVGSVGElement>;

  height = input<number>(380);
  nodeClick = output<TopologyNode>();

  selectedNode = signal<TopologyNode | null>(null);

  private simulation!: d3.Simulation<TopologyNode, TopologyLink>;
  private zoomBehavior!: d3.ZoomBehavior<SVGSVGElement, unknown>;
  private gContainer!: d3.Selection<SVGGElement, unknown, null, undefined>;
  private animFrameId: number | null = null;
  private particles: Particle[] = [];

  nodes: TopologyNode[] = [
    { id: 'client', name: 'External Client', type: 'client', status: 'healthy', x: 60, y: 190 },
    { id: 'gw-edge', name: 'sg-edge-gw', type: 'gateway', status: 'healthy', namespace: 'default', x: 220, y: 190 },
    { id: 'tunnel-c2', name: 'Transit Cluster-2', type: 'tunnel', status: 'healthy', namespace: 'straitgateway-system', x: 220, y: 80 },
    { id: 'node-cp', name: 'Control Plane 01', type: 'node', status: 'healthy', x: 380, y: 110 },
    { id: 'node-w1', name: 'Worker Node 01', type: 'node', status: 'healthy', x: 380, y: 190 },
    { id: 'node-w2', name: 'Worker Node 02', type: 'node', status: 'healthy', x: 380, y: 270 },
    { id: 'svc-core', name: 'core-api (Maglev)', type: 'service', status: 'healthy', namespace: 'default', x: 540, y: 190 },
  ];

  links: TopologyLink[] = [
    { source: 'client', target: 'gw-edge', trafficRate: '4.2k req/s', active: true },
    { source: 'gw-edge', target: 'tunnel-c2', trafficRate: '850 req/s', active: true },
    { source: 'gw-edge', target: 'node-w1', trafficRate: '2.1k req/s', active: true },
    { source: 'gw-edge', target: 'node-w2', trafficRate: '1.2k req/s', active: true },
    { source: 'node-w1', target: 'svc-core', trafficRate: '2.1k req/s', active: true },
    { source: 'node-w2', target: 'svc-core', trafficRate: '1.2k req/s', active: true },
  ];

  private platformId = inject(PLATFORM_ID);
  private contextMenu = inject(ContextMenuService);
  private router = inject(Router);

  ngOnInit() {
    if (isPlatformBrowser(this.platformId)) {
      this.initGraph();
    }
  }

  ngOnDestroy() {
    if (this.simulation) this.simulation.stop();
    if (this.animFrameId && typeof cancelAnimationFrame !== 'undefined') {
      cancelAnimationFrame(this.animFrameId);
    }
  }

  private initGraph() {
    if (!isPlatformBrowser(this.platformId) || !this.svgRef) return;
    const svg = d3.select(this.svgRef.nativeElement);
    svg.selectAll('*').remove();

    const width = this.svgRef.nativeElement.clientWidth || 680;
    const height = this.height() || 380;

    // Zoom setup
    this.gContainer = svg.append('g').attr('class', 'main-group');
    this.zoomBehavior = d3
      .zoom<SVGSVGElement, unknown>()
      .scaleExtent([0.4, 3])
      .on('zoom', (event) => {
        this.gContainer.attr('transform', event.transform);
      });
    svg.call(this.zoomBehavior);

    // Marker defs for links
    const defs = svg.append('defs');
    defs
      .append('marker')
      .attr('id', 'arrow')
      .attr('viewBox', '0 -5 10 10')
      .attr('refX', 22)
      .attr('refY', 0)
      .attr('markerWidth', 6)
      .attr('markerHeight', 6)
      .attr('orient', 'auto')
      .append('path')
      .attr('d', 'M0,-5L10,0L0,5')
      .attr('fill', 'rgba(129,140,248,0.4)');

    // Link layer
    const linkGroup = this.gContainer.append('g').attr('class', 'links');
    const particleGroup = this.gContainer.append('g').attr('class', 'particles');
    const nodeGroup = this.gContainer.append('g').attr('class', 'nodes');

    // Force simulation
    this.simulation = d3
      .forceSimulation<TopologyNode>(this.nodes)
      .force(
        'link',
        d3
          .forceLink<TopologyNode, TopologyLink>(this.links)
          .id((d) => d.id)
          .distance(120)
      )
      .force('charge', d3.forceManyBody().strength(-300))
      .force('center', d3.forceCenter(width / 2, height / 2))
      .force('collision', d3.forceCollide().radius(35));

    // Render links
    const linkElements = linkGroup
      .selectAll<SVGLineElement, TopologyLink>('line')
      .data(this.links)
      .join('line')
      .attr('stroke', 'rgba(255,255,255,0.12)')
      .attr('stroke-width', 2)
      .attr('stroke-dasharray', (d) => (d.target === 'tunnel-c2' ? '4 3' : 'none'))
      .attr('marker-end', 'url(#arrow)');

    // Render nodes
    const nodeElements = nodeGroup
      .selectAll<SVGGElement, TopologyNode>('g')
      .data(this.nodes)
      .join('g')
      .attr('cursor', 'pointer')
      .call(
        d3
          .drag<SVGGElement, TopologyNode>()
          .on('start', (event, d) => {
            if (!event.active) this.simulation.alphaTarget(0.3).restart();
            d.fx = d.x;
            d.fy = d.y;
          })
          .on('drag', (event, d) => {
            d.fx = event.x;
            d.fy = event.y;
          })
          .on('end', (event, d) => {
            if (!event.active) this.simulation.alphaTarget(0);
            d.fx = null;
            d.fy = null;
          })
      );

    // Node circles & halos
    nodeElements
      .append('circle')
      .attr('r', 20)
      .attr('fill', (d) => this.getNodeColor(d.type))
      .attr('stroke', (d) => (d.status === 'healthy' ? '#818cf8' : '#f59e0b'))
      .attr('stroke-width', 2)
      .attr('fill-opacity', 0.25);

    nodeElements
      .append('circle')
      .attr('r', 7)
      .attr('fill', (d) => this.getNodeColor(d.type));

    // Node labels
    nodeElements
      .append('text')
      .attr('text-anchor', 'middle')
      .attr('dy', 34)
      .attr('fill', 'var(--sg-text)')
      .attr('font-size', '11px')
      .attr('font-weight', '500')
      .text((d) => d.name);

    nodeElements
      .append('text')
      .attr('text-anchor', 'middle')
      .attr('dy', 47)
      .attr('fill', 'var(--sg-text-3)')
      .attr('font-size', '9px')
      .attr('font-family', 'monospace')
      .text((d) => d.type.toUpperCase());

    // Click handler
    nodeElements.on('click', (event, d) => {
      event.stopPropagation();
      this.selectedNode.set(d);
      this.nodeClick.emit(d);
    });

    // Context menu handler
    nodeElements.on('contextmenu', (event, d) => {
      event.preventDefault();
      event.stopPropagation();
      this.contextMenu.open(
        event.clientX,
        event.clientY,
        [
          {
            label: `Inspect ${d.name}`,
            action: () => {
              this.selectedNode.set(d);
              this.nodeClick.emit(d);
            },
          },
          {
            label: 'View Live Flows',
            action: () => this.router.navigate(['/flows']),
          },
          {
            label: 'Capture Packets',
            action: () => this.router.navigate(['/packets']),
          },
          {
            label: 'View Metrics',
            action: () => this.router.navigate(['/metrics']),
          },
        ],
        d
      );
    });

    // Tick simulation
    this.simulation.on('tick', () => {
      linkElements
        .attr('x1', (d) => (d.source as TopologyNode).x!)
        .attr('y1', (d) => (d.source as TopologyNode).y!)
        .attr('x2', (d) => (d.target as TopologyNode).x!)
        .attr('y2', (d) => (d.target as TopologyNode).y!);

      nodeElements.attr('transform', (d) => `translate(${d.x},${d.y})`);
    });

    // Particle flow animation
    this.spawnParticles();
    const renderParticles = () => {
      this.updateParticles();
      const pElements = particleGroup
        .selectAll<SVGCircleElement, Particle>('circle')
        .data(this.particles);

      pElements
        .join('circle')
        .attr('r', 3.5)
        .attr('fill', (d) => d.color)
        .attr('cx', (d) => {
          const sx = d.sourceNode.x || 0;
          const tx = d.targetNode.x || 0;
          return sx + (tx - sx) * d.progress;
        })
        .attr('cy', (d) => {
          const sy = d.sourceNode.y || 0;
          const ty = d.targetNode.y || 0;
          return sy + (ty - sy) * d.progress;
        });

      this.animFrameId = requestAnimationFrame(renderParticles);
    };
    renderParticles();
  }

  private spawnParticles() {
    this.particles = [];
    this.links.forEach((link, idx) => {
      const src = typeof link.source === 'object' ? (link.source as TopologyNode) : this.nodes.find(n => n.id === link.source)!;
      const tgt = typeof link.target === 'object' ? (link.target as TopologyNode) : this.nodes.find(n => n.id === link.target)!;
      if (src && tgt) {
        for (let i = 0; i < 3; i++) {
          this.particles.push({
            sourceNode: src,
            targetNode: tgt,
            progress: i * 0.33,
            speed: 0.006 + Math.random() * 0.004,
            color: idx === 1 ? '#38bdf8' : '#22c55e',
          });
        }
      }
    });
  }

  private updateParticles() {
    this.particles.forEach((p) => {
      p.progress += p.speed;
      if (p.progress >= 1) {
        p.progress = 0;
      }
    });
  }

  private getNodeColor(type: string): string {
    switch (type) {
      case 'gateway': return '#6366f1';
      case 'tunnel': return '#38bdf8';
      case 'node': return '#10b981';
      case 'service': return '#a855f7';
      default: return '#9ca3af';
    }
  }

  zoomIn() {
    const svg = d3.select(this.svgRef.nativeElement);
    svg.transition().duration(250).call(this.zoomBehavior.scaleBy, 1.3);
  }

  zoomOut() {
    const svg = d3.select(this.svgRef.nativeElement);
    svg.transition().duration(250).call(this.zoomBehavior.scaleBy, 0.7);
  }

  resetZoom() {
    const svg = d3.select(this.svgRef.nativeElement);
    svg.transition().duration(250).call(this.zoomBehavior.transform, d3.zoomIdentity);
  }
}
