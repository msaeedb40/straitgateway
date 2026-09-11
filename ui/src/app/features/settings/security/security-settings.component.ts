// Copyright 2026 straitgateway Authors
// SPDX-License-Identifier: Apache-2.0

import { Component, input, output } from '@angular/core';
import { CommonModule } from '@angular/common';
import { FormsModule } from '@angular/forms';

@Component({
  selector: 'sg-security-settings',
  standalone: true,
  imports: [CommonModule, FormsModule],
  template: `
    <div class="sg-settings-section">
      <div class="sg-section-header">
        <h3 class="sg-section-title">Security & Encryption</h3>
        <p class="sg-section-subtitle">eBPF NetworkPolicy enforcement, WireGuard key rotation, and mTLS</p>
      </div>

      <div class="sg-form-grid">
        <div class="sg-form-group">
          <label class="sg-form-label">NetworkPolicy Enforcement</label>
          <div style="display:flex;align-items:center;gap:10px;margin-top:8px">
            <input
              type="checkbox"
              id="enforcePolicy"
              [(ngModel)]="security().enforceNetworkPolicy"
              (ngModelChange)="changed.emit()"
              style="accent-color:var(--sg-accent);width:16px;height:16px"
            />
            <label for="enforcePolicy" style="font-size:13px;color:var(--sg-text-2);cursor:pointer">
              Enforce Kubernetes NetworkPolicy and StraitNetworkPolicy in eBPF dataplane
            </label>
          </div>
        </div>

        <div class="sg-form-group">
          <label class="sg-form-label">Default Policy Action for Unmatched Flows</label>
          <select
            class="sg-select"
            [(ngModel)]="security().defaultPolicyAction"
            (ngModelChange)="changed.emit()"
          >
            <option value="allow">Allow (Permissive / Default Kubernetes behavior)</option>
            <option value="deny">Deny (Zero-Trust Isolation)</option>
          </select>
        </div>

        <div class="sg-form-group">
          <label class="sg-form-label">WireGuard Key Rotation Period (days)</label>
          <input
            type="number"
            class="sg-input"
            [(ngModel)]="security().wireguardKeyRotationDays"
            (ngModelChange)="changed.emit()"
            min="1"
            max="365"
          />
          <span class="sg-form-hint">Automatic re-keying schedule for inter-cluster tunnels</span>
        </div>

        <div class="sg-form-group">
          <label class="sg-form-label">Strict mTLS Integration</label>
          <div style="display:flex;align-items:center;gap:10px;margin-top:8px">
            <input
              type="checkbox"
              id="mtlsStrict"
              [(ngModel)]="security().mTLSStrict"
              (ngModelChange)="changed.emit()"
              style="accent-color:var(--sg-accent);width:16px;height:16px"
            />
            <label for="mtlsStrict" style="font-size:13px;color:var(--sg-text-2);cursor:pointer">
              Verify client certificates using cluster Kubernetes CA root
            </label>
          </div>
        </div>
      </div>
    </div>
  `,
})
export class SecuritySettingsComponent {
  security = input.required<any>();
  changed = output<void>();
}
