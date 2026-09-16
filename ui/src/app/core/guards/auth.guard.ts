import { inject } from '@angular/core';
import { CanActivateFn, Router } from '@angular/router';
import { AuthService } from '../auth/auth.service';

export const authGuard: CanActivateFn = async () => {
  const auth = inject(AuthService);
  const router = inject(Router);

  if (auth.isAuthenticated) return true;

  // Attempt silent initialization first (e.g. returning from OIDC callback)
  await auth.initialize();

  if (auth.isAuthenticated) return true;

  // Redirect to login — the login component triggers the OIDC flow
  return router.createUrlTree(['/login']);
};
