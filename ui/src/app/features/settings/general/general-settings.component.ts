// Copyright 2026 straitgateway Authors
// SPDX-License-Identifier: Apache-2.0

import { Component, input, output } from '@angular/core';
import { CommonModule } from '@angular/common';
import { FormsModule } from '@angular/forms';

@Component({
  selector: 'sg-general-settings',
  standalone: true,
  imports: [CommonModule, FormsModule],
  template: `
    <div class="sg-settings-section">
      <div class="sg-section-header">
        <h3 class="sg-section-title">General Preferences</h3>
        <p class="sg-section-subtitle">Display preferences, themes, and default namespaces</p>
      </div>

      <div class="sg-form-grid">
        <div class="sg-form-group">
          <label class="sg-form-label">Theme Mode</label>
          <select
            class="sg-select"
            [(ngModel)]="general().theme"
            (ngModelChange)="changed.emit()"
          >
            <option value="dark">Dark Theme (Default)</option>
            <option value="light">Light Theme</option>
          </select>
        </div>

        <div class="sg-form-group">
          <label class="sg-form-label">Default Selected Namespace</label>
          <input
            type="text"
            class="sg-input"
            [(ngModel)]="general().defaultNamespace"
            (ngModelChange)="changed.emit()"
            placeholder="default"
          />
          <span class="sg-form-hint">Selected automatically on application launch</span>
        </div>

        <div class="sg-form-group">
          <label class="sg-form-label">Table Density</label>
          <div style="display:flex;align-items:center;gap:10px;margin-top:8px">
            <input
              type="checkbox"
              id="compactTables"
              [(ngModel)]="general().compactTables"
              (ngModelChange)="changed.emit()"
              style="accent-color:var(--sg-accent);width:16px;height:16px"
            />
            <label for="compactTables" style="font-size:13px;color:var(--sg-text-2);cursor:pointer">
              Compact table row padding for high-density monitoring
            </label>
          </div>
        </div>

        <div class="sg-form-group">
          <label class="sg-form-label">Sound Notifications</label>
          <div style="display:flex;align-items:center;gap:10px;margin-top:8px">
            <input
              type="checkbox"
              id="soundAlerts"
              [(ngModel)]="general().enableSoundAlerts"
              (ngModelChange)="changed.emit()"
              style="accent-color:var(--sg-accent);width:16px;height:16px"
            />
            <label for="soundAlerts" style="font-size:13px;color:var(--sg-text-2);cursor:pointer">
              Audible alert chime on critical packet drops or BGP peer resets
            </label>
          </div>
        </div>
      </div>
    </div>
  `,
})
export class GeneralSettingsComponent {
  general = input.required<any>();
  changed = output<void>();
}
