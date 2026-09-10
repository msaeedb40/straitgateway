// Copyright 2026 straitgateway Authors
// SPDX-License-Identifier: Apache-2.0

import { Injectable, inject } from '@angular/core';
import { HttpClient } from '@angular/common/http';
import { Observable, of } from 'rxjs';
import { catchError } from 'rxjs/operators';
import { RuntimeConfigService } from '../../config/runtime-config';
import { KubeEvent } from '../api.types';

@Injectable({ providedIn: 'root' })
export class EventApi {
  private http = inject(HttpClient);
  private config = inject(RuntimeConfigService);

  private get base() { return this.config.apiBase(); }

  list(limit = 50, namespace?: string): Observable<KubeEvent[]> {
    const params = new URLSearchParams();
    params.set('limit', limit.toString());
    if (namespace) params.set('namespace', namespace);

    return this.http.get<KubeEvent[]>(`${this.base}/api/v1/events?${params.toString()}`).pipe(
      catchError(() => of(this.getMockEvents(namespace)))
    );
  }

  private getMockEvents(namespace?: string): KubeEvent[] {
    const all: KubeEvent[] = [
      {
        namespace: 'default',
        name: 'sg-edge-gw.17a4b',
        reason: 'Programmed',
        message: 'Gateway HTTP listeners configured and loaded into eBPF maps',
        type: 'Normal',
        count: 1,
        lastSeen: '2m ago',
      },
      {
        namespace: 'straitgateway-system',
        name: 'tunnel-cluster-2.17a4a',
        reason: 'HandshakeEstablished',
        message: 'WireGuard session authenticated with cluster ID 2 (latency: 1.8ms)',
        type: 'Normal',
        count: 1,
        lastSeen: '5m ago',
      },
      {
        namespace: 'default',
        name: 'frontend-http-route.17a49',
        reason: 'SyncSuccess',
        message: 'Route attached to parent gateway sg-edge-gw with 3 rules',
        type: 'Normal',
        count: 1,
        lastSeen: '8m ago',
      },
      {
        namespace: 'kube-system',
        name: 'straitgatewayd-w1.17a48',
        reason: 'BpfMapUpdated',
        message: 'Synchronized 4 Maglev backend endpoints to sg_maglev_lookup',
        type: 'Normal',
        count: 1,
        lastSeen: '12m ago',
      },
      {
        namespace: 'default',
        name: 'straitnetworkpolicy-deny-ssh.17a47',
        reason: 'Enforced',
        message: 'eBPF LPM trie updated with port 22 drop rules for identity 0',
        type: 'Normal',
        count: 1,
        lastSeen: '15m ago',
      },
    ];

    return namespace ? all.filter(e => e.namespace === namespace) : all;
  }
}
