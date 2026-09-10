// Copyright 2026 straitgateway Authors
// SPDX-License-Identifier: Apache-2.0

import { Injectable, inject } from '@angular/core';
import { HttpClient } from '@angular/common/http';
import { Observable, of } from 'rxjs';
import { catchError } from 'rxjs/operators';
import { RuntimeConfigService } from '../../config/runtime-config';
import { CniNodeStatus } from '../api.types';

@Injectable({ providedIn: 'root' })
export class CniApi {
  private http = inject(HttpClient);
  private config = inject(RuntimeConfigService);

  private get base() { return this.config.apiBase(); }

  listNodes(): Observable<CniNodeStatus[]> {
    return this.http.get<CniNodeStatus[]>(`${this.base}/api/v1/cni/nodes`).pipe(
      catchError(() => of(this.getMockCniNodes()))
    );
  }

  private getMockCniNodes(): CniNodeStatus[] {
    return [
      {
        nodeName: 'sg-control-plane-01',
        podCIDR: '10.244.0.0/24',
        assignedIPs: 18,
        totalIPs: 254,
        netkitInterface: 'nk-cplane0',
        mtu: 1500,
        status: 'Ready',
      },
      {
        nodeName: 'sg-worker-node-01',
        podCIDR: '10.244.1.0/24',
        assignedIPs: 64,
        totalIPs: 254,
        netkitInterface: 'nk-worker0',
        mtu: 1500,
        status: 'Ready',
      },
      {
        nodeName: 'sg-worker-node-02',
        podCIDR: '10.244.2.0/24',
        assignedIPs: 49,
        totalIPs: 254,
        netkitInterface: 'nk-worker1',
        mtu: 1500,
        status: 'Ready',
      },
    ];
  }
}
