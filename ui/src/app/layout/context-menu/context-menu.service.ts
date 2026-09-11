// Copyright 2026 straitgateway Authors
// SPDX-License-Identifier: Apache-2.0

import { Injectable, signal } from '@angular/core';

export interface ContextMenuItem {
  label: string;
  icon?: string;
  action: (data?: any) => void;
  danger?: boolean;
  disabled?: boolean;
  divider?: boolean;
}

export interface ContextMenuState {
  isOpen: boolean;
  x: number;
  y: number;
  items: ContextMenuItem[];
  data?: any;
}

@Injectable({
  providedIn: 'root',
})
export class ContextMenuService {
  private readonly _state = signal<ContextMenuState>({
    isOpen: false,
    x: 0,
    y: 0,
    items: [],
  });

  readonly state = this._state.asReadonly();

  open(x: number, y: number, items: ContextMenuItem[], data?: any): void {
    // Prevent menu from overflowing window bounds
    const menuWidth = 200;
    const menuHeight = items.length * 36 + 16;
    const maxX = typeof window !== 'undefined' ? window.innerWidth - menuWidth - 10 : x;
    const maxY = typeof window !== 'undefined' ? window.innerHeight - menuHeight - 10 : y;

    this._state.set({
      isOpen: true,
      x: Math.min(x, maxX),
      y: Math.min(y, maxY),
      items,
      data,
    });
  }

  close(): void {
    this._state.update(s => ({ ...s, isOpen: false }));
  }
}
