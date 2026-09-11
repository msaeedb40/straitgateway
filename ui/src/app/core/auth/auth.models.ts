// Copyright 2026 straitgateway Authors
// SPDX-License-Identifier: Apache-2.0

export interface UserSession {
  username: string;
  role: 'admin' | 'operator' | 'viewer';
  token?: string;
  tokenType: 'bearer' | 'serviceaccount' | 'anonymous';
  expiresAt?: number;
}

export interface AuthState {
  isAuthenticated: boolean;
  user: UserSession | null;
  allowedNamespaces: string[];
}
