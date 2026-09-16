import { Injectable, signal } from '@angular/core';

export type NotificationSeverity = 'info' | 'success' | 'warning' | 'error';

export interface Notification {
  readonly id: string;
  readonly severity: NotificationSeverity;
  readonly title: string;
  readonly message?: string;
  readonly durationMs?: number;
}

@Injectable({ providedIn: 'root' })
export class NotificationService {
  readonly notifications = signal<Notification[]>([]);

  push(n: Omit<Notification, 'id'>): string {
    const id = crypto.randomUUID();
    this.notifications.update((list) => [...list, { ...n, id }]);
    if (n.durationMs) {
      setTimeout(() => this.dismiss(id), n.durationMs);
    }
    return id;
  }

  dismiss(id: string): void {
    this.notifications.update((list) => list.filter((n) => n.id !== id));
  }

  clear(): void {
    this.notifications.set([]);
  }

  info(title: string, message?: string): string {
    return this.push({ severity: 'info', title, message, durationMs: 5000 });
  }

  success(title: string, message?: string): string {
    return this.push({ severity: 'success', title, message, durationMs: 4000 });
  }

  warning(title: string, message?: string): string {
    return this.push({ severity: 'warning', title, message, durationMs: 7000 });
  }

  error(title: string, message?: string): string {
    return this.push({ severity: 'error', title, message });
  }
}
