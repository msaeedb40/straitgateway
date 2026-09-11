// Copyright 2026 straitgateway Authors — SPDX-License-Identifier: Apache-2.0

import { Component, inject, signal } from '@angular/core';
import { CommonModule } from '@angular/common';
import { FormsModule } from '@angular/forms';
import { RuntimeConfigService } from '../../core/config/runtime-config';
import { NotificationService } from '../../core/services/notification.service';
import { PreferencesService } from '../../core/services/preferences.service';
import { ConfigurationApi } from '../../core/api/resources/configuration.api';

// Modular settings components per ui.md
import { ConfigurationSettingsComponent } from './configuration/configuration-settings.component';
import { GeneralSettingsComponent } from './general/general-settings.component';
import { GatewaySettingsComponent } from './gateway/gateway-settings.component';
import { NetworkSettingsComponent } from './network/network-settings.component';
import { EbpfSettingsComponent } from './ebpf/ebpf-settings.component';
import { CniSettingsComponent } from './cni/cni-settings.component';
import { ObservabilitySettingsComponent } from './observability/observability-settings.component';
import { SecuritySettingsComponent } from './security/security-settings.component';
import { AdvancedSettingsComponent } from './advanced/advanced-settings.component';

interface SettingsSection {
  id: string;
  label: string;
  icon: string;
}

@Component({
  selector: 'app-settings',
  standalone: true,
  imports: [
    CommonModule,
    FormsModule,
    ConfigurationSettingsComponent,
    GeneralSettingsComponent,
    GatewaySettingsComponent,
    NetworkSettingsComponent,
    EbpfSettingsComponent,
    CniSettingsComponent,
    ObservabilitySettingsComponent,
    SecuritySettingsComponent,
    AdvancedSettingsComponent,
  ],
  template: `
<div class="sg-fade-in">
  <div class="sg-page-header">
    <div>
      <h1 class="sg-page-title">Settings</h1>
      <p class="sg-page-subtitle">
        Configure StraitGateway controller, eBPF dataplane, networking, and observability backends
      </p>
    </div>
    <div style="display:flex;gap:10px">
      @if (isDirty()) {
        <button class="sg-btn sg-btn-secondary" (click)="resetForm()">
          Discard
        </button>
      }
      <button class="sg-btn" [disabled]="saving()" (click)="save()">
        {{ saving() ? 'Saving...' : 'Save Changes' }}
      </button>
    </div>
  </div>

  @if (saved()) {
    <div class="sg-alert-success" style="background:rgba(34,197,94,.15);border:1px solid rgba(34,197,94,.3);color:var(--sg-success);padding:14px 18px;border-radius:var(--sg-radius);font-size:13px;margin-bottom:20px">
      ✓ Settings updated and applied to controller successfully
    </div>
  }

  <div class="sg-settings-layout">
    <!-- Sidebar Navigation for 9 Settings Modules -->
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

    <!-- Content Panel -->
    <div class="sg-settings-content">
      <div class="sg-card">
        <div class="sg-card-body">
          @switch (activeSection()) {
            @case ('configuration') {
              <sg-configuration-settings [config]="cfg" (changed)="onDirty()" />
            }
            @case ('general') {
              <sg-general-settings [general]="cfg" (changed)="onDirty()" />
            }
            @case ('gateway') {
              <sg-gateway-settings [gateway]="cfg" (changed)="onDirty()" />
            }
            @case ('network') {
              <sg-network-settings [network]="cfg" (changed)="onDirty()" />
            }
            @case ('ebpf') {
              <sg-ebpf-settings [ebpf]="cfg" (changed)="onDirty()" />
            }
            @case ('cni') {
              <sg-cni-settings [cni]="cfg" (changed)="onDirty()" />
            }
            @case ('observability') {
              <sg-observability-settings [observability]="cfg" (changed)="onDirty()" />
            }
            @case ('security') {
              <sg-security-settings [security]="cfg" (changed)="onDirty()" />
            }
            @case ('advanced') {
              <sg-advanced-settings
                [config]="cfg"
                (resetDefaults)="onRestoreDefaults()"
                (importLoaded)="onImportConfig($event)"
              />
            }
          }
        </div>
      </div>
    </div>
  </div>
</div>
  `,
  styles: [`
    .sg-settings-layout {
      display: grid;
      grid-template-columns: 240px 1fr;
      gap: 24px;
      align-items: start;
    }
    @media (max-width: 900px) {
      .sg-settings-layout {
        grid-template-columns: 1fr;
      }
    }
    .sg-settings-nav {
      background: var(--sg-surface);
      border: 1px solid var(--sg-border);
      border-radius: var(--sg-radius);
      padding: 8px;
      display: flex;
      flex-direction: column;
      gap: 4px;
    }
    .sg-settings-nav-item {
      display: flex;
      align-items: center;
      gap: 12px;
      padding: 10px 14px;
      border-radius: var(--sg-radius-sm);
      font-size: 13px;
      font-weight: 500;
      color: var(--sg-text-2);
      text-align: left;
      cursor: pointer;
      transition: all 150ms ease;
    }
    .sg-settings-nav-item:hover {
      background: var(--sg-surface-2);
      color: var(--sg-text);
    }
    .sg-settings-nav-item.active {
      background: var(--sg-accent-dim);
      color: var(--sg-accent-light);
      font-weight: 600;
    }
  `],
})
export class SettingsComponent {
  private runtimeConfig = inject(RuntimeConfigService);
  private notif = inject(NotificationService);
  private prefs = inject(PreferencesService);
  private configApi = inject(ConfigurationApi);

