import { Component, input } from '@angular/core';
import { ReconciliationPhase } from '../../../core/api/api.types';

interface PhaseStep {
  readonly phase: ReconciliationPhase;
  readonly label: string;
}

const STEPS: PhaseStep[] = [
  { phase: 'Pending',     label: 'Pending' },
  { phase: 'Reconciling', label: 'Reconciling' },
  { phase: 'Ready',       label: 'Ready' },
];

@Component({
  selector: 'sg-reconciliation-status',
  template: `
    <div class="sg-recon" role="status" [attr.aria-label]="'Reconciliation: ' + currentPhase()">
      <ol class="sg-recon-steps" role="list">
        @for (step of steps; track step.phase; let i = $index) {
          <li
            class="sg-recon-step"
            [class.sg-recon-active]="step.phase === currentPhase()"
            [class.sg-recon-done]="isDone(step.phase)"
            [class.sg-recon-failed]="currentPhase() === 'Failed' && step.phase === 'Reconciling'"
            role="listitem"
            [attr.aria-current]="step.phase === currentPhase() ? 'step' : null"
          >
            <span class="sg-recon-dot" aria-hidden="true"></span>
            <span class="sg-recon-label">{{ step.label }}</span>
            @if (i < steps.length - 1) {
              <span class="sg-recon-line" aria-hidden="true"></span>
            }
          </li>
        }
      </ol>
      @if (currentPhase() === 'Failed' && message()) {
        <p class="sg-recon-error" role="alert">{{ message() }}</p>
      }
    </div>
  `,
  styles: [`
    .sg-recon-steps {
      display: flex;
      align-items: center;
      gap: 0;
      list-style: none;
      margin: 0; padding: 0;
    }
    .sg-recon-step {
      display: flex;
      align-items: center;
      gap: 6px;
      font-size: 0.75rem;
      color: var(--sg-text-muted);
    }
    .sg-recon-dot {
      width: 8px; height: 8px;
      border-radius: 50%;
      border: 1.5px solid currentColor;
      flex-shrink: 0;
    }
    .sg-recon-line {
      width: 32px; height: 1px;
      background: var(--sg-border);
      margin: 0 6px;
    }
    .sg-recon-active  { color: var(--sg-accent); }
    .sg-recon-active  .sg-recon-dot { background: var(--sg-accent); border-color: var(--sg-accent); }
    .sg-recon-done    { color: var(--sg-healthy); }
    .sg-recon-done    .sg-recon-dot { background: var(--sg-healthy); border-color: var(--sg-healthy); }
    .sg-recon-failed  { color: var(--sg-failed); }
    .sg-recon-failed  .sg-recon-dot { background: var(--sg-failed); border-color: var(--sg-failed); }
    .sg-recon-error { font-size: 0.75rem; color: var(--sg-failed); margin-top: 6px; }
  `],
})
export class ReconciliationStatusComponent {
  readonly currentPhase = input<ReconciliationPhase>('Unknown');
  readonly message      = input<string | undefined>(undefined);

  readonly steps = STEPS;

  isDone(phase: ReconciliationPhase): boolean {
    const order: ReconciliationPhase[] = ['Pending', 'Reconciling', 'Ready'];
    const current = order.indexOf(this.currentPhase());
    const target  = order.indexOf(phase);
    return current > target && this.currentPhase() !== 'Failed';
  }
}
