// Copyright 2026 straitgateway Authors
// SPDX-License-Identifier: Apache-2.0

import { Component, input, output } from '@angular/core';
import { CommonModule } from '@angular/common';
import { FormsModule } from '@angular/forms';

@Component({
  selector: 'sg-cni-settings',
  standalone: true,
  imports: [CommonModule, FormsModule],
  template: `
    <div class="sg-settings-section">
      <div class="sg-section-header">
        <h3 class="sg-section-title">CNI & Pod Networking</h3>
        <p class="sg-section-subtitle">StraitGateway CNI plugin, NetKit interfaces, and IPAM allocation</p>
      </div>

      <div class="sg-form-grid">
        <div class="sg-form-group">
          <label class="sg-form-label">CNI Binary Directory</label>
          <input
            type="text"
            class="sg-input mono"
            [(ngModel)]="cni().cniBinDir"
            (ngModelChange)="changed.emit()"
            placeholder="/opt/cni/bin"
          />
          <span class="sg-form-hint">Host directory where straitgateway-cni binary is placed</span>
        </div>

        <div class="sg-form-group">
          <label class="sg-form-label">CNI Config Directory</label>
          <input
            type="text"
            class="sg-input mono"
            [(ngModel)]="cni().cniConfDir"
            (ngModelChange)="changed.emit()"
            placeholder="/etc/cni/net.d"
          />
          <span class="sg-form-hint">Directory for CNI network configuration lists (.conflist)</span>
        </div>

        <div class="sg-form-group">
          <label class="sg-form-label">NetKit Interface Prefix</label>
          <input
            type="text"
            class="sg-input mono"
            [(ngModel)]="cni().interfacePrefix"
            (ngModelChange)="changed.emit()"
            placeholder="sg"
          />
          <span class="sg-form-hint">Prefix for pod virtual network interface pairs (e.g. sg-pod0)</span>
        </div>

        <div class="sg-form-group">
          <label class="sg-form-label">IPAM Backend Engine</label>
          <select
            class="sg-select"
            [(ngModel)]="cni().ipamBackend"
            (ngModelChange)="changed.emit()"
          >
            <option value="strait-ipam">strait-ipam (eBPF Native dynamic pool)</option>
            <option value="host-local">host-local (Static node-range)</option>
          </select>
        </div>
      </div>
    </div>
  `,
})
export class CniSettingsComponent {
  cni = input.required<any>();
  changed = output<void>();
}
