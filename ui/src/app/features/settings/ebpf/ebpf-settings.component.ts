// Copyright 2026 straitgateway Authors
// SPDX-License-Identifier: Apache-2.0

import { Component, input, output } from '@angular/core';
import { CommonModule } from '@angular/common';
import { FormsModule } from '@angular/forms';

@Component({
  selector: 'sg-ebpf-settings',
  standalone: true,
  imports: [CommonModule, FormsModule],
  template: `
    <div class="sg-settings-section">
      <div class="sg-section-header">
        <h3 class="sg-section-title">eBPF Dataplane & Maps</h3>
        <p class="sg-section-subtitle">Kernel dataplane parameters, map capacities, and load balancing algorithms</p>
      </div>

      <div class="sg-form-grid">
        <div class="sg-form-group">
          <label class="sg-form-label">BPFFS Mount Path</label>
          <input
            type="text"
            class="sg-input mono"
            [(ngModel)]="ebpf().bpfFsPath"
            (ngModelChange)="changed.emit()"
            placeholder="/sys/fs/bpf"
          />
          <span class="sg-form-hint">Pinned eBPF maps location in host virtual filesystem</span>
        </div>

        <div class="sg-form-group">
          <label class="sg-form-label">RingBuffer Size (MB)</label>
          <input
            type="number"
            class="sg-input"
            [(ngModel)]="ebpf().ringBufferSizeMb"
            (ngModelChange)="changed.emit()"
            min="4"
            max="128"
          />
          <span class="sg-form-hint">Kernel-to-userspace flow event buffer allocation</span>
        </div>

        <div class="sg-form-group">
          <label class="sg-form-label">Conntrack Table Max Entries</label>
          <input
            type="number"
            class="sg-input"
            [(ngModel)]="ebpf().conntrackMaxEntries"
            (ngModelChange)="changed.emit()"
            step="65536"
            min="65536"
            max="2097152"
          />
          <span class="sg-form-hint">Maximum concurrent L4 connection tracking states</span>
        </div>

        <div class="sg-form-group">
          <label class="sg-form-label">Maglev Consistent Hashing</label>
          <div style="display:flex;align-items:center;gap:10px;margin-top:8px">
            <input
              type="checkbox"
              id="maglevHash"
              [(ngModel)]="ebpf().enableMaglevHash"
              (ngModelChange)="changed.emit()"
              style="accent-color:var(--sg-accent);width:16px;height:16px"
            />
            <label for="maglevHash" style="font-size:13px;color:var(--sg-text-2);cursor:pointer">
              Enable Maglev lookup table (lookup table size: 128 / consistent hash)
            </label>
          </div>
        </div>

        <div class="sg-form-group">
          <label class="sg-form-label">Direct Server Return (DSR)</label>
          <div style="display:flex;align-items:center;gap:10px;margin-top:8px">
            <input
              type="checkbox"
              id="dsrMode"
              [(ngModel)]="ebpf().enableDsr"
              (ngModelChange)="changed.emit()"
              style="accent-color:var(--sg-accent);width:16px;height:16px"
            />
            <label for="dsrMode" style="font-size:13px;color:var(--sg-text-2);cursor:pointer">
              Bypass gateway return hops for North-South egress traffic
            </label>
          </div>
        </div>
      </div>
    </div>
  `,
})
export class EbpfSettingsComponent {
  ebpf = input.required<any>();
  changed = output<void>();
}
