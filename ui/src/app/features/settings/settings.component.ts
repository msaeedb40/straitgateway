// Copyright 2026 straitgateway Authors — SPDX-License-Identifier: Apache-2.0
import { Component, inject, signal } from '@angular/core';
import { CommonModule } from '@angular/common';
import { FormsModule } from '@angular/forms';
import { RuntimeConfigService } from '../../core/config/runtime-config';
import { NotificationService } from '../../core/services/notification.service';
import { ConfigurationApi } from '../../core/api/resources/configuration.api';

interface SettingsSection {
  id: string;
  label: string;
  icon: string;
}

@Component({
  selector: 'app-settings',
  standalone: true,
  imports: [CommonModule, FormsModule],
  template: `
<div class="sg-fade-in">
  <div class="sg-page-header">
    <div>
      <h1 class="sg-page-title">Settings</h1>
      <p class="sg-page-subtitle">
        Configure StraitGateway controller, eBPF dataplane, and observability backends
      </p>
    </div>
    <button class="sg-btn" (click)="save()">
      Save Changes
    </button>
  </div>

  @if (saved()) {
    <div class="sg-alert-success" style="background:rgba(34,197,94,.15);border:1px solid rgba(34,197,94,.3);color:var(--sg-success);padding:14px 18px;border-radius:var(--sg-radius);font-size:13px;margin-bottom:20px">
      ✓ Settings updated and applied successfully
    </div>
  }

  <div class="sg-settings-layout">
    <!-- Sidebar navigation -->
    <nav class="sg-settings-nav">
      @for (s of sections; track s.id) {
        <button
          class="sg-settings-nav-item"
          [class.active]="activeSection() === s.id"
          (click)="activeSection.set(s.id)"
        >
          <span [innerHTML]="s.icon" style="display:flex;align-items:center;width:16px;height:16px"></span>
          <span>{{ s.label }}</span>
        </button>
      }
    </nav>

    <!-- Content -->
    <div class="sg-settings-content">
      <!-- General -->
      @if (activeSection() === 'general') {
        <div class="sg-card">
          <div class="sg-card-header">
            <span class="sg-card-title">General Settings</span>
          </div>
          <div class="sg-card-body">
            <div class="sg-form-row">
              <label class="sg-form-label">Cluster Name</label>
              <input class="sg-input" [(ngModel)]="cfg.clusterName">
              <span class="sg-form-hint">Identifies this Kubernetes cluster in multi-cluster transit topologies.</span>
            </div>
            <div class="sg-form-row">
              <label class="sg-form-label">Default Namespace</label>
              <input class="sg-input" [(ngModel)]="cfg.defaultNamespace">
              <span class="sg-form-hint">Initial namespace selected on dashboard startup.</span>
            </div>
            <div class="sg-form-row" style="margin-bottom:0">
              <label class="sg-form-label">Refresh Interval (ms)</label>
              <input type="number" class="sg-input" [(ngModel)]="cfg.refreshIntervalMs">
              <span class="sg-form-hint">Polling frequency for flows and packet telemetry.</span>
            </div>
          </div>
        </div>
      }

      <!-- Configuration (CRD Apply) -->
      @if (activeSection() === 'configuration') {
        <div class="sg-card">
          <div class="sg-card-header">
            <span class="sg-card-title">CRD Configuration Apply</span>
          </div>
          <div class="sg-card-body" style="display:flex;flex-direction:column;gap:16px">
            <p style="font-size:13px;color:var(--sg-text-2)">Paste YAML manifests for StraitNetworkPolicy, Gateway, or TransitGateway to apply directly to the cluster.</p>
            <textarea
              [(ngModel)]="crdYaml"
              rows="12"
              class="sg-yaml-code"
              style="width:100%;outline:none;resize:vertical"
            ></textarea>
            <button (click)="applyCRD()" class="sg-btn" style="align-self:flex-start">
              Apply Manifest
            </button>
          </div>
        </div>
      }

      <!-- Gateway -->
      @if (activeSection() === 'gateway') {
        <div class="sg-card">
          <div class="sg-card-header">
            <span class="sg-card-title">Gateway API Settings</span>
          </div>
          <div class="sg-card-body">
            <div class="sg-form-row">
              <label class="sg-form-label">GatewayClass Controller Name</label>
              <input class="sg-input" [(ngModel)]="cfg.gatewayControllerName">
            </div>
            <div class="sg-form-row" style="margin-bottom:0">
              <label class="sg-form-label">Default HTTP Listen Port</label>
              <input type="number" class="sg-input" [(ngModel)]="cfg.defaultHttpPort">
            </div>
          </div>
        </div>
      }

      <!-- Network -->
      @if (activeSection() === 'network') {
        <div class="sg-card">
          <div class="sg-card-header">
            <span class="sg-card-title">Network Configuration</span>
          </div>
          <div class="sg-card-body">
            <div class="sg-form-row">
              <label class="sg-form-label">Pod MTU</label>
              <input type="number" class="sg-input" [(ngModel)]="cfg.mtu">
              <span class="sg-form-hint">Default MTU for NetKit interfaces (e.g. 1500 or 1420 with WireGuard).</span>
            </div>
            <div class="sg-form-row" style="margin-bottom:0">
              <label class="sg-form-label">Pod Network CIDR</label>
              <input class="sg-input" [(ngModel)]="cfg.podCidr">
            </div>
          </div>
        </div>
      }

      <!-- eBPF -->
      @if (activeSection() === 'ebpf') {
        <div class="sg-card">
          <div class="sg-card-header">
            <span class="sg-card-title">eBPF Dataplane Limits</span>
          </div>
          <div class="sg-card-body">
            <div class="sg-form-row">
              <label class="sg-form-label">Flow Table Capacity (LRU Hash)</label>
              <input type="number" class="sg-input" [(ngModel)]="cfg.flowTableMax">
            </div>
            <div class="sg-form-row">
              <label class="sg-form-label">Maglev Table Size (M)</label>
              <input type="number" class="sg-input" [(ngModel)]="cfg.maglevM">
              <span class="sg-form-hint">Prime number for consistent hashing (default: 65537).</span>
            </div>
            <div class="sg-form-row" style="margin-bottom:0">
              <label class="sg-form-label">BPFFS Mount Path</label>
              <input class="sg-input" [(ngModel)]="cfg.bpffsPath">
            </div>
          </div>
        </div>
      }

      <!-- CNI -->
      @if (activeSection() === 'cni') {
        <div class="sg-card">
          <div class="sg-card-header">
            <span class="sg-card-title">CNI & IPAM Plugin</span>
          </div>
          <div class="sg-card-body">
            <div class="sg-form-row">
              <label class="sg-form-label">CNI Binary Directory</label>
              <input class="sg-input" [(ngModel)]="cfg.cniBinDir">
            </div>
            <div class="sg-form-row" style="margin-bottom:0">
              <label class="sg-form-label">CNI Config Path</label>
              <input class="sg-input" [(ngModel)]="cfg.cniConfPath">
            </div>
          </div>
        </div>
      }

      <!-- Observability -->
      @if (activeSection() === 'observability') {
        <div class="sg-card">
          <div class="sg-card-header">
            <span class="sg-card-title">Observability Backends</span>
          </div>
          <div class="sg-card-body">
            <div class="sg-form-row">
              <label class="sg-form-label">Prometheus Base URL</label>
              <input class="sg-input" [(ngModel)]="cfg.prometheusBase">
            </div>
            <div class="sg-form-row">
              <label class="sg-form-label">Grafana Base URL</label>
              <input class="sg-input" [(ngModel)]="cfg.grafanaBase">
            </div>
            <div class="sg-form-row" style="margin-bottom:0">
              <label class="sg-form-label">Jaeger Base URL</label>
              <input class="sg-input" [(ngModel)]="cfg.jaegerBase">
            </div>
          </div>
        </div>
      }

      <!-- Security -->
      @if (activeSection() === 'security') {
        <div class="sg-card">
          <div class="sg-card-header">
            <span class="sg-card-title">Security & Encryption</span>
          </div>
          <div class="sg-card-body">
            <div class="sg-form-row">
              <label class="sg-form-label">WireGuard Transit Port</label>
              <input type="number" class="sg-input" [(ngModel)]="cfg.wireguardPort">
            </div>
            <div class="sg-form-row" style="margin-bottom:0">
              <label class="sg-form-label">mTLS Certificate Secret</label>
              <input class="sg-input" [(ngModel)]="cfg.tlsSecret">
            </div>
          </div>
        </div>
      }

      <!-- Advanced -->
      @if (activeSection() === 'advanced') {
        <div class="sg-card">
          <div class="sg-card-header">
            <span class="sg-card-title">Advanced Diagnostics</span>
          </div>
          <div class="sg-card-body" style="display:flex;flex-direction:column;gap:18px">
            <div style="display:flex;align-items:center;justify-content:space-between;padding:14px 18px;background:var(--sg-surface-2);border-radius:var(--sg-radius-sm);border:1px solid var(--sg-border)">
              <div>
                <div style="font-size:13px;font-weight:500;color:var(--sg-text)">Offline Standalone Mock Fallback</div>
                <div style="font-size:12px;color:var(--sg-text-3);margin-top:2px">Provide realistic mock data when controller is unreachable</div>
              </div>
              <input type="checkbox" [(ngModel)]="cfg.mockMode" style="width:18px;height:18px;accent-color:var(--sg-accent);cursor:pointer">
            </div>
            <div style="display:flex;gap:12px">
              <button (click)="exportDiagnostics()" class="sg-btn sg-btn-secondary">
                Export System Diagnostics
              </button>
            </div>
          </div>
        </div>
      }
    </div>
  </div>
</div>
  `,
})
export class SettingsComponent {
  private runtimeCfg = inject(RuntimeConfigService);
  private notif = inject(NotificationService);
  private configApi = inject(ConfigurationApi);

