// Copyright 2026 straitgateway Authors
// SPDX-License-Identifier: Apache-2.0

import { Injectable, signal } from '@angular/core';

export interface NotificationItem {
  id: string;
  type: 'success' | 'info' | 'warn' | 'danger';
  title: string;
  message: string;
  timestamp: number;
  durationMs?: number;
}

@Injectable({ providedIn: 'root' })
export class NotificationService {
  private readonly _notifications = signal<NotificationItem[]>([]);
  readonly notifications = this._notifications.asReadonly();

  show(notif: Omit<NotificationItem, 'id' | 'timestamp'>) {
    const id = `notif-${Date.now()}-${Math.random().toString(36).substr(2, 6)}`;
    const item: NotificationItem = {
      ...notif,
      id,
      timestamp: Date.now(),
      durationMs: notif.durationMs ?? 4000,
    };

    this._notifications.update((list) => [item, ...list]);

    if (item.durationMs && item.durationMs > 0) {
      setTimeout(() => this.dismiss(id), item.durationMs);
    }
  }

  success(title: string, message: string) {
    this.show({ type: 'success', title, message });
  }

  info(title: string, message: string) {
    this.show({ type: 'info', title, message });
  }

  warn(title: string, message: string) {
    this.show({ type: 'warn', title, message });
  }

  danger(title: string, message: string) {
    this.show({ type: 'danger', title, message, durationMs: 7000 });
  }

  dismiss(id: string) {
    this._notifications.update((list) => list.filter((n) => n.id !== id));
  }
}
