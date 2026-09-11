// Copyright 2026 straitgateway Authors
// SPDX-License-Identifier: Apache-2.0

import { Component, input, output } from '@angular/core';
import { CommonModule } from '@angular/common';
import { FormsModule } from '@angular/forms';

@Component({
  selector: 'sg-network-settings',
  standalone: true,
  imports: [CommonModule, FormsModule],
  template: `
    <div class="sg-settings-section">
      <div class="sg-section-header">
        <h3 class="sg-section-title">Network & Subnet Topology</h3>
        <p class="sg-section-subtitle">Pod subnets, Service subnets, MTU auto-discovery, and overlay tunnels</p>
      </div>

      <div class="sg-form-grid">
        <div class="sg-form-group">
          <label class="sg-form-label">Cluster PodCIDR</label>
          <input
            type="text"
            class="sg-input mono"
            [(ngModel)]="network().clusterPodCIDR"
            (ngModelChange)="changed.emit()"
            placeholder="10.244.0.0/16"
          />
          <span class="sg-form-hint">Overall Pod IP allocation range across all cluster nodes</span>
        </div>

        <div class="sg-form-group">
          <label class="sg-form-label">ServiceCIDR (VIPs)</label>
          <input
            type="text"
            class="sg-input mono"
            [(ngModel)]="network().serviceCIDR"
            (ngModelChange)="changed.emit()"
            placeholder="10.96.0.0/12"
          />
          <span class="sg-form-hint">Kubernetes ClusterIP service subnet range</span>
        </div>

        <div class="sg-form-group">
          <label class="sg-form-label">Tunnel Encapsulation Mode</label>
          <select
            class="sg-select"
            [(ngModel)]="network().tunnelMode"
            (ngModelChange)="changed.emit()"
          >
            <option value="wireguard">WireGuard (Encrypted Mesh)</option>
            <option value="vxlan">VXLAN (Standard Overlay)</option>
            <option value="geneve">Geneve (eBPF Metadata)</option>
            <option value="direct">Direct Routing (BGP / L2 Fastpath)</option>
          </select>
        </div>

        <div class="sg-form-group">
          <label class="sg-form-label">IPv6 Dual-Stack Support</label>
          <div style="display:flex;align-items:center;gap:10px;margin-top:8px">
            <input
              type="checkbox"
              id="enableIPv6"
              [(ngModel)]="network().enableIPv6"
              (ngModelChange)="changed.emit()"
              style="accent-color:var(--sg-accent);width:16px;height:16px"
            />
            <label for="enableIPv6" style="font-size:13px;color:var(--sg-text-2);cursor:pointer">
              Enable IPv6 forwarding and dual-stack translation (NAT64)
            </label>
          </div>
        </div>

        <div class="sg-form-group">
          <label class="sg-form-label">MTU Auto-Discovery</label>
          <div style="display:flex;align-items:center;gap:10px;margin-top:8px">
            <input
              type="checkbox"
              id="mtuDisc"
              [(ngModel)]="network().mtuDiscovery"
              (ngModelChange)="changed.emit()"
              style="accent-color:var(--sg-accent);width:16px;height:16px"
            />
            <label for="mtuDisc" style="font-size:13px;color:var(--sg-text-2);cursor:pointer">
              Automatically detect path MTU and avoid packet fragmentation
            </label>
          </div>
        </div>
      </div>
    </div>
  `,
})
export class NetworkSettingsComponent {
  network = input.required<any>();
  changed = output<void>();
}
