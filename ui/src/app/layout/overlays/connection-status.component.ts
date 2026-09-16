import { Component, inject, computed } from '@angular/core';
import { ConnectionService } from '../../core/services/connection.service';
import { ConnectionStatus } from '../../core/api/api.types';

interface BackendEntry {
  readonly key: keyof ReturnType<ConnectionService['snapshot']['constructor']['prototype']>;
  readonly label: string;
}

@Component({
  selector: 'sg-connection-status',
  template: `
    <aside
      class="sg-conn-strip"
      aria-label="Backend connection status"
      role="status"
      aria-live="polite"
    >
      @for (entry of entries; track entry.label) {
        <span
          class="sg-conn-entry"
          [class.sg-conn-connected]="statusOf(entry.signal()) === 'connected'"
          [class.sg-conn-unavailable]="statusOf(entry.signal()) === 'unavailable'"
          [class.sg-conn-checking]="statusOf(entry.signal()) === 'checking'"
          [attr.aria-label]="entry.label + ': ' + statusOf(entry.signal())"
        >
          <span class="sg-conn-dot" aria-hidden="true"></span>
          <span class="sg-conn-label">{{ entry.label }}</span>
        </span>
      }
    </aside>
  `,
  styles: [`
    .sg-conn-strip {
      position: fixed;
      bottom: 0;
      left: var(--sg-sidebar-width);
      right: 0;
      display: flex;
      align-items: center;
      gap: 16px;
      padding: 4px 16px;
      background: var(--sg-bg-surface);
      border-top: 1px solid var(--sg-border);
      font-size: 0.6875rem;
      z-index: 50;
      transition: left var(--sg-transition);
    }

    .sg-conn-entry {
      display: inline-flex;
      align-items: center;
      gap: 5px;
      color: var(--sg-text-muted);
      white-space: nowrap;
    }

    .sg-conn-dot {
      width: 6px;
      height: 6px;
      border-radius: 50%;
      background: var(--sg-unknown);
      flex-shrink: 0;
    }

    .sg-conn-connected   .sg-conn-dot { background: var(--sg-healthy); }
    .sg-conn-connected   { color: var(--sg-text-secondary); }
    .sg-conn-unavailable .sg-conn-dot { background: var(--sg-failed); }
    .sg-conn-unavailable { color: var(--sg-failed); }
    .sg-conn-checking    .sg-conn-dot { background: var(--sg-unknown); }
  `],
})
export class ConnectionStatusComponent {
  private readonly conn = inject(ConnectionService);

  statusOf(s: ConnectionStatus): ConnectionStatus { return s; }

  readonly entries = [
    { label: 'Controller', signal: this.conn.controller },
    { label: 'Prometheus', signal: this.conn.prometheusStatus },
    { label: 'Grafana',    signal: this.conn.grafanaStatus },
    { label: 'Jaeger',     signal: this.conn.jaegerStatus },
    { label: 'Logs',       signal: this.conn.logsStatus },
  ];
}
