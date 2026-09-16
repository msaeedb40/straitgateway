import { Component, inject, OnInit, signal, computed } from '@angular/core';
import { takeUntilDestroyed } from '@angular/core/rxjs-interop';
import { DestroyRef } from '@angular/core';
import { AuthService } from '../../../core/auth/auth.service';

@Component({
  selector: 'sg-login',
  template: `
    <div class="sg-login-page" role="main">
      <div class="sg-login-card sg-glass">
        <div class="sg-login-brand">
          <h1 class="sg-login-title">StraitGateway</h1>
          <p class="sg-login-sub">Management Console</p>
        </div>

        @if (auth.state().status === 'Error') {
          <div class="sg-login-error" role="alert">
            <strong>Authentication failed:</strong> {{ auth.state().error }}
          </div>
        }

        <button
          class="sg-btn sg-btn-primary sg-login-btn"
          (click)="login()"
          [disabled]="auth.state().status === 'Authenticating'"
          type="button"
          autofocus
        >
          {{ auth.state().status === 'Authenticating' ? 'Redirecting…' : 'Sign in with OIDC' }}
        </button>

        <p class="sg-login-info">
          Authentication is provided via your configured identity provider.
        </p>
      </div>
    </div>
  `,
  styles: [`
    .sg-login-page {
      min-height: 100dvh;
      display: flex;
      align-items: center;
      justify-content: center;
      background:
        radial-gradient(ellipse at 30% 40%, rgba(56,189,248,.08) 0%, transparent 60%),
        radial-gradient(ellipse at 70% 60%, rgba(45,212,191,.06) 0%, transparent 60%),
        var(--sg-bg-base);
    }
    .sg-login-card {
      width: min(400px, calc(100vw - 2rem));
      padding: 2.5rem;
      border-radius: var(--sg-radius-xl);
      display: flex;
      flex-direction: column;
      gap: 1.5rem;
      text-align: center;
    }
    .sg-login-brand { display: flex; flex-direction: column; gap: 4px; }
    .sg-login-title {
      font-size: 1.5rem;
      font-weight: 700;
      color: var(--sg-accent);
      letter-spacing: -0.03em;
    }
    .sg-login-sub { font-size: 0.875rem; color: var(--sg-text-secondary); }
    .sg-login-btn { width: 100%; justify-content: center; padding: 10px; }
    .sg-login-info { font-size: 0.75rem; color: var(--sg-text-muted); }
    .sg-login-error {
      background: rgba(248,113,113,.1);
      border: 1px solid rgba(248,113,113,.3);
      border-radius: var(--sg-radius);
      padding: 10px 14px;
      font-size: 0.8125rem;
      color: var(--sg-failed);
      text-align: left;
    }
  `],
})
export class LoginComponent {
  readonly auth = inject(AuthService);
  async login(): Promise<void> { await this.auth.login(); }
}
