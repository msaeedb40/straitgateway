// Copyright 2026 straitgateway Authors
// SPDX-License-Identifier: Apache-2.0

import { Component, inject, signal, OnInit, ElementRef, ViewChild, AfterViewInit, PLATFORM_ID } from '@angular/core';
import { CommonModule, isPlatformBrowser } from '@angular/common';
import { ApiClient, NodeStatus, GatewaySummary, ServiceSummary } from '../../core/api/api-client';
import { catchError, forkJoin, of } from 'rxjs';

interface TopologyNode {
  id: string;
  label: string;
  type: 'node' | 'gateway' | 'service' | 'pod' | 'transit';
  x: number;
  y: number;
  status: 'healthy' | 'warning' | 'error';
  metadata?: Record<string, string>;
}

interface TopologyEdge {
  from: string;
  to: string;
  label?: string;
  type: 'data' | 'control' | 'tunnel';
}

@Component({
  selector: 'app-topology',
  standalone: true,
  imports: [CommonModule],
  template: `
<div class="sg-fade-in">
  <div class="sg-page-header">
    <div>
      <h1 class="sg-page-title">Network Topology</h1>
      <p class="sg-page-subtitle">Interactive cluster network visualization</p>
    </div>
    <div style="display:flex;gap:8px;align-items:center">
      <div class="topo-legend">
        <span class="topo-legend-item"><span class="topo-dot" style="background:var(--sg-accent)"></span>Node</span>
        <span class="topo-legend-item"><span class="topo-dot" style="background:var(--sg-success)"></span>Gateway</span>
        <span class="topo-legend-item"><span class="topo-dot" style="background:#a78bfa"></span>Service</span>
        <span class="topo-legend-item"><span class="topo-dot" style="background:var(--sg-warn)"></span>Transit</span>
      </div>
      <button class="sg-btn sg-btn-secondary" (click)="load()">Refresh</button>
    </div>
  </div>

  <div class="sg-card">
    <div class="sg-card-header">
      <span class="sg-card-title">Cluster Topology</span>
      <span class="text-muted text-sm">Nodes: {{ nodes().length }} | Services: {{ services().length }} | Gateways: {{ gateways().length }}</span>
    </div>
    <div class="sg-card-body topo-container" style="min-height:500px;position:relative;overflow:hidden">
      <canvas #topoCanvas width="1200" height="500" style="width:100%;height:100%;display:block"></canvas>
    </div>
  </div>

  <!-- Node detail panel -->
  @if (selectedNode(); as node) {
    <div class="sg-card" style="margin-top:24px">
      <div class="sg-card-header">
        <span class="sg-card-title">{{ node.label }}</span>
        <button class="sg-btn sg-btn-secondary sg-btn-sm" (click)="selectedNode.set(null)">Close</button>
      </div>
      <div class="sg-card-body">
        <div style="display:grid;grid-template-columns:repeat(auto-fill,minmax(180px,1fr));gap:16px">
          <div class="info-pair"><span class="info-label">Type</span><span class="info-value">{{ node.type }}</span></div>
          <div class="info-pair"><span class="info-label">Status</span><span class="sg-badge" [class]="node.status === 'healthy' ? 'active' : node.status === 'warning' ? 'warn' : 'error'">{{ node.status }}</span></div>
          @for (entry of objectEntries(node.metadata || {}); track entry[0]) {
            <div class="info-pair"><span class="info-label">{{ entry[0] }}</span><span class="info-value mono">{{ entry[1] }}</span></div>
          }
        </div>
      </div>
    </div>
  }
</div>
`,
  styles: [`
    .topo-legend { display:flex;gap:12px;font-size:12px; }
    .topo-legend-item { display:flex;align-items:center;gap:4px; }
    .topo-dot { width:8px;height:8px;border-radius:50%;display:inline-block; }
    .info-pair { display:flex;flex-direction:column;gap:2px; }
    .info-label { font-size:11px;color:var(--sg-text-3);text-transform:uppercase;letter-spacing:0.5px; }
    .info-value { font-size:13px;color:var(--sg-text-1); }
  `],
})
export class TopologyComponent implements OnInit, AfterViewInit {
  private api = inject(ApiClient);
  @ViewChild('topoCanvas') canvasRef!: ElementRef<HTMLCanvasElement>;

  nodes    = signal<NodeStatus[]>([]);
  gateways = signal<GatewaySummary[]>([]);
  services = signal<ServiceSummary[]>([]);
  selectedNode = signal<TopologyNode | null>(null);

  private topoNodes: TopologyNode[] = [];
  private topoEdges: TopologyEdge[] = [];

  objectEntries(obj: Record<string, string>) { return Object.entries(obj); }

  ngOnInit() { this.load(); }
  ngAfterViewInit() { }

  load() {
    forkJoin({
      nodes: this.api.getNodes().pipe(catchError(() => of([]))),
      gateways: this.api.getGateways().pipe(catchError(() => of([]))),
      services: this.api.getServices().pipe(catchError(() => of([]))),
    }).subscribe(({ nodes, gateways, services }) => {
      this.nodes.set(nodes);
      this.gateways.set(gateways);
      this.services.set(services);
      this.buildTopology(nodes, gateways, services);
      this.render();
    });
  }

