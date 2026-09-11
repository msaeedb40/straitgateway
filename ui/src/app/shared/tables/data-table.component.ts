// Copyright 2026 straitgateway Authors — SPDX-License-Identifier: Apache-2.0

import { Component, input, output, signal } from '@angular/core';
import { CommonModule } from '@angular/common';

export interface TableColumn {
  key: string;
  label: string;
  sortable?: boolean;
  mono?: boolean;
  width?: string;
}

@Component({
  selector: 'sg-data-table',
  standalone: true,
  imports: [CommonModule],
  template: `
<div class="sg-table-wrap">
  <table class="sg-table">
    <thead>
      <tr>
        @for (col of columns(); track col.key) {
          <th [style.width]="col.width ?? 'auto'"
              [class.sortable]="col.sortable"
              (click)="col.sortable ? toggleSort(col.key) : null">
            {{ col.label }}
            @if (col.sortable && sortKey() === col.key) {
              <span class="sort-arrow">{{ sortDir() === 'asc' ? '↑' : '↓' }}</span>
            }
          </th>
        }
      </tr>
    </thead>
    <tbody>
      @for (row of sortedRows(); track trackBy() ? row[trackBy()!] : $index) {
        <tr (click)="rowClick.emit(row)" [class.clickable]="clickable()">
          @for (col of columns(); track col.key) {
            <td [class.mono]="col.mono">{{ row[col.key] ?? '—' }}</td>
          }
        </tr>
      } @empty {
        <tr><td [attr.colspan]="columns().length">
          <div class="sg-empty" style="padding:24px"><p>{{ emptyMessage() }}</p></div>
        </td></tr>
      }
    </tbody>
  </table>
</div>
`,
  styles: [`
    th.sortable { cursor: pointer; user-select: none; }
    th.sortable:hover { color: var(--sg-accent); }
    .sort-arrow { margin-left: 4px; font-size: 10px; }
    tr.clickable { cursor: pointer; }
    tr.clickable:hover { background: var(--sg-surface-2); }
  `],
})
export class DataTableComponent {
  columns      = input.required<TableColumn[]>();
  rows         = input.required<Record<string, any>[]>();
  trackBy      = input<string | null>(null);
  emptyMessage = input<string>('No data');
  clickable    = input<boolean>(true);
  rowClick     = output<Record<string, any>>();

  sortKey = signal<string>('');
  sortDir = signal<'asc' | 'desc'>('asc');

  toggleSort(key: string) {
    if (this.sortKey() === key) {
      this.sortDir.set(this.sortDir() === 'asc' ? 'desc' : 'asc');
    } else {
      this.sortKey.set(key);
      this.sortDir.set('asc');
    }
  }

  sortedRows(): Record<string, any>[] {
    const data = this.rows();
    const key = this.sortKey();
    if (!key) return data;
    const dir = this.sortDir() === 'asc' ? 1 : -1;
    return [...data].sort((a, b) => {
      const va = a[key], vb = b[key];
      if (va == null && vb == null) return 0;
      if (va == null) return dir;
      if (vb == null) return -dir;
      return va < vb ? -dir : va > vb ? dir : 0;
    });
  }
}