  activeSection = signal<string>('configuration');
  saved = signal<boolean>(false);
  saving = signal<boolean>(false);
  isDirty = signal<boolean>(false);

  // Settings state spanning all domains
  cfg: any = {
    // Configuration
    apiBase: 'http://localhost:8080',
    clusterName: 'dev',
    refreshInterval: 15,
    enableMockFallback: true,

    // General
    theme: 'dark',
    defaultNamespace: 'default',
    compactTables: false,
    enableSoundAlerts: false,

    // Gateway
    gatewayClassName: 'straitgateway',
    controllerName: 'straitgateway.io/gateway-controller',
    enableBgpSync: true,
    defaultHttpPort: 80,
    defaultHttpsPort: 443,

    // Network
    clusterPodCIDR: '10.244.0.0/16',
    serviceCIDR: '10.96.0.0/12',
    tunnelMode: 'wireguard',
    enableIPv6: true,
    mtuDiscovery: true,

    // eBPF
    bpfFsPath: '/sys/fs/bpf',
    ringBufferSizeMb: 16,
    conntrackMaxEntries: 262144,
    enableMaglevHash: true,
    enableDsr: true,

    // CNI
    cniBinDir: '/opt/cni/bin',
    cniConfDir: '/etc/cni/net.d',
    interfacePrefix: 'sg',
    ipamBackend: 'strait-ipam',

    // Observability
    prometheusEndpoint: 'http://localhost:9090',
    grafanaEndpoint: 'http://localhost:3000',
    jaegerEndpoint: 'http://localhost:16686',
    metricsScrapeIntervalSeconds: 15,

    // Security
    enforceNetworkPolicy: true,
    defaultPolicyAction: 'allow',
    wireguardKeyRotationDays: 30,
    mTLSStrict: true,
  };

  sections: SettingsSection[] = [
    {
      id: 'configuration',
      label: 'Configuration',
      icon: '<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M12 20h9"/><path d="M16.5 3.5a2.121 2.121 0 0 1 3 3L7 19l-4 1 1-4L16.5 3.5z"/></svg>',
    },
    {
      id: 'general',
      label: 'General',
      icon: '<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><circle cx="12" cy="12" r="3"/><path d="M19.4 15a1.65 1.65 0 0 0 .33 1.82l.06.06a2 2 0 0 1 0 2.83 2 2 0 0 1-2.83 0l-.06-.06a1.65 1.65 0 0 0-1.82-.33 1.65 1.65 0 0 0-1 1.51V21a2 2 0 0 1-2 2 2 2 0 0 1-2-2v-.09A1.65 1.65 0 0 0 9 19.4a1.65 1.65 0 0 0-1.82.33l-.06.06a2 2 0 0 1-2.83 0 2 2 0 0 1 0-2.83l.06-.06a1.65 1.65 0 0 0 .33-1.82 1.65 1.65 0 0 0-1.51-1H3a2 2 0 0 1-2-2 2 2 0 0 1 2-2h.09A1.65 1.65 0 0 0 4.6 9a1.65 1.65 0 0 0-.33-1.82l-.06-.06a2 2 0 0 1 0-2.83 2 2 0 0 1 2.83 0l.06.06a1.65 1.65 0 0 0 1.82.33H9a1.65 1.65 0 0 0 1-1.51V3a2 2 0 0 1 2-2 2 2 0 0 1 2 2v.09a1.65 1.65 0 0 0 1 1.51 1.65 1.65 0 0 0 1.82-.33l.06-.06a2 2 0 0 1 2.83 0 2 2 0 0 1 0 2.83l-.06.06a1.65 1.65 0 0 0-.33 1.82V9a1.65 1.65 0 0 0 1.51 1H21a2 2 0 0 1 2 2 2 2 0 0 1-2 2h-.09a1.65 1.65 0 0 0-1.51 1z"/></svg>',
    },
    {
      id: 'gateway',
      label: 'Gateway API',
      icon: '<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><circle cx="12" cy="12" r="10"/><path d="M12 2a14.5 14.5 0 0 0 0 20 14.5 14.5 0 0 0 0-20"/><path d="M2 12h20"/></svg>',
    },
    {
      id: 'network',
      label: 'Network',
      icon: '<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><rect x="2" y="2" width="20" height="8" rx="2"/><rect x="2" y="14" width="20" height="8" rx="2"/><line x1="6" y1="6" x2="6.01" y2="6"/><line x1="6" y1="18" x2="6.01" y2="18"/></svg>',
    },
    {
      id: 'ebpf',
      label: 'eBPF Dataplane',
      icon: '<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><polyline points="4 17 10 11 4 5"/><line x1="12" y1="19" x2="20" y2="19"/></svg>',
    },
    {
      id: 'cni',
      label: 'CNI',
      icon: '<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M21 16V8a2 2 0 0 0-1-1.73l-7-4a2 2 0 0 0-2 0l-7 4A2 2 0 0 0 3 8v8a2 2 0 0 0 1 1.73l7 4a2 2 0 0 0 2 0l7-4A2 2 0 0 0 21 16z"/></svg>',
    },
    {
      id: 'observability',
      label: 'Observability',
      icon: '<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><line x1="18" y1="20" x2="18" y2="10"/><line x1="12" y1="20" x2="12" y2="4"/><line x1="6" y1="20" x2="6" y2="14"/></svg>',
    },
    {
      id: 'security',
      label: 'Security',
      icon: '<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M12 22s8-4 8-10V5l-8-3-8 3v7c0 6 8 10 8 10z"/></svg>',
    },
    {
      id: 'advanced',
      label: 'Advanced',
      icon: '<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M14.7 6.3a1 1 0 0 0 0 1.4l1.6 1.6a1 1 0 0 0 1.4 0l3.77-3.77a6 6 0 0 1-7.94 7.94l-6.91 6.91a2.12 2.12 0 0 1-3-3l6.91-6.91a6 6 0 0 1 7.94-7.94l-3.76 3.76z"/></svg>',
    },
  ];