  activeSection = signal('general');
  saved = signal(false);

  crdYaml = `apiVersion: straitgateway.io/v1alpha1
kind: StraitNetworkPolicy
metadata:
  name: allow-cluster-traffic
  namespace: default
spec:
  policyType: Both
  ingress:
    - action: Allow
      from:
        - podSelector:
            matchLabels:
              app: core-api
  egress:
    - action: Allow
      to:
        - ipBlock:
            cidr: 10.250.0.0/16
`;

  cfg = {
    clusterName: 'dev',
    defaultNamespace: 'default',
    refreshIntervalMs: 15000,
    gatewayControllerName: 'straitgateway.io/gateway-controller',
    defaultHttpPort: 80,
    mtu: 1500,
    podCidr: '10.244.0.0/16',
    flowTableMax: 262144,
    maglevM: 65537,
    bpffsPath: '/sys/fs/bpf/straitgateway',
    cniBinDir: '/opt/cni/bin',
    cniConfPath: '/etc/cni/net.d/10-straitgateway.conflist',
    prometheusBase: 'http://localhost:9090',
    grafanaBase: 'http://localhost:3000',
    jaegerBase: 'http://localhost:16686',
    wireguardPort: 51820,
    tlsSecret: 'sg-controller-mtls',
    mockMode: true,
  };

