// Copyright 2026 straitgateway Authors
// SPDX-License-Identifier: Apache-2.0

import { Injectable, inject, signal } from '@angular/core';
import { StorageService } from '../services/storage.service';
import { AuthState, UserSession } from './auth.models';

const AUTH_STORAGE_KEY = 'straitgateway_auth_v1';

@Injectable({
  providedIn: 'root',
})
export class AuthService {
  private storage = inject(StorageService);

  private readonly _state = signal<AuthState>(this.initAuth());
  readonly state = this._state.asReadonly();

  readonly isAuthenticated = signal<boolean>(true); // Default to cluster-authenticated session
  readonly currentUser = signal<UserSession>({
    username: 'admin',
    role: 'admin',
    tokenType: 'serviceaccount',
  });

  private initAuth(): AuthState {
    const saved = this.storage.getSession<AuthState | null>(AUTH_STORAGE_KEY, null);
    if (saved) return saved;

    return {
      isAuthenticated: true,
      user: {
        username: 'cluster-admin',
        role: 'admin',
        tokenType: 'serviceaccount',
      },
      allowedNamespaces: ['*'],
    };
  }

  setSession(session: UserSession, allowedNamespaces: string[] = ['*']): void {
    const state: AuthState = {
      isAuthenticated: true,
      user: session,
      allowedNamespaces,
    };
    this._state.set(state);
    this.isAuthenticated.set(true);
    this.currentUser.set(session);
    this.storage.setSession(AUTH_STORAGE_KEY, state);
  }

  logout(): void {
    const defaultState: AuthState = {
      isAuthenticated: false,
      user: null,
      allowedNamespaces: [],
    };
    this._state.set(defaultState);
    this.isAuthenticated.set(false);
    this.storage.removeSession(AUTH_STORAGE_KEY);
  }

  canWrite(): boolean {
    const role = this.currentUser().role;
    return role === 'admin' || role === 'operator';
  }

  canAdmin(): boolean {
    return this.currentUser().role === 'admin';
  }
}