  private buildTopology(nodes: NodeStatus[], gateways: GatewaySummary[], services: ServiceSummary[]) {
    this.topoNodes = [];
    this.topoEdges = [];
    const cx = 600, cy = 250;

    // Nodes in a circle
    nodes.forEach((n, i) => {
      const angle = (2 * Math.PI * i) / Math.max(nodes.length, 1) - Math.PI / 2;
      const r = 160;
      this.topoNodes.push({
        id: `node-${n.name}`, label: n.name, type: 'node',
        x: cx + r * Math.cos(angle), y: cy + r * Math.sin(angle),
        status: n.cniReady && n.serviceReady ? 'healthy' : 'warning',
        metadata: { ip: n.ip, podCIDR: n.podCIDR, kernel: n.kernelVersion },
      });
    });

    // Gateways around the outside
    gateways.forEach((gw, i) => {
      const angle = (2 * Math.PI * i) / Math.max(gateways.length, 1);
      const r = 220;
      this.topoNodes.push({
        id: `gw-${gw.namespace}-${gw.name}`, label: gw.name, type: 'gateway',
        x: cx + r * Math.cos(angle), y: cy + r * Math.sin(angle),
        status: gw.ready ? 'healthy' : 'warning',
        metadata: { namespace: gw.namespace, class: gw.gatewayClass, addresses: gw.addresses.join(', ') },
      });
      // Connect gateway to all nodes
      nodes.forEach(n => {
        this.topoEdges.push({ from: `gw-${gw.namespace}-${gw.name}`, to: `node-${n.name}`, type: 'control' });
      });
    });

    // Services in the center
    services.slice(0, 12).forEach((svc, i) => {
      const angle = (2 * Math.PI * i) / Math.min(services.length, 12);
      const r = 80;
      this.topoNodes.push({
        id: `svc-${svc.namespace}-${svc.name}`, label: svc.name, type: 'service',
        x: cx + r * Math.cos(angle), y: cy + r * Math.sin(angle),
        status: svc.backendCount > 0 ? 'healthy' : 'error',
        metadata: { clusterIP: svc.clusterIP, type: svc.type, backends: String(svc.backendCount) },
      });
    });

    // Node-to-node mesh edges
    for (let i = 0; i < nodes.length; i++) {
      for (let j = i + 1; j < nodes.length; j++) {
        this.topoEdges.push({
          from: `node-${nodes[i].name}`, to: `node-${nodes[j].name}`, type: 'data',
        });
      }
    }
  }

  private platformId = inject(PLATFORM_ID);

  private render() {
    if (!isPlatformBrowser(this.platformId)) return;
    const canvas = this.canvasRef?.nativeElement;
    if (!canvas) return;
    const ctx = canvas.getContext('2d');
    if (!ctx) return;

    const dpr = window.devicePixelRatio || 1;
    const rect = canvas.getBoundingClientRect();
    canvas.width = rect.width * dpr;
    canvas.height = rect.height * dpr;
    ctx.scale(dpr, dpr);

    ctx.clearRect(0, 0, rect.width, rect.height);

    const scaleX = rect.width / 1200;
    const scaleY = rect.height / 500;

    // Draw edges
    for (const edge of this.topoEdges) {
      const from = this.topoNodes.find(n => n.id === edge.from);
      const to = this.topoNodes.find(n => n.id === edge.to);
      if (!from || !to) continue;

      ctx.beginPath();
      ctx.moveTo(from.x * scaleX, from.y * scaleY);
      ctx.lineTo(to.x * scaleX, to.y * scaleY);
      ctx.strokeStyle = edge.type === 'data' ? 'rgba(100,180,255,0.15)' :
                        edge.type === 'tunnel' ? 'rgba(250,200,50,0.3)' :
                        'rgba(160,120,255,0.2)';
      ctx.lineWidth = edge.type === 'data' ? 1 : 1.5;
      if (edge.type === 'tunnel') { ctx.setLineDash([4, 4]); } else { ctx.setLineDash([]); }
      ctx.stroke();
    }

    ctx.setLineDash([]);

    // Draw nodes
    for (const node of this.topoNodes) {
      const x = node.x * scaleX;
      const y = node.y * scaleY;
      const r = node.type === 'node' ? 18 : node.type === 'gateway' ? 14 : 10;

      // Glow
      const gradient = ctx.createRadialGradient(x, y, 0, x, y, r * 2.5);
      const color = node.type === 'node' ? '100,180,255' :
                    node.type === 'gateway' ? '80,220,160' :
                    node.type === 'service' ? '167,139,250' : '250,200,50';
      gradient.addColorStop(0, `rgba(${color},0.3)`);
      gradient.addColorStop(1, `rgba(${color},0)`);
      ctx.fillStyle = gradient;
      ctx.beginPath();
      ctx.arc(x, y, r * 2.5, 0, Math.PI * 2);
      ctx.fill();

      // Circle
      ctx.beginPath();
      ctx.arc(x, y, r, 0, Math.PI * 2);
      ctx.fillStyle = node.status === 'healthy' ? `rgba(${color},0.9)` :
                      node.status === 'warning' ? `rgba(250,200,50,0.9)` : `rgba(239,68,68,0.9)`;
      ctx.fill();
      ctx.strokeStyle = `rgba(${color},0.5)`;
      ctx.lineWidth = 2;
      ctx.stroke();

      // Label
      ctx.fillStyle = 'rgba(255,255,255,0.85)';
      ctx.font = '10px Inter, sans-serif';
      ctx.textAlign = 'center';
      ctx.fillText(node.label, x, y + r + 14);
    }

    // Click handler
    canvas.onclick = (e: MouseEvent) => {
      const cr = canvas.getBoundingClientRect();
      const mx = e.clientX - cr.left;
      const my = e.clientY - cr.top;
      for (const node of this.topoNodes) {
        const x = node.x * scaleX;
        const y = node.y * scaleY;
        const dist = Math.sqrt((mx - x) ** 2 + (my - y) ** 2);
        if (dist < 20) {
          this.selectedNode.set(node);
          return;
        }
      }
      this.selectedNode.set(null);
    };
  }
}
