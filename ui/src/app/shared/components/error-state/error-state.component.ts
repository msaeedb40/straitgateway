import { Component, input, output } from '@angular/core';
import { ApiError } from '../../../core/api/api-error';

@Component({
  selector: 'sg-error-state',
  template: `
    <div class="sg-error-state" role="alert" aria-live="assertive">
      <div class="sg-error-icon" aria-hidden="true">
        <svg width="40" height="40" viewBox="0 0 40 40" fill="none">
          <circle cx="20" cy="20" r="18" stroke="var(--sg-failed)" stroke-width="1.5" opacity=".4"/>
          <path d="M20 12v10M20 27v1" stroke="var(--sg-failed)" stroke-width="2" stroke-linecap="round"/>
        </svg>
      </div>
      <p class="sg-error-title">{{ title() }}</p>
      @if (error()) {
        <p class="sg-error-detail">
          {{ error()!.status ? 'HTTP ' + error()!.status + ' — ' : '' }}{{ error()!.message }}
        </p>
        @if (error()!.detail) {
          <pre class="sg-error-raw">{{ error()!.detail }}</pre>
        }
      }
      @if (error()?.retryable !== false) {
        <button class="sg-btn sg-btn-secondary" (click)="retry.emit()" type="button">
          Retry
        </button>
      }
    </div>
  `,
  styles: [`
    .sg-error-state {
      display: flex;
      flex-direction: column;
      align-items: center;
      gap: 12px;
      padding: 3rem 2rem;
      text-align: center;
    }
    .sg-error-title {
      font-size: 0.9375rem;
      font-weight: 600;
      color: var(--sg-text-primary);
    }
    .sg-error-detail {
      font-size: 0.8125rem;
      color: var(--sg-text-secondary);
      max-width: 480px;
    }
    .sg-error-raw {
      font-family: var(--sg-font-mono);
      font-size: 0.75rem;
      color: var(--sg-failed);
      background: rgba(248,113,113,0.08);
      border: 1px solid rgba(248,113,113,0.2);
      border-radius: var(--sg-radius);
      padding: 8px 12px;
      max-width: 480px;
      white-space: pre-wrap;
      word-break: break-word;
      text-align: left;
    }
  `],
})
export class ErrorStateComponent {
  readonly title  = input<string>('Failed to load data');
  readonly error  = input<ApiError | null>(null);
  readonly retry  = output<void>();
}

@Component({
  selector: 'sg-empty-state',
  template: `
    <div class="sg-empty-state" role="status">
      <div class="sg-empty-icon" aria-hidden="true">
        <svg width="40" height="40" viewBox="0 0 40 40" fill="none">
          <circle cx="20" cy="20" r="18" stroke="var(--sg-text-muted)" stroke-width="1.5" opacity=".4"/>
          <path d="M14 20h12M20 14v12" stroke="var(--sg-text-muted)" stroke-width="1.5" stroke-linecap="round" opacity=".5"/>
        </svg>
      </div>
      <p class="sg-empty-title">{{ title() }}</p>
      @if (message()) {
        <p class="sg-empty-message">{{ message() }}</p>
      }
      <ng-content />
    </div>
  `,
  styles: [`
    .sg-empty-state {
      display: flex;
      flex-direction: column;
      align-items: center;
      gap: 10px;
      padding: 3rem 2rem;
      text-align: center;
    }
    .sg-empty-title {
      font-size: 0.9375rem;
      font-weight: 600;
      color: var(--sg-text-secondary);
    }
    .sg-empty-message {
      font-size: 0.8125rem;
      color: var(--sg-text-muted);
      max-width: 380px;
    }
  `],
})
export class EmptyStateComponent {
  readonly title   = input<string>('No data');
  readonly message = input<string | undefined>(undefined);
}
