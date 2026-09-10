// Copyright 2026 straitgateway Authors
// SPDX-License-Identifier: Apache-2.0

import { Injectable, inject } from '@angular/core';
import { HttpClient } from '@angular/common/http';
import { Observable, of } from 'rxjs';
import { catchError } from 'rxjs/operators';
import { RuntimeConfigService } from '../../config/runtime-config';
import { TraceSpan } from '../api.types';

@Injectable({ providedIn: 'root' })
export class JaegerApi {
  private http = inject(HttpClient);
  private config = inject(RuntimeConfigService);

  private get base() { return this.config.jaegerBase(); }

  listTraces(service = 'straitgateway-controller', limit = 50): Observable<TraceSpan[]> {
    return this.http.get<TraceSpan[]>(`${this.base}/api/traces?service=${service}&limit=${limit}`).pipe(
      catchError(() => of(this.getMockTraces()))
    );
  }

  getTrace(traceId: string): Observable<TraceSpan[]> {
    return this.http.get<TraceSpan[]>(`${this.base}/api/traces/${traceId}`).pipe(
      catchError(() => of(this.getMockTraces().filter(t => t.traceId === traceId)))
    );
  }

  private getMockTraces(): TraceSpan[] {
    return [
      {
        traceId: '7a19c5b204e13d9f',
        spanId: '101',
        operationName: 'eBPF::ProcessIngressPacket',
        serviceName: 'straitgatewayd',
        durationMs: 0.12,
        startTime: '12:44:02.105',
        status: 'ok',
        tags: { interface: 'nk-worker0', proto: 'TCP', dstPort: '443' },
      },
      {
        traceId: '7a19c5b204e13d9f',
        spanId: '102',
        operationName: 'Maglev::BackendLookup',
        serviceName: 'straitgatewayd',
        durationMs: 0.04,
        startTime: '12:44:02.105',
        status: 'ok',
        tags: { service: 'default/core-api', backend: '10.244.1.25:8080' },
      },
      {
        traceId: '9d38e21a00fc24a8',
        spanId: '201',
        operationName: 'TransitGateway::RouteLookup',
        serviceName: 'straitgateway-transit',
        durationMs: 0.28,
        startTime: '12:44:00.320',
        status: 'ok',
        tags: { targetCluster: '2', tunnel: 'wg-transit' },
      },
      {
        traceId: '4b61aa91d09e88cb',
        spanId: '301',
        operationName: 'PolicyCompiler::CompileCRD',
        serviceName: 'straitgateway-controller',
        durationMs: 4.8,
        startTime: '12:43:55.000',
        status: 'ok',
        tags: { policy: 'default/deny-ssh', rules: '1' },
      },
    ];
  }
}
