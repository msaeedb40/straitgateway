import { HttpInterceptorFn } from '@angular/common/http';
import { inject } from '@angular/core';
import { AuthService } from '../auth/auth.service';
import { RuntimeConfigService } from '../config/runtime-config.service';

export const authInterceptor: HttpInterceptorFn = (req, next) => {
  const auth = inject(AuthService);
  const runtimeConfig = inject(RuntimeConfigService);

  const token = auth.accessToken;
  if (!token) return next(req);

  // Only attach token to controller and observability backends — not external CDNs
  const controllerBase = runtimeConfig.controllerBase;
  const observabilityBases = [
    runtimeConfig.prometheusBase,
    runtimeConfig.grafanaBase,
    runtimeConfig.jaegerBase,
    runtimeConfig.logsBase,
  ];

  const isProtected =
    req.url.startsWith(controllerBase) ||
    observabilityBases.some((base) => req.url.startsWith(base));

  if (!isProtected) return next(req);

  return next(
    req.clone({
      setHeaders: { Authorization: `Bearer ${token}` },
    })
  );
};
