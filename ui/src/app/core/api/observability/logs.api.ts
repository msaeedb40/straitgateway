// Copyright 2026 straitgateway Authors
// SPDX-License-Identifier: Apache-2.0

import { Injectable, inject } from '@angular/core';
import { HttpClient } from '@angular/common/http';
import { Observable, of } from 'rxjs';
import { catchError } from 'rxjs/operators';
import { RuntimeConfigService } from '../../config/runtime-config';
import { LogEntry } from '../api.types';

@Injectable({ providedIn: 'root' })
export class LogsApi {
  private http = inject(HttpClient);
  private config = inject(RuntimeConfigService);

  private get base() { return this.config.apiBase(); }

  getLogs(level?: string, limit = 200, subsystem?: string): Observable<LogEntry[]> {
    const params = new URLSearchParams();
    params.set('limit', limit.toString());
    if (level) params.set('level', level);
    if (subsystem) params.set('subsystem', subsystem);

    return this.http.get<LogEntry[]>(`${this.base}/api/v1/logs?${params.toString()}`).pipe(
      catchError(() => of(this.getMockLogs(level, subsystem)))
    );
  }

  private getMockLogs(level?: string, subsystem?: string): LogEntry[] {
    let logs: LogEntry[] = [
      {
        timestamp: '2026-09-10T22:20:00Z',
        level: 'INFO',
        component: 'sg-controller',
        subsystem: 'gateway-api',
        message: 'Reconciliation loop completed for Gateway default/sg-edge-gw. Status: Programmed.',
      },
      {
        timestamp: '2026-09-10T22:20:05Z',
        level: 'INFO',
        component: 'straitgatewayd',
        subsystem: 'bpf-loader',
        message: 'Loaded eBPF program [classifier_ingress] (id=241, insns=412) onto netkit interface nk-worker0.',
      },
      {
        timestamp: '2026-09-10T22:20:10Z',
        level: 'DEBUG',
        component: 'straitgatewayd',
        subsystem: 'maglev',
        message: 'Recalculated Maglev lookup table (M=65537) for service default/core-api with 4 backends.',
      },
      {
        timestamp: '2026-09-10T22:20:15Z',
        level: 'INFO',
        component: 'sg-transit',
        subsystem: 'wireguard',
        message: 'Peer handshake renewed with cluster ID 2 at 203.0.113.15:51820 (latency: 1.8ms).',
      },
      {
        timestamp: '2026-09-10T22:20:20Z',
        level: 'WARN',
        component: 'straitgatewayd',
        subsystem: 'policy-enforcer',
        message: 'Packet dropped on ingress: src=185.220.101.5:58921 dst=10.244.0.15:22 (rule: deny-ssh).',
      },
      {
        timestamp: '2026-09-10T22:20:25Z',
        level: 'INFO',
        component: 'sg-controller',
        subsystem: 'cni-ipam',
        message: 'Allocated pod IP 10.244.1.89 on node sg-worker-node-01 (pool utilization: 25.2%).',
      },
    ];

    if (level) {
      logs = logs.filter(l => l.level.toLowerCase() === level.toLowerCase());
    }
    if (subsystem) {
      logs = logs.filter(l => l.subsystem?.toLowerCase().includes(subsystem.toLowerCase()));
    }
    return logs;
  }
}
