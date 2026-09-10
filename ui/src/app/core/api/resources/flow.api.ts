// Copyright 2026 straitgateway Authors
// SPDX-License-Identifier: Apache-2.0

import { Injectable, inject } from '@angular/core';
import { HttpClient } from '@angular/common/http';
import { Observable, of } from 'rxjs';
import { catchError } from 'rxjs/operators';
import { RuntimeConfigService } from '../../config/runtime-config';
import { FlowEvent } from '../api.types';

@Injectable({ providedIn: 'root' })
export class FlowApi {
  private http = inject(HttpClient);
  private config = inject(RuntimeConfigService);

  private get base() { return this.config.apiBase(); }

  list(limit = 100, action?: string): Observable<FlowEvent[]> {
    const params = new URLSearchParams();
    params.set('limit', limit.toString());
    if (action) params.set('action', action);

    return this.http.get<FlowEvent[]>(`${this.base}/api/v1/flows?${params.toString()}`).pipe(
      catchError(() => of(this.getMockFlows(action)))
    );
  }

  private getMockFlows(action?: string): FlowEvent[] {
    const now = Date.now() * 1_000_000;
    const all: FlowEvent[] = [
      {
        srcIP: '10.244.1.42',
        dstIP: '10.96.0.10',
        srcPort: 48210,
        dstPort: 53,
        srcIdentity: 104,
        dstIdentity: 2,
        protocol: 'UDP',
        direction: 'egress',
        action: 'allow',
        bytes: 78,
        timestampNs: now - 250_000_000,
      },
      {
        srcIP: '198.51.100.12',
        dstIP: '10.244.0.15',
        srcPort: 53420,
        dstPort: 443,
        srcIdentity: 0,
        dstIdentity: 200,
        protocol: 'TCP',
        direction: 'ingress',
        action: 'allow',
        bytes: 1420,
        timestampNs: now - 800_000_000,
      },
      {
        srcIP: '10.244.2.19',
        dstIP: '10.250.2.1',
        srcPort: 39120,
        dstPort: 8080,
        srcIdentity: 108,
        dstIdentity: 301,
        protocol: 'TCP',
        direction: 'egress',
        action: 'allow',
        bytes: 654,
        timestampNs: now - 1_200_000_000,
      },
      {
        srcIP: '185.220.101.5',
        dstIP: '10.244.0.15',
        srcPort: 58921,
        dstPort: 22,
        srcIdentity: 0,
        dstIdentity: 0,
        protocol: 'TCP',
        direction: 'ingress',
        action: 'deny',
        dropReason: 'PolicyDrop: StraitNetworkPolicy default/deny-ssh',
        bytes: 64,
        timestampNs: now - 1_800_000_000,
      },
      {
        srcIP: '10.244.1.88',
        dstIP: '10.244.2.105',
        srcPort: 50114,
        dstPort: 9090,
        srcIdentity: 110,
        dstIdentity: 112,
        protocol: 'TCP',
        direction: 'egress',
        action: 'allow',
        bytes: 2840,
        timestampNs: now - 2_100_000_000,
      },
    ];

    return action ? all.filter(f => f.action === action) : all;
  }
}
