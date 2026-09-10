// Copyright 2026 straitgateway Authors
// SPDX-License-Identifier: Apache-2.0

import { Injectable, inject, signal } from '@angular/core';
import { HttpClient } from '@angular/common/http';
import { RuntimeConfigService } from '../config/runtime-config';
import { catchError, of, tap } from 'rxjs';

/**
 * NamespaceService manages the active namespace context across all feature views.
 * Default namespace is 'default'. All resources are scoped to custom or default namespace.
 */
@Injectable({ providedIn: 'root' })
export class NamespaceService {
  private http = inject(HttpClient);
  private config = inject(RuntimeConfigService);

  /** All available namespaces in the cluster. */
  namespaces = signal<string[]>([
    'default',
    'kube-system',
    'straitgateway-system',
    'ingress-nginx',
    'monitoring',
  ]);

  /** The currently selected namespace. 'default' by default, or '' for all. */
  active = signal<string>('default');

  constructor() {
    this.refresh();
  }

  /** Load namespaces from the API. */
  refresh() {
    const base = typeof this.config.apiBase === 'function' ? this.config.apiBase() : this.config.get('apiBase');
    this.http.get<string[]>(`${base}/api/v1/namespaces`).pipe(
      catchError(() => of(['default', 'kube-system', 'straitgateway-system', 'ingress-nginx', 'monitoring'])),
      tap(ns => this.namespaces.set(ns)),
    ).subscribe();
  }

  /** Set the active namespace. Pass empty string for all namespaces. */
  select(ns: string) {
    this.active.set(ns);
  }

  /** Returns query param string for the active namespace. */
  queryParam(): string {
    const ns = this.active();
    return ns ? `?namespace=${encodeURIComponent(ns)}` : '';
  }
}
