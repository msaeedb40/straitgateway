// Copyright 2026 straitgateway Authors
// SPDX-License-Identifier: Apache-2.0

import { Component, input, output } from '@angular/core';
import { CommonModule } from '@angular/common';
import { FormsModule } from '@angular/forms';

@Component({
  selector: 'sg-configuration-settings',
  standalone: true,
  imports: [CommonModule, FormsModule],
  template: `
    <div class="sg-settings-section">
      <div class="sg-section-header">
        <h3 class="sg-section-title">Controller & API Configuration</h3>
        <p class="sg-section-subtitle">StraitGateway controller endpoints and polling intervals</p>
      </div>

      <div class="sg-form-grid">
        <div class="sg-form-group">
          <label class="sg-form-label">Controller API Endpoint</label>
          <input
            type="text"
            class="sg-input"
            [(ngModel)]="config().apiBase"
            (ngModelChange)="changed.emit()"
            placeholder="http://straitgateway-controller:8080"
          />
          <span class="sg-form-hint">Direct gRPC-Web / REST endpoint for gateway controller</span>
        </div>

        <div class="sg-form-group">
          <label class="sg-form-label">Cluster Environment Name</label>
          <input
            type="text"
            class="sg-input"
            [(ngModel)]="config().clusterName"
            (ngModelChange)="changed.emit()"
            placeholder="production-us-east"
          />
          <span class="sg-form-hint">Unique identifier for this multi-cluster gateway instance</span>
        </div>

        <div class="sg-form-group">
          <label class="sg-form-label">UI Polling Refresh Interval (seconds)</label>
          <input
            type="number"
            class="sg-input"
            [(ngModel)]="config().refreshInterval"
            (ngModelChange)="changed.emit()"
            min="3"
            max="120"
          />
          <span class="sg-form-hint">Frequency for syncing telemetry flows, packets, and node statuses</span>
        </div>

        <div class="sg-form-group">
          <label class="sg-form-label">Mock Fallback Mode</label>
          <div style="display:flex;align-items:center;gap:10px;margin-top:8px">
            <input
              type="checkbox"
              id="mockFallback"
              [(ngModel)]="config().enableMockFallback"
              (ngModelChange)="changed.emit()"
              style="accent-color:var(--sg-accent);width:16px;height:16px"
            />
            <label for="mockFallback" style="font-size:13px;color:var(--sg-text-2);cursor:pointer">
              Simulate metrics and traffic flows if controller is unreachable
            </label>
          </div>
        </div>
      </div>
    </div>
  `,
})
export class ConfigurationSettingsComponent {
  config = input.required<any>();
  changed = output<void>();
}
