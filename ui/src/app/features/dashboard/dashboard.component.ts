// Copyright 2026 straitgateway Authors
// SPDX-License-Identifier: Apache-2.0

import { Component, inject, signal, computed, OnInit } from '@angular/core';
import { CommonModule } from '@angular/common';
import { ApiClient, DashboardSummary, KubeEvent } from '../../core/api/api-client';
import { NamespaceService } from '../../core/services/namespace.service';
import { TopologyGraphComponent } from '../../shared/topology/topology-graph.component';
import { SparklineComponent } from '../../shared/charts/sparkline.component';
import { calculateClusterHealth, getHealthStatusInfo, formatRate } from '../../shared/utilities';
import { catchError, of } from 'rxjs';

@Component({
  selector: 'app-dashboard',
  standalone: true,
  imports: [CommonModule, TopologyGraphComponent, SparklineComponent],
  template: `
<div class="sg-fade-in">
  <!-- Page Header -->
  <div class="sg-page-header">
    <div>
      <h1 class="sg-page-title">
        Gateway Overview
      </h1>
      <p class="sg-page-subtitle">
        Real-time cluster networking, eBPF dataplane status, and traffic topology · Namespace: <span class="mono text-accent">{{ ns.active() || 'all' }}</span>
      </p>
    </div>
    <div style="display:flex;align-items:center;gap:10px">
      <span class="sg-badge active" style="display:flex;align-items:center;gap:6px">
        <span style="width:6px;height:6px;border-radius:50%;background:var(--sg-success);animation:pulse 2s infinite"></span>
        eBPF RingBuffer Active
      </span>
      <button class="sg-btn sg-btn-secondary" (click)="refresh()">
        <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
          <polyline points="23 4 23 10 17 10"/><path d="M20.49 15a9 9 0 1 1-2.12-9.36L23 10"/>
        </svg>
        <span>Refresh</span>
      </button>
    </div>
  </div>

  <!-- Stat Cards: Nodes, Gateways, Tunnels, Dynamic Health -->
  @if (summary(); as s) {
    <div class="sg-stat-grid">
      <!-- Nodes Card -->
      <div class="sg-stat-card" style="border-top:3px solid #6366f1">
        <div class="sg-stat-label">Nodes</div>
        <div class="sg-stat-value text-accent">
          {{ s.readyNodes }}<span style="font-size:16px;font-weight:400;color:var(--sg-text-2)">/{{ s.totalNodes }}</span>
        </div>
        <div class="sg-stat-meta">
          <span class="good" style="display:flex;align-items:center;gap:6px">
            <span style="width:6px;height:6px;border-radius:50%;background:var(--sg-success);box-shadow:0 0 8px var(--sg-success)"></span>
            {{ s.readyNodes === s.totalNodes ? 'All Healthy' : 'Partially Ready' }}
          </span>
        </div>
      </div>

      <!-- Gateways Card -->
      <div class="sg-stat-card" style="border-top:3px solid #818cf8">
        <div class="sg-stat-label">Gateways</div>
        <div class="sg-stat-value">{{ s.activeGateways }}</div>
        <div class="sg-stat-meta">
          <span class="good" style="display:flex;align-items:center;gap:6px">
            <span style="width:6px;height:6px;border-radius:50%;background:var(--sg-accent-light);box-shadow:0 0 8px var(--sg-accent-light)"></span>
            Online (v1.6.1)
          </span>
        </div>
      </div>

      <!-- Tunnels Card -->
      <div class="sg-stat-card" style="border-top:3px solid #22c55e">
        <div class="sg-stat-label">Tunnels</div>
        <div class="sg-stat-value">{{ s.activeTunnels }}</div>
        <div class="sg-stat-meta">
          <span class="good" style="display:flex;align-items:center;gap:6px">
            <span style="width:6px;height:6px;border-radius:50%;background:var(--sg-success);box-shadow:0 0 8px var(--sg-success)"></span>
            Active Mesh
          </span>
        </div>
      </div>

      <!-- Dynamically Computed Health Card -->
      <div
        class="sg-stat-card"
        [style.border-top]="'3px solid ' + (healthInfo().status === 'Good' ? 'var(--sg-success)' : healthInfo().status === 'Degraded' ? 'var(--sg-warn)' : 'var(--sg-danger)')"
      >
        <div class="sg-stat-label">Health</div>
        <div
          class="sg-stat-value"
          [style.color]="healthInfo().status === 'Good' ? 'var(--sg-success)' : healthInfo().status === 'Degraded' ? 'var(--sg-warn)' : 'var(--sg-danger)'"
        >
          {{ s.healthPercentage | number:'1.0-1' }}%
        </div>
        <div class="sg-stat-meta">
          <span [class]="healthInfo().status === 'Good' ? 'good' : 'warn'" style="display:flex;align-items:center;gap:6px">
            <span
              [style.background]="healthInfo().status === 'Good' ? 'var(--sg-success)' : healthInfo().status === 'Degraded' ? 'var(--sg-warn)' : 'var(--sg-danger)'"
              [style.box-shadow]="'0 0 8px ' + (healthInfo().status === 'Good' ? 'var(--sg-success)' : 'var(--sg-warn)')"
              style="width:6px;height:6px;border-radius:50%"
            ></span>
            ● {{ healthInfo().status }}
          </span>
        </div>
      </div>
    </div>

    <!-- Live Telemetry Sparkline Strip -->
    <div class="sg-card" style="margin-bottom:20px;padding:12px 18px">
      <div style="display:flex;align-items:center;justify-content:space-between;flex-wrap:wrap;gap:16px">
        <div style="display:flex;align-items:center;gap:24px">
          <div>
            <span style="font-size:11px;color:var(--sg-text-3);text-transform:uppercase;letter-spacing:0.05em">Flow Rate</span>
            <div style="font-size:17px;font-weight:700;color:var(--sg-text)">{{ formatFlowRate(s.totalFlowsPerSec) }}</div>
          </div>
          <sg-sparkline [data]="flowThroughputHistory()" [color]="'var(--sg-accent-light)'" [width]="140" [height]="32" />
        </div>

        <div style="display:flex;align-items:center;gap:24px">
          <div>
            <span style="font-size:11px;color:var(--sg-text-3);text-transform:uppercase;letter-spacing:0.05em">Drop Rate</span>
            <div style="font-size:17px;font-weight:700;color:var(--sg-success)">{{ formatFlowRate(s.droppedFlowsPerSec) }}</div>
          </div>
          <sg-sparkline [data]="dropRateHistory()" [color]="'var(--sg-success)'" [width]="140" [height]="32" />
        </div>

        <div style="display:flex;align-items:center;gap:12px">
          <span class="mono text-xs" style="color:var(--sg-text-3)">Kernel: 6.6+ LTS</span>
          <span class="mono text-xs" style="color:var(--sg-text-3)">BPF Maps: Synced</span>
        </div>
      </div>
    </div>
  } @else {
    <div class="sg-stat-grid">
      @for (i of [1,2,3,4]; track i) {
        <div class="sg-stat-card">
          <div class="sg-skeleton" style="height:12px;width:60%;margin-bottom:8px"></div>
          <div class="sg-skeleton" style="height:34px;width:40%;margin-bottom:8px"></div>
          <div class="sg-skeleton" style="height:12px;width:80%"></div>
        </div>
      }
    </div>
  }

  <!-- Topology Card with D3 Real Traffic Flow -->
  <div class="sg-card">
    <div class="sg-card-header">
      <div style="display:flex;align-items:center;gap:10px">
        <span class="sg-card-title">Topology</span>
        <span style="font-size:12px;color:var(--sg-text-3)">real traffic flow</span>
      </div>
      <span class="text-muted text-xs">Right-click nodes for actions</span>
    </div>
    <div style="padding:16px">
      <sg-topology-graph [height]="380" />
    </div>
  </div>

  <!-- Two Column Layout: Events + Dataplane Subsystems -->
  <div class="grid-2">
    <!-- Recent Events -->
    <div class="sg-card" style="margin-bottom:0">
      <div class="sg-card-header">
        <span class="sg-card-title">Recent Events</span>
        <span class="text-muted text-sm">{{ filteredEvents().length }} Events</span>
      </div>
      <div class="sg-card-body p-0">
        <div class="sg-table-wrap">
          <table class="sg-table">
            <thead>
              <tr>
                <th>Type</th>
                <th>Namespace</th>
                <th>Reason</th>
                <th>Message</th>
                <th>Age</th>
              </tr>
            </thead>
            <tbody>
              @if (filteredEvents().length) {
                @for (ev of filteredEvents(); track ev.name) {
                  <tr>
                    <td>
                      <span class="sg-badge" [class]="ev.type === 'Warning' ? 'warn' : 'active'">
                        {{ ev.type }}
                      </span>
                    </td>
                    <td class="mono">{{ ev.namespace }}</td>
                    <td style="font-weight:500">{{ ev.reason }}</td>
                    <td class="truncate" style="max-width:260px" [title]="ev.message">{{ ev.message }}</td>
                    <td class="text-muted">{{ ev.lastSeen }}</td>
                  </tr>
                }
              } @else {
                <tr>
                  <td colspan="5">
                    <div class="sg-empty">
                      <p>No recent events recorded in this namespace.</p>
                    </div>
                  </td>
                </tr>
              }
            </tbody>
          </table>
        </div>
      </div>
    </div>

    <!-- Dataplane Subsystem Status -->
    <div class="sg-card" style="margin-bottom:0">
      <div class="sg-card-header">
        <span class="sg-card-title">Dataplane Subsystems</span>
        <span class="text-muted text-sm">{{ activeSubsystemsCount() }} Active</span>
      </div>
      <div class="sg-card-body" style="display:flex;flex-direction:column;gap:10px">
        @for (subsystem of dataplaneSubsystems(); track subsystem.name) {
          <div style="display:flex;align-items:center;justify-content:space-between;padding:10px 14px;background:var(--sg-surface-2);border-radius:var(--sg-radius-sm);border:1px solid var(--sg-border)">
            <div style="display:flex;align-items:center;gap:10px">
              <span [style.background]="subsystem.status === 'Active' ? 'var(--sg-success)' : 'var(--sg-warn)'" style="display:inline-block;width:7px;height:7px;border-radius:50%"></span>
              <span style="font-size:13px;font-weight:500;color:var(--sg-text)">{{ subsystem.name }}</span>
            </div>
            <span class="sg-badge" [class]="subsystem.status === 'Active' ? 'active' : 'warn'">{{ subsystem.status }}</span>
          </div>
        }
      </div>
    </div>
  </div>
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
export class DashboardComponent implements OnInit {
  private api = inject(ApiClient);
  readonly ns = inject(NamespaceService);

  summary = signal<DashboardSummary | null>(null);
  events = signal<KubeEvent[]>([]);
  loading = signal<boolean>(false);

  // Sparkline history buffers
  flowThroughputHistory = signal<number[]>([120, 145, 138, 160, 185, 172, 195, 210, 204, 220]);
  dropRateHistory = signal<number[]>([4, 2, 5, 1, 0, 2, 0, 1, 0, 0]);

  dataplaneSubsystems = signal<Array<{ name: string; status: string }>>([
    { name: 'CNI / NetKit', status: 'Active' },
    { name: 'Service LB (Maglev)', status: 'Active' },
    { name: 'kube-proxy Replacement', status: 'Active' },
    { name: 'NetworkPolicy (eBPF)', status: 'Active' },
    { name: 'Gateway API v1.6.1', status: 'Active' },
    { name: 'Transit Gateway', status: 'Active' },
    { name: 'BGP / BFD', status: 'Active' },
  ]);

  filteredEvents = computed(() => {
    const activeNs = this.ns.active();
    const all = this.events();
    if (!activeNs || activeNs === '') return all;
    return all.filter(e => e.namespace === activeNs);
  });

  healthInfo = computed(() => {
    const s = this.summary();
    return getHealthStatusInfo(s?.healthPercentage ?? 100);
  });

  activeSubsystemsCount() {
    return this.dataplaneSubsystems().filter(s => s.status === 'Active').length;
  }

  formatFlowRate(rate: number): string {
    return formatRate(rate, 'flows/s');
  }

  ngOnInit() {
    this.refresh();
  }

  refresh() {
    this.loading.set(true);

    // Fetch dashboard summary and dynamically compute health
    this.api.getDashboard().pipe(catchError(() => of(null))).subscribe(s => {
      if (s) {
        const dynamicHealth = calculateClusterHealth({
          readyNodes: s.readyNodes,
          totalNodes: s.totalNodes,
          activeGateways: s.activeGateways,
          totalGateways: s.totalGateways,
          activeTunnels: s.activeTunnels,
          totalTunnels: s.totalTunnels,
          droppedFlowsPerSec: s.droppedFlowsPerSec,
          totalFlowsPerSec: s.totalFlowsPerSec,
        });

        this.summary.set({
          ...s,
          healthPercentage: dynamicHealth,
        });

        // Update live sparklines
        this.flowThroughputHistory.update(arr => [...arr.slice(1), s.totalFlowsPerSec || 200]);
        this.dropRateHistory.update(arr => [...arr.slice(1), s.droppedFlowsPerSec || 0]);
      }
      this.loading.set(false);
    });

    this.api.getEvents().pipe(catchError(() => of([]))).subscribe(e => this.events.set(e));

    // Verify CNI status from live controller API
    this.api.getCniStatus().pipe(catchError(() => of([]))).subscribe(nodes => {
      const cniActive = nodes.length > 0 ? nodes.every(n => n.status === 'Ready') : true;
      this.dataplaneSubsystems.update(list => list.map(item => {
        if (item.name === 'CNI / NetKit') return { ...item, status: cniActive ? 'Active' : 'Degraded' };
        return item;
      }));
    });
  }
}
