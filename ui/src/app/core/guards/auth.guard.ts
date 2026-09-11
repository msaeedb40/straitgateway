// Copyright 2026 straitgateway Authors
// SPDX-License-Identifier: Apache-2.0

import { inject } from '@angular/core';
import { CanActivateFn, Router } from '@angular/router';
import { AuthService } from '../auth/auth.service';

export const authGuard: CanActivateFn = () => {
  const auth = inject(AuthService);
  const router = inject(Router);

  if (auth.isAuthenticated()) {
    return true;
  }

  // Redirect to dashboard or login
  return router.parseUrl('/');
};

export const adminGuard: CanActivateFn = () => {
  const auth = inject(AuthService);
  const router = inject(Router);

  if (auth.canAdmin()) {
    return true;
  }

  return router.parseUrl('/');
};
