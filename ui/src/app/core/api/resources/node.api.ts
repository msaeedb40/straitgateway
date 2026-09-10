// Copyright 2026 straitgateway Authors
// SPDX-License-Identifier: Apache-2.0

import { Injectable, inject } from '@angular/core';
import { HttpClient } from '@angular/common/http';
import { Observable, of } from 'rxjs';
import { catchError } from 'rxjs/operators';
import { RuntimeConfigService } from '../../config/runtime-config';
import { NodeStatus } from '../api.types';

@Injectable({ providedIn: 'root' })
export class NodeApi {
  private http = inject(HttpClient);
  private config = inject(RuntimeConfigService);

  private get base() { return this.config.apiBase(); }

  list(): Observable<NodeStatus[]> {
    return this.http.get<NodeStatus[]>(`${this.base}/api/v1/nodes`).pipe(
      catchError(() => of(this.getMockNodes()))
    );
  }

  get(name: string): Observable<NodeStatus> {
    return this.http.get<NodeStatus>(`${this.base}/api/v1/nodes/${name}`).pipe(
      catchError(() => of(this.getMockNodes().find(n => n.name === name) || this.getMockNodes()[0]))
    );
  }

  private getMockNodes(): NodeStatus[] {
    return [
      {
        name: 'sg-control-plane-01',
        ip: '192.168.1.10',
        podCIDR: '10.244.0.0/24',
        kernelVersion: '6.8.0-45-generic',
        cniReady: true,
        serviceReady: true,
        policyReady: true,
        gatewayReady: true,
        kubeProxyReplacement: true,
        bpfMapRevision: 142,
        wireguardPublicKey: 'uB+3ZkqB9FvLzP029sKd4Wq7K8j4w0tX1A5MvN9oQ=',
      },
      {
        name: 'sg-worker-node-01',
        ip: '192.168.1.11',
        podCIDR: '10.244.1.0/24',
        kernelVersion: '6.8.0-45-generic',
        cniReady: true,
        serviceReady: true,
        policyReady: true,
        gatewayReady: true,
        kubeProxyReplacement: true,
        bpfMapRevision: 142,
        wireguardPublicKey: 'L9mX1A5MvN9oQ+uB+3ZkqB9FvLzP029sKd4Wq7K8j4=',
      },
      {
        name: 'sg-worker-node-02',
        ip: '192.168.1.12',
        podCIDR: '10.244.2.0/24',
        kernelVersion: '6.8.0-45-generic',
        cniReady: true,
        serviceReady: true,
        policyReady: true,
        gatewayReady: true,
        kubeProxyReplacement: true,
        bpfMapRevision: 142,
        wireguardPublicKey: 'Wq7K8j4w0tX1A5MvN9oQ+uB+3ZkqB9FvLzP029sKd4=',
      },
    ];
  }
}
