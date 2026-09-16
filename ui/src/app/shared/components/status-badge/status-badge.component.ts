import { Component, input } from '@angular/core';
import { HealthStatus } from '../../../core/models/topology.model';
import { ReconciliationPhase } from '../../../core/api/api.types';

export type StatusVariant = 'healthy' | 'degraded' | 'failed' | 'unknown' | 'pending' | 'reconciling' | 'ready';

const HEALTH_TO_VARIANT: Record<HealthStatus, StatusVariant> = {
  Healthy:  'healthy',
  Degraded: 'degraded',
  Failed:   'failed',
  Unknown:  'unknown',
};

const PHASE_TO_VARIANT: Record<ReconciliationPhase, StatusVariant> = {
  Pending:      'pending',
  Reconciling:  'reconciling',
  Ready:        'ready',
  Failed:       'failed',
  Unknown:      'unknown',
};

const VARIANT_LABEL: Record<StatusVariant, string> = {
  healthy:     '● Healthy',
  degraded:    '⚠ Degraded',
  failed:      '✕ Failed',
  unknown:     '? Unknown',
  pending:     '◌ Pending',
  reconciling: '↻ Reconciling',
  ready:       '● Ready',
};

@Component({
  selector: 'sg-status-badge',
  template: `
    <span
      class="sg-status-badge"
      [class]="'sg-status-badge-' + resolvedVariant()"
      [attr.aria-label]="ariaLabel()"
      role="img"
    >
      {{ label() }}
    </span>
  `,
  styles: [`
    .sg-status-badge {
      display: inline-flex;
      align-items: center;
      padding: 2px 8px;
      border-radius: 999px;
      font-size: 0.6875rem;
      font-weight: 600;
      white-space: nowrap;
    }
    .sg-status-badge-healthy     { background: rgba(52,211,153,.15); color: var(--sg-healthy);  }
    .sg-status-badge-ready       { background: rgba(52,211,153,.15); color: var(--sg-healthy);  }
    .sg-status-badge-degraded    { background: rgba(251,191,36,.15);  color: var(--sg-degraded); }
    .sg-status-badge-failed      { background: rgba(248,113,113,.15); color: var(--sg-failed);  }
    .sg-status-badge-unknown     { background: rgba(148,163,184,.1);  color: var(--sg-unknown); }
    .sg-status-badge-pending     { background: rgba(148,163,184,.1);  color: var(--sg-unknown); }
    .sg-status-badge-reconciling { background: rgba(56,189,248,.15);  color: var(--sg-accent);  }
  `],
})
export class StatusBadgeComponent {
  /** Explicit variant — use when you already have a UI variant string */
  readonly variant = input<StatusVariant | undefined>(undefined);
  /** Health status from topology model */
  readonly health  = input<HealthStatus | undefined>(undefined);
  /** Reconciliation phase from API types */
  readonly phase   = input<ReconciliationPhase | undefined>(undefined);

  resolvedVariant(): StatusVariant {
    if (this.variant()) return this.variant()!;
    if (this.health())  return HEALTH_TO_VARIANT[this.health()!];
    if (this.phase())   return PHASE_TO_VARIANT[this.phase()!];
    return 'unknown';
  }

  label(): string {
    return VARIANT_LABEL[this.resolvedVariant()];
  }

  ariaLabel(): string {
    return `Status: ${this.resolvedVariant()}`;
  }
}
