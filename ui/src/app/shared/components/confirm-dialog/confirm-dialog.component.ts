import { Component, input, output } from '@angular/core';
import { ApiError } from '../../../core/api/api-error';

export interface ConfirmDialogImpact {
  readonly label: string;
  readonly count: number;
}

@Component({
  selector: 'sg-confirm-dialog',
  template: `
    @if (open()) {
      <div class="sg-modal-backdrop" (click)="cancel.emit()" role="presentation">
        <div
          class="sg-modal"
          role="alertdialog"
          aria-modal="true"
          [attr.aria-labelledby]="'confirm-title-' + dialogId"
          [attr.aria-describedby]="'confirm-desc-' + dialogId"
          (click)="$event.stopPropagation()"
          cdkTrapFocus
        >
          <div class="sg-modal-header">
            <h2 [id]="'confirm-title-' + dialogId" class="sg-modal-title">{{ title() }}</h2>
          </div>

          <div class="sg-modal-body" [id]="'confirm-desc-' + dialogId">
            @if (message()) {
              <p class="sg-confirm-message">{{ message() }}</p>
            }
            @if (impacts().length > 0) {
              <div class="sg-confirm-impacts">
                <p class="sg-confirm-impacts-title">This will affect:</p>
                <ul class="sg-confirm-impacts-list">
                  @for (item of impacts(); track item.label) {
                    <li>{{ item.count }} {{ item.label }}</li>
                  }
                </ul>
              </div>
            }
            @if (error()) {
              <p class="sg-confirm-error" role="alert">{{ error()!.message }}</p>
            }
          </div>

          <div class="sg-modal-footer">
            <button
              class="sg-btn sg-btn-secondary"
              (click)="cancel.emit()"
              type="button"
              autofocus
            >Cancel</button>
            <button
              class="sg-btn sg-btn-danger"
              (click)="confirm.emit()"
              [disabled]="loading()"
              type="button"
            >
              {{ loading() ? 'Working…' : confirmLabel() }}
            </button>
          </div>
        </div>
      </div>
    }
  `,
  styles: [`
    .sg-modal-backdrop {
      position: fixed; inset: 0;
      background: rgba(0,0,0,0.6);
      display: flex; align-items: center; justify-content: center;
      z-index: 1000;
      backdrop-filter: blur(4px);
    }
    .sg-modal {
      background: var(--sg-bg-elevated);
      border: 1px solid var(--sg-border);
      border-radius: var(--sg-radius-xl);
      width: min(480px, calc(100vw - 2rem));
      box-shadow: var(--sg-shadow-lg);
    }
    .sg-modal-header { padding: 1.25rem 1.5rem 0; }
    .sg-modal-title  { font-size: 1rem; font-weight: 600; color: var(--sg-text-primary); }
    .sg-modal-body   { padding: 1rem 1.5rem; }
    .sg-modal-footer {
      padding: 1rem 1.5rem;
      display: flex; justify-content: flex-end; gap: 8px;
      border-top: 1px solid var(--sg-border);
    }
    .sg-confirm-message { color: var(--sg-text-secondary); font-size: 0.875rem; margin-bottom: 12px; }
    .sg-confirm-impacts { background: var(--sg-bg-overlay); border-radius: var(--sg-radius); padding: 12px; }
    .sg-confirm-impacts-title { font-size: 0.75rem; font-weight: 600; color: var(--sg-text-secondary); margin-bottom: 6px; }
    .sg-confirm-impacts-list { list-style: disc; padding-left: 1.25rem; font-size: 0.8125rem; color: var(--sg-text-primary); }
    .sg-confirm-impacts-list li { margin-bottom: 4px; }
    .sg-confirm-error { color: var(--sg-failed); font-size: 0.8125rem; margin-top: 8px; }
  `],
})
export class ConfirmDialogComponent {
  readonly open         = input<boolean>(false);
  readonly title        = input<string>('Confirm action');
  readonly message      = input<string | undefined>(undefined);
  readonly confirmLabel = input<string>('Confirm');
  readonly impacts      = input<ConfirmDialogImpact[]>([]);
  readonly loading      = input<boolean>(false);
  readonly error        = input<ApiError | null>(null);

  readonly confirm = output<void>();
  readonly cancel  = output<void>();

  readonly dialogId = Math.random().toString(36).slice(2, 8);
}