  sections: SettingsSection[] = [
    { id: 'general',       label: 'General',       icon: '<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><circle cx="12" cy="12" r="3"/><path d="M19.4 15a1.65 1.65 0 0 0 .33 1.82l.06.06a2 2 0 0 1 0 2.83 2 2 0 0 1-2.83 0l-.06-.06a1.65 1.65 0 0 0-1.82-.33 1.65 1.65 0 0 0-1 1.51V21a2 2 0 0 1-2 2 2 2 0 0 1-2-2v-.09A1.65 1.65 0 0 0 9 19.4a1.65 1.65 0 0 0-1.82.33l-.06.06a2 2 0 0 1-2.83 0 2 2 0 0 1 0-2.83l.06-.06a1.65 1.65 0 0 0 .33-1.82 1.65 1.65 0 0 0-1.51-1H3a2 2 0 0 1-2-2 2 2 0 0 1 2-2h.09A1.65 1.65 0 0 0 4.6 9a1.65 1.65 0 0 0-.33-1.82l-.06-.06a2 2 0 0 1 0-2.83 2 2 0 0 1 2.83 0l.06.06a1.65 1.65 0 0 0 1.82.33H9a1.65 1.65 0 0 0 1-1.51V3a2 2 0 0 1 2-2 2 2 0 0 1 2 2v.09a1.65 1.65 0 0 0 1 1.51 1.65 1.65 0 0 0 1.82-.33l.06-.06a2 2 0 0 1 2.83 0 2 2 0 0 1 0 2.83l-.06.06a1.65 1.65 0 0 0-.33 1.82V9a1.65 1.65 0 0 0 1.51 1H21a2 2 0 0 1 2 2 2 2 0 0 1-2 2h-.09a1.65 1.65 0 0 0-1.51 1z"/></svg>' },
    { id: 'configuration', label: 'Configuration', icon: '<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M14 2H6a2 2 0 0 0-2 2v16a2 2 0 0 0 2 2h12a2 2 0 0 0 2-2V8z"/><polyline points="14 2 14 8 20 8"/><line x1="16" y1="13" x2="8" y2="13"/><line x1="16" y1="17" x2="8" y2="17"/></svg>' },
    { id: 'gateway',       label: 'Gateway API',   icon: '<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="m18 8 4 4-4 4"/><path d="M2 12h20"/><path d="m6 16-4-4 4-4"/></svg>' },
    { id: 'network',       label: 'Network',       icon: '<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><circle cx="12" cy="12" r="10"/><line x1="2" y1="12" x2="22" y2="12"/><path d="M12 2a15.3 15.3 0 0 1 4 10 15.3 15.3 0 0 1-4 10 15.3 15.3 0 0 1-4-10 15.3 15.3 0 0 1 4-10z"/></svg>' },
    { id: 'ebpf',          label: 'eBPF',          icon: '<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><polygon points="13 2 3 14 12 14 11 22 21 10 12 10 13 2"/></svg>' },
    { id: 'cni',           label: 'CNI',           icon: '<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><rect width="18" height="18" x="3" y="3" rx="2"/><path d="M3 9h18"/><path d="M9 21V9"/></svg>' },
    { id: 'observability', label: 'Observability', icon: '<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><polyline points="22 12 18 12 15 21 9 3 6 12 2 12"/></svg>' },
    { id: 'security',      label: 'Security',      icon: '<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><rect width="18" height="11" x="3" y="11" rx="2" ry="2"/><path d="M7 11V7a5 5 0 0 1 10 0v4"/></svg>' },
    { id: 'advanced',      label: 'Advanced',      icon: '<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><polyline points="4 17 10 11 4 5"/><line x1="12" y1="19" x2="20" y2="19"/></svg>' },
  ];

  save() {
    this.runtimeCfg.update({
      clusterName: this.cfg.clusterName,
      refreshIntervalMs: this.cfg.refreshIntervalMs,
      prometheusBase: this.cfg.prometheusBase,
      grafanaBase: this.cfg.grafanaBase,
      jaegerBase: this.cfg.jaegerBase,
      mockMode: this.cfg.mockMode,
    });
    this.saved.set(true);
    this.notif.success('Settings Saved', 'Configuration successfully synchronized');
    setTimeout(() => this.saved.set(false), 3000);
  }

  applyCRD() {
    this.configApi.applyYaml(this.crdYaml).subscribe((res) => {
      this.notif.success('Manifest Applied', res.message);
    });
  }

  exportDiagnostics() {
    const dump = {
      timestamp: new Date().toISOString(),
      config: this.cfg,
    };
    const blob = new Blob([JSON.stringify(dump, null, 2)], { type: 'application/json' });
    const url = URL.createObjectURL(blob);
    const a = document.createElement('a');
    a.href = url;
    a.download = `straitgateway-diagnostics-${Date.now()}.json`;
    a.click();
    this.notif.info('Diagnostics Exported', 'Downloaded system diagnostics report');
  }
}
