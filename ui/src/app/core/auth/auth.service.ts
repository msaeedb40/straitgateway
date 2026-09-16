import { inject, Injectable, signal } from '@angular/core';
import { Router } from '@angular/router';
import { AuthState, AuthUser } from './auth.model';
import { OidcService } from './oidc.service';

@Injectable({ providedIn: 'root' })
export class AuthService {
  private readonly oidc = inject(OidcService);
  private readonly router = inject(Router);

  readonly state = signal<AuthState>({
    status: 'Unauthenticated',
    user: null,
    error: null,
  });

  get isAuthenticated(): boolean {
    return this.state().status === 'Authenticated' && this.state().user !== null;
  }

  get currentUser(): AuthUser | null {
    return this.state().user;
  }

  get accessToken(): string | null {
    return this.state().user?.accessToken ?? null;
  }

  async initialize(): Promise<void> {
    this.state.update((s) => ({ ...s, status: 'Authenticating' }));
    try {
      const user = await this.oidc.initialize();
      if (user) {
        this.state.set({ status: 'Authenticated', user, error: null });
      } else {
        this.state.set({ status: 'Unauthenticated', user: null, error: null });
      }
    } catch (err) {
      this.state.set({
        status: 'Error',
        user: null,
        error: err instanceof Error ? err.message : 'Authentication failed',
      });
    }
  }

  async login(): Promise<void> {
    await this.oidc.startLogin();
  }

  async logout(): Promise<void> {
    await this.oidc.logout();
    this.state.set({ status: 'Unauthenticated', user: null, error: null });
    await this.router.navigate(['/login']);
  }

  async handleCallback(): Promise<void> {
    try {
      const user = await this.oidc.handleCallback();
      this.state.set({ status: 'Authenticated', user, error: null });
    } catch (err) {
      this.state.set({
        status: 'Error',
        user: null,
        error: err instanceof Error ? err.message : 'Callback handling failed',
      });
    }
  }
}
