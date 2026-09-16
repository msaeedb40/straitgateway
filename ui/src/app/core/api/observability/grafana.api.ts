import { inject, Injectable } from '@angular/core';
import { Observable } from 'rxjs';
import { ApiClient } from '../api-client';
import { RuntimeConfigService } from '../../config/runtime-config.service';

export interface GrafanaDashboard {
  readonly uid: string;
  readonly title: string;
  readonly url: string;
  readonly tags: string[];
}

@Injectable({ providedIn: 'root' })
export class GrafanaApiService {
  private readonly client = inject(ApiClient);
  private readonly runtimeConfig = inject(RuntimeConfigService);

  listDashboards(): Observable<GrafanaDashboard[]> {
    return this.client.external<GrafanaDashboard[]>(
      this.runtimeConfig.grafanaBase,
      '/api/search?type=dash-db'
    );
  }

  /** Returns the deep-link URL to a specific dashboard, optionally scoped to time range */
  dashboardUrl(uid: string, params?: { from?: string; to?: string; vars?: Record<string, string> }): string {
    const base = `${this.runtimeConfig.grafanaBase}/d/${uid}`;
    const qs = new URLSearchParams();
    if (params?.from) qs.set('from', params.from);
    if (params?.to) qs.set('to', params.to);
    for (const [k, v] of Object.entries(params?.vars ?? {})) {
      qs.set(`var-${k}`, v);
    }
    const query = qs.toString();
    return query ? `${base}?${query}` : base;
  }

  healthCheck(): Observable<boolean> {
    return new Observable((observer) => {
      this.client
        .external<unknown>(this.runtimeConfig.grafanaBase, '/api/health')
        .subscribe({
          next: () => { observer.next(true); observer.complete(); },
          error: () => { observer.next(false); observer.complete(); },
        });
    });
  }
}
