// Copyright 2026 straitgateway Authors
// SPDX-License-Identifier: Apache-2.0

import { Component, input, output } from '@angular/core';
import { CommonModule } from '@angular/common';
import { FormsModule } from '@angular/forms';

@Component({
  selector: 'sg-gateway-settings',
  standalone: true,
  imports: [CommonModule, FormsModule],
  template: `
    <div class="sg-settings-section">
      <div class="sg-section-header">
        <h3 class="sg-section-title">Gateway API & Ingress Settings</h3>
        <p class="sg-section-subtitle">Kubernetes Gateway API v1.6.1 integration and routing classes</p>
      </div>

      <div class="sg-form-grid">
        <div class="sg-form-group">
          <label class="sg-form-label">GatewayClass Controller Name</label>
          <input
            type="text"
            class="sg-input mono"
            [(ngModel)]="gateway().controllerName"
            (ngModelChange)="changed.emit()"
            placeholder="straitgateway.io/gateway-controller"
          />
          <span class="sg-form-hint">Matches GatewayClass.spec.controllerName</span>
        </div>

        <div class="sg-form-group">
          <label class="sg-form-label">Default GatewayClass Name</label>
          <input
            type="text"
            class="sg-input"
            [(ngModel)]="gateway().gatewayClassName"
            (ngModelChange)="changed.emit()"
            placeholder="straitgateway"
          />
        </div>

        <div class="sg-form-group">
          <label class="sg-form-label">Default HTTP Listener Port</label>
          <input
            type="number"
            class="sg-input"
            [(ngModel)]="gateway().defaultHttpPort"
            (ngModelChange)="changed.emit()"
          />
        </div>

        <div class="sg-form-group">
          <label class="sg-form-label">Default HTTPS Listener Port</label>
          <input
            type="number"
            class="sg-input"
            [(ngModel)]="gateway().defaultHttpsPort"
            (ngModelChange)="changed.emit()"
          />
        </div>

        <div class="sg-form-group">
          <label class="sg-form-label">BGP Route Synchronization</label>
          <div style="display:flex;align-items:center;gap:10px;margin-top:8px">
            <input
              type="checkbox"
              id="bgpSync"
              [(ngModel)]="gateway().enableBgpSync"
              (ngModelChange)="changed.emit()"
              style="accent-color:var(--sg-accent);width:16px;height:16px"
            />
            <label for="bgpSync" style="font-size:13px;color:var(--sg-text-2);cursor:pointer">
              Automatically advertise Gateway VIPs to external BGP peers
            </label>
          </div>
        </div>
      </div>
    </div>
  `,
})
export class GatewaySettingsComponent {
  gateway = input.required<any>();
  changed = output<void>();
}