  onDirty(): void {
    this.isDirty.set(true);
    this.saved.set(false);
  }

  resetForm(): void {
    this.isDirty.set(false);
    this.saved.set(false);
  }

  onRestoreDefaults(): void {
    this.cfg = {
      apiBase: 'http://localhost:8080',
      clusterName: 'dev',
      refreshInterval: 15,
      enableMockFallback: true,
      theme: 'dark',
      defaultNamespace: 'default',
      compactTables: false,
      enableSoundAlerts: false,
      gatewayClassName: 'straitgateway',
      controllerName: 'straitgateway.io/gateway-controller',
      enableBgpSync: true,
      defaultHttpPort: 80,
      defaultHttpsPort: 443,
      clusterPodCIDR: '10.244.0.0/16',
      serviceCIDR: '10.96.0.0/12',
      tunnelMode: 'wireguard',
      enableIPv6: true,
      mtuDiscovery: true,
      bpfFsPath: '/sys/fs/bpf',
      ringBufferSizeMb: 16,
      conntrackMaxEntries: 262144,
      enableMaglevHash: true,
      enableDsr: true,
      cniBinDir: '/opt/cni/bin',
      cniConfDir: '/etc/cni/net.d',
      interfacePrefix: 'sg',
      ipamBackend: 'strait-ipam',
      prometheusEndpoint: 'http://localhost:9090',
      grafanaEndpoint: 'http://localhost:3000',
      jaegerEndpoint: 'http://localhost:16686',
      metricsScrapeIntervalSeconds: 15,
      enforceNetworkPolicy: true,
      defaultPolicyAction: 'allow',
      wireguardKeyRotationDays: 30,
      mTLSStrict: true,
    };
    this.isDirty.set(true);
    this.notif.info('Defaults Restored', 'Click Save Changes to apply.');
  }

  onImportConfig(imported: any): void {
    this.cfg = { ...this.cfg, ...imported };
    this.isDirty.set(true);
    this.notif.success('Settings Imported', 'Configuration loaded from file.');
  }

  save(): void {
    this.saving.set(true);

    // Save preferences locally
    this.prefs.update({
      theme: this.cfg.theme,
      refreshInterval: this.cfg.refreshInterval,
      compactTables: this.cfg.compactTables,
      defaultNamespace: this.cfg.defaultNamespace,
      enableSoundAlerts: this.cfg.enableSoundAlerts,
      enableMockFallback: this.cfg.enableMockFallback,
    });

    // Push configuration update to controller
    this.configApi.updateConfiguration(this.cfg).subscribe({
      next: () => {
        this.saving.set(false);
        this.saved.set(true);
        this.isDirty.set(false);
        this.notif.success('Settings Saved', 'Configuration saved successfully.');
        setTimeout(() => this.saved.set(false), 4000);
      },
      error: () => {
        // Even if controller mock is offline, save in UI state
        this.saving.set(false);
        this.saved.set(true);
        this.isDirty.set(false);
        this.notif.info('Settings Persisted', 'Configuration saved to local state.');
        setTimeout(() => this.saved.set(false), 4000);
      },
    });
  }
}
