// Copyright 2026 straitgateway Authors
// SPDX-License-Identifier: Apache-2.0

import { Injectable, inject } from '@angular/core';
import { HttpClient } from '@angular/common/http';
import { Observable, of } from 'rxjs';
import { catchError } from 'rxjs/operators';
import { RuntimeConfigService } from '../../config/runtime-config';
import { EndpointSummary } from '../api.types';

@Injectable({ providedIn: 'root' })
export class EndpointApi {
  private http = inject(HttpClient);
  private config = inject(RuntimeConfigService);

  private get base() { return this.config.apiBase(); }

  list(namespace?: string): Observable<EndpointSummary[]> {
    const q = namespace ? `?namespace=${encodeURIComponent(namespace)}` : '';
    return this.http.get<EndpointSummary[]>(`${this.base}/api/v1/endpoints${q}`).pipe(
      catchError(() => of(this.getMockEndpoints(namespace)))
    );
  }

  private getMockEndpoints(namespace?: string): EndpointSummary[] {
    const all: EndpointSummary[] = [
      {
        name: 'core-api-slice-0',
        namespace: 'default',
        serviceName: 'core-api',
        addressType: 'IPv4',
        readyEndpoints: [
          { ip: '10.244.1.25', nodeName: 'sg-worker-node-01', port: 8080 },
          { ip: '10.244.1.26', nodeName: 'sg-worker-node-01', port: 8080 },
          { ip: '10.244.2.14', nodeName: 'sg-worker-node-02', port: 8080 },
        ],
        notReadyEndpoints: [],
      },
      {
        name: 'edge-gateway-lb-slice-0',
        namespace: 'default',
        serviceName: 'edge-gateway-lb',
        addressType: 'IPv4',
        readyEndpoints: [
          { ip: '10.244.0.15', nodeName: 'sg-control-plane-01', port: 80 },
          { ip: '10.244.1.18', nodeName: 'sg-worker-node-01', port: 80 },
        ],
      },
    ];

    return namespace ? all.filter(e => e.namespace === namespace) : all;
  }
}
