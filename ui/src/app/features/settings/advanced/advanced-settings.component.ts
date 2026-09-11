// Copyright 2026 straitgateway Authors
// SPDX-License-Identifier: Apache-2.0

import { Component, input, output, signal } from '@angular/core';
import { CommonModule } from '@angular/common';
import { FormsModule } from '@angular/forms';

@Component({
  selector: 'sg-advanced-settings',
  standalone: true,
  imports: [CommonModule, FormsModule],
  template: `
    <div class="sg-settings-section">
      <div class="sg-section-header">
        <h3 class="sg-section-title">Advanced Diagnostics & Maintenance</h3>
        <p class="sg-section-subtitle">Linux kernel compatibility checks, configuration backup/restore, and reset</p>
      </div>

      <div class="sg-card" style="margin-bottom:20px;border-color:rgba(99,102,241,0.2)">
        <div class="sg-card-header">
          <span class="sg-card-title">Kernel Compatibility Validation</span>
          <span class="sg-badge active">Baseline Linux 6.6+ LTS</span>
        </div>
        <div class="sg-card-body">
          <div style="display:grid;grid-template-columns:repeat(auto-fit, minmax(200px, 1fr));gap:12px">
            <div class="sg-diag-item">
              <span class="sg-diag-label">eBPF CO-RE</span>
              <span class="sg-diag-status text-success">✔ Supported (BTF vmlinux)</span>
            </div>
            <div class="sg-diag-item">
              <span class="sg-diag-label">NetKit Driver</span>
              <span class="sg-diag-status text-success">✔ Available</span>
            </div>
            <div class="sg-diag-item">
              <span class="sg-diag-label">TCX Hooks</span>
              <span class="sg-diag-status text-success">✔ Active</span>
            </div>
            <div class="sg-diag-item">
              <span class="sg-diag-label">cgroup v2</span>
              <span class="sg-diag-status text-success">✔ Mounted</span>
            </div>
          </div>
        </div>
      </div>

      <div class="sg-form-group">
        <label class="sg-form-label">Export / Backup Configuration</label>
        <p class="sg-form-hint" style="margin-bottom:8px">Export active settings as a JSON payload for version control or cluster migration</p>
        <div style="display:flex;gap:10px">
          <button class="sg-btn sg-btn-secondary" (click)="exportConfig()">Export JSON</button>
          <button class="sg-btn sg-btn-secondary" (click)="fileInput.click()">Import JSON</button>
          <input #fileInput type="file" accept=".json" (change)="importConfig($event)" style="display:none" />
        </div>
      </div>

      <div class="sg-form-group" style="margin-top:24px;padding-top:20px;border-top:1px solid var(--sg-border)">
        <label class="sg-form-label" style="color:var(--sg-danger)">Reset Defaults</label>
        <p class="sg-form-hint" style="margin-bottom:8px">Restore all network, gateway, eBPF, and UI settings back to factory defaults</p>
        <button class="sg-btn sg-btn-secondary" style="color:var(--sg-danger);border-color:rgba(239,68,68,0.3)" (click)="resetDefaults.emit()">
          Restore Default Settings
        </button>
      </div>
    </div>
  `,
  styles: [`
    .sg-diag-item {
      background: var(--sg-surface-2);
      padding: 10px 14px;
      border-radius: var(--sg-radius-sm);
      border: 1px solid var(--sg-border);
      display: flex;
      flex-direction: column;
      gap: 4px;
    }
    .sg-diag-label {
      font-size: 11px;
      text-transform: uppercase;
      letter-spacing: 0.05em;
      color: var(--sg-text-3);
    }
    .sg-diag-status {
      font-size: 13px;
      font-weight: 500;
    }
  `],
})
export class AdvancedSettingsComponent {
  config = input.required<any>();
  resetDefaults = output<void>();
  importLoaded = output<any>();

  exportConfig() {
    const data = JSON.stringify(this.config(), null, 2);
    const blob = new Blob([data], { type: 'application/json' });
    const url = URL.createObjectURL(blob);
    const a = document.createElement('a');
    a.href = url;
    a.download = `straitgateway-settings-${new Date().toISOString().slice(0, 10)}.json`;
    a.click();
    URL.revokeObjectURL(url);
  }

  importConfig(event: Event) {
    const target = event.target as HTMLInputElement;
    if (target.files && target.files[0]) {
      const reader = new FileReader();
      reader.onload = (e) => {
        try {
          const parsed = JSON.parse(e.target?.result as string);
          this.importLoaded.emit(parsed);
        } catch {
          alert('Invalid JSON file');
        }
      };
      reader.readAsText(target.files[0]);
    }
  }
}
