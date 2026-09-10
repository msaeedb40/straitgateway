// Copyright 2026 straitgateway Authors
// SPDX-License-Identifier: Apache-2.0

import { Component, inject, signal, OnInit } from '@angular/core';
import { CommonModule } from '@angular/common';
import { ApiClient, NodeStatus } from '../../core/api/api-client';
import { catchError, of } from 'rxjs';

import { CountPipe } from '../../shared';

@Component({
  selector: 'app-nodes',
  standalone: true,
  imports: [CommonModule, CountPipe],
  template: `
<div class="sg-fade-in">
  <div class="sg-page-header">
    <div>
      <h1 class="sg-page-title">Nodes</h1>
      <p class="sg-page-subtitle">straitgatewayd agent status per node</p>
    </div>
    <button class="sg-btn sg-btn-secondary" (click)="load()">Refresh</button>
  </div>

  <!-- Dataplane readiness grid -->
  @if (nodes().length) {
    <div class="sg-stat-grid">
      <div class="sg-stat-card">
        <div class="sg-stat-label">Total Nodes</div>
        <div class="sg-stat-value text-accent">{{ nodes().length }}</div>
      </div>
      <div class="sg-stat-card">
        <div class="sg-stat-label">CNI Ready</div>
        <div class="sg-stat-value text-success">{{ nodes() | count:'cniReady' }}</div>
      </div>
      <div class="sg-stat-card">
        <div class="sg-stat-label">Service Ready</div>
        <div class="sg-stat-value text-success">{{ nodes() | count:'serviceReady' }}</div>
      </div>
      <div class="sg-stat-card">
        <div class="sg-stat-label">kube-proxy Replaced</div>
        <div class="sg-stat-value text-success">{{ nodes() | count:'kubeProxyReplacement' }}</div>
      </div>
    </div>
  }

  <div class="sg-card">
    <div class="sg-card-header">
      <span class="sg-card-title">Node Agent Status</span>
      <span class="text-muted text-sm">Each subsystem is independently tracked</span>
    </div>
    <div class="sg-table-wrap">
      <table class="sg-table">
        <thead>
          <tr>
            <th>Node</th>
            <th>IP</th>
            <th>Pod CIDR</th>
            <th>Kernel</th>
            <th>CNI</th>
            <th>Service</th>
            <th>Policy</th>
            <th>Gateway</th>
            <th>kube-proxy</th>
            <th>BPF Rev</th>
          </tr>
        </thead>
        <tbody>
          @for (node of nodes(); track node.name) {
            <tr>
              <td class="mono">{{ node.name }}</td>
              <td class="mono">{{ node.ip }}</td>
              <td class="mono">{{ node.podCIDR }}</td>
              <td class="mono text-muted">{{ node.kernelVersion }}</td>
              <td><span class="sg-badge" [class]="node.cniReady ? 'active' : 'error'">{{ node.cniReady ? '✓' : '✗' }}</span></td>
              <td><span class="sg-badge" [class]="node.serviceReady ? 'active' : 'error'">{{ node.serviceReady ? '✓' : '✗' }}</span></td>
              <td><span class="sg-badge" [class]="node.policyReady ? 'active' : 'error'">{{ node.policyReady ? '✓' : '✗' }}</span></td>
              <td><span class="sg-badge" [class]="node.gatewayReady ? 'active' : 'error'">{{ node.gatewayReady ? '✓' : '✗' }}</span></td>
              <td><span class="sg-badge" [class]="node.kubeProxyReplacement ? 'active' : 'warn'">{{ node.kubeProxyReplacement ? 'eBPF' : 'iptables' }}</span></td>
              <td class="mono text-muted">{{ node.bpfMapRevision }}</td>
            </tr>
          } @empty {
            <tr><td colspan="10">
              <div class="sg-empty">
                <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5"><rect x="3" y="3" width="7" height="7"/><rect x="14" y="3" width="7" height="7"/></svg>
                <p>No nodes found. Ensure straitgatewayd DaemonSet is running.</p>
              </div>
            </td></tr>
          }
        </tbody>
      </table>
    </div>
  </div>
</div>
`,
})
export class NodesComponent implements OnInit {
  private api = inject(ApiClient);
  nodes = signal<NodeStatus[]>([]);
  ngOnInit() { this.load(); }
  load() { this.api.getNodes().pipe(catchError(() => of([]))).subscribe(n => this.nodes.set(n)); }
}
