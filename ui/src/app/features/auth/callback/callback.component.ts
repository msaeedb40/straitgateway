import { Component, inject, OnInit } from '@angular/core';
import { Router } from '@angular/router';
import { AuthService } from '../../../core/auth/auth.service';

@Component({
  selector: 'sg-callback',
  template: `
    <div class="sg-callback-page" role="status" aria-live="polite">
      <p>Completing sign-in…</p>
      @if (error()) {
        <p class="sg-callback-error" role="alert">{{ error() }}</p>
      }
    </div>
  `,
  styles: [`
    .sg-callback-page {
      min-height: 100dvh; display: flex; flex-direction: column;
      align-items: center; justify-content: center;
      color: var(--sg-text-secondary); font-size: 0.9375rem; gap: 12px;
    }
    .sg-callback-error { color: var(--sg-failed); }
  `],
})
export class CallbackComponent implements OnInit {
  private readonly auth   = inject(AuthService);
  private readonly router = inject(Router);
  error = () => this.auth.state().error;

  async ngOnInit(): Promise<void> {
    await this.auth.handleCallback();
    if (this.auth.isAuthenticated) {
      await this.router.navigate(['/dashboard']);
    }
  }
}
