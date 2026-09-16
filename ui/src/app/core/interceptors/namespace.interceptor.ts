import { HttpInterceptorFn } from '@angular/common/http';
import { inject } from '@angular/core';
import { ContextService } from '../services/context.service';
import { RuntimeConfigService } from '../config/runtime-config.service';

export const namespaceInterceptor: HttpInterceptorFn = (req, next) => {
  const context = inject(ContextService);
  const runtimeConfig = inject(RuntimeConfigService);

  // Only inject context headers on controller API requests
  if (!req.url.startsWith(runtimeConfig.controllerBase)) return next(req);

  const cluster = context.selectedCluster();
  const namespace = context.selectedNamespace();

  const headers: Record<string, string> = {};
  if (cluster) headers['X-Cluster'] = cluster;
  if (namespace) headers['X-Namespace'] = namespace;

  if (Object.keys(headers).length === 0) return next(req);

  return next(req.clone({ setHeaders: headers }));
};
