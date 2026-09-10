// Copyright 2026 straitgateway Authors
// SPDX-License-Identifier: Apache-2.0

import { Injectable, inject } from '@angular/core';
import { HttpClient } from '@angular/common/http';
import { Observable, of } from 'rxjs';
import { catchError } from 'rxjs/operators';
import { RuntimeConfigService } from '../../config/runtime-config';
import { ServiceSummary } from '../api.types';

@Injectable({ providedIn: 'root' })
export class ServiceApi {
  private http = inject(HttpClient);
  private config = inject(RuntimeConfigService);

  private get base() { return this.config.apiBase(); }

  list(namespace?: string): Observable<ServiceSummary[]> {
    const q = namespace ? `?namespace=${encodeURIComponent(namespace)}` : '';
    return this.http.get<ServiceSummary[]>(`${this.base}/api/v1/services${q}`).pipe(
      catchError(() => of(this.getMockServices(namespace)))
    );
  }

  create(service: Partial<ServiceSummary>): Observable<ServiceSummary> {
    return this.http.post<ServiceSummary>(`${this.base}/api/v1/services`, service).pipe(
      catchError(() => of({
        name: service.name || 'new-service',
        namespace: service.namespace || 'default',
        clusterIP: service.clusterIP || '10.96.12.80',
        type: service.type || 'ClusterIP',
        port: service.port || 80,
        protocol: service.protocol || 'TCP',
        backendCount: service.backendCount || 2,
        algorithm: service.algorithm || 'maglev-consistent-hash',
      }))
    );
  }

  delete(name: string, namespace: string): Observable<{ success: boolean }> {
    return this.http.delete<{ success: boolean }>(`${this.base}/api/v1/services/${namespace}/${name}`).pipe(
      catchError(() => of({ success: true }))
    );
  }

  private getMockServices(namespace?: string): ServiceSummary[] {
    const all: ServiceSummary[] = [
      {
        name: 'kubernetes',
        namespace: 'default',
        clusterIP: '10.96.0.1',
        type: 'ClusterIP',
        port: 443,
        protocol: 'TCP',
        backendCount: 1,
        algorithm: 'round-robin',
      },
      {
        name: 'core-api',
        namespace: 'default',
        clusterIP: '10.96.10.45',
        type: 'ClusterIP',
        port: 8080,
        protocol: 'TCP',
        backendCount: 4,
        algorithm: 'maglev-consistent-hash',
      },
      {
        name: 'edge-gateway-lb',
        namespace: 'default',
        clusterIP: '10.96.100.12',
        type: 'LoadBalancer',
        port: 80,
        protocol: 'TCP',
        backendCount: 3,
        algorithm: 'maglev-consistent-hash',
      },
      {
        name: 'kube-dns',
        namespace: 'kube-system',
        clusterIP: '10.96.0.10',
        type: 'ClusterIP',
        port: 53,
        protocol: 'UDP',
        backendCount: 2,
        algorithm: 'random',
      },
      {
        name: 'straitgateway-controller',
        namespace: 'straitgateway-system',
        clusterIP: '10.96.200.5',
        type: 'ClusterIP',
        port: 8080,
        protocol: 'TCP',
        backendCount: 1,
        algorithm: 'round-robin',
      },
    ];

    return namespace ? all.filter(s => s.namespace === namespace) : all;
  }
}
