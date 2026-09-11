// Copyright 2026 straitgateway Authors
// SPDX-License-Identifier: Apache-2.0

import { Component, input, output } from '@angular/core';
import { CommonModule } from '@angular/common';
import { FormsModule } from '@angular/forms';

@Component({
  selector: 'sg-observability-settings',
  standalone: true,
  imports: [CommonModule, FormsModule],
  template: `
    <div class="sg-settings-section">
      <div class="sg-section-header">
        <h3 class="sg-section-title">Observability & Telemetry Backends</h3>
        <p class="sg-section-subtitle">Prometheus, Grafana, and Jaeger distributed tracing integrations</p>
      </div>

      <div class="sg-form-grid">
        <div class="sg-form-group">
          <label class="sg-form-label">Prometheus Base URL</label>
          <input
            type="text"
            class="sg-input"
            [(ngModel)]="observability().prometheusEndpoint"
            (ngModelChange)="changed.emit()"
            placeholder="http://prometheus-operated.monitoring:9090"
          />
          <span class="sg-form-hint">Backend for PromQL queries, metrics, and alerts</span>
        </div>

        <div class="sg-form-group">
          <label class="sg-form-label">Grafana Base URL</label>
          <input
            type="text"
            class="sg-input"
            [(ngModel)]="observability().grafanaEndpoint"
            (ngModelChange)="changed.emit()"
            placeholder="http://grafana.monitoring:3000"
          />
          <span class="sg-form-hint">Deep link target for predefined networking dashboards</span>
        </div>

        <div class="sg-form-group">
          <label class="sg-form-label">Jaeger Query / Collector Endpoint</label>
          <input
            type="text"
            class="sg-input"
            [(ngModel)]="observability().jaegerEndpoint"
            (ngModelChange)="changed.emit()"
            placeholder="http://jaeger-query.monitoring:16686"
          />
          <span class="sg-form-hint">Distributed tracing backend for packet hop inspection</span>
        </div>

        <div class="sg-form-group">
          <label class="sg-form-label">Metrics Scrape Frequency (seconds)</label>
          <input
            type="number"
            class="sg-input"
            [(ngModel)]="observability().metricsScrapeIntervalSeconds"
            (ngModelChange)="changed.emit()"
            min="5"
            max="60"
          />
        </div>
      </div>
    </div>
  `,
})
export class ObservabilitySettingsComponent {
  observability = input.required<any>();
  changed = output<void>();
}
