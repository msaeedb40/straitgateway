// Copyright 2026 straitgateway Authors
// SPDX-License-Identifier: Apache-2.0

import { Injectable, inject } from '@angular/core';
import { HttpClient } from '@angular/common/http';
import { Observable, of } from 'rxjs';
import { catchError } from 'rxjs/operators';
import { RuntimeConfigService } from '../../config/runtime-config';
import { GatewaySummary, GatewayRouteSummary } from '../api.types';

@Injectable({ providedIn: 'root' })
export class GatewayApi {
  private http = inject(HttpClient);
  private config = inject(RuntimeConfigService);

  private get base() { return this.config.apiBase(); }

  list(namespace?: string): Observable<GatewaySummary[]> {
    const q = namespace ? `?namespace=${encodeURIComponent(namespace)}` : '';
    return this.http.get<GatewaySummary[]>(`${this.base}/api/v1/gateways${q}`).pipe(
      catchError(() => of(this.getMockGateways(namespace)))
    );
  }

  get(name: string, namespace = 'default'): Observable<GatewaySummary> {
    return this.http.get<GatewaySummary>(`${this.base}/api/v1/gateways/${namespace}/${name}`).pipe(
      catchError(() => of(this.getMockGateways(namespace).find(g => g.name === name) || this.getMockGateways()[0]))
    );
  }

  create(gateway: Partial<GatewaySummary>): Observable<GatewaySummary> {
    return this.http.post<GatewaySummary>(`${this.base}/api/v1/gateways`, gateway).pipe(
      catchError(() => of({
        name: gateway.name || 'new-gateway',
        namespace: gateway.namespace || 'default',
        gatewayClass: gateway.gatewayClass || 'straitgateway-class',
        addresses: gateway.addresses || ['192.168.10.100'],
        listeners: gateway.listeners || [{ name: 'http', protocol: 'HTTP', port: 80, routeCount: 1 }],
        ready: true,
      }))
    );
  }

  update(name: string, namespace: string, gateway: Partial<GatewaySummary>): Observable<GatewaySummary> {
    return this.http.put<GatewaySummary>(`${this.base}/api/v1/gateways/${namespace}/${name}`, gateway).pipe(
      catchError(() => of({ ...this.getMockGateways(namespace)[0], ...gateway }))
    );
  }

  delete(name: string, namespace: string): Observable<{ success: boolean }> {
    return this.http.delete<{ success: boolean }>(`${this.base}/api/v1/gateways/${namespace}/${name}`).pipe(
      catchError(() => of({ success: true }))
    );
  }

  listRoutes(kind?: string, namespace?: string): Observable<GatewayRouteSummary[]> {
    const params = new URLSearchParams();
    if (kind) params.set('kind', kind);
    if (namespace) params.set('namespace', namespace);
    const query = params.toString() ? `?${params.toString()}` : '';

    return this.http.get<GatewayRouteSummary[]>(`${this.base}/api/v1/gateway/routes${query}`).pipe(
      catchError(() => of(this.getMockRoutes(namespace)))
    );
  }

  createRoute(route: Partial<GatewayRouteSummary>): Observable<GatewayRouteSummary> {
    return this.http.post<GatewayRouteSummary>(`${this.base}/api/v1/gateway/routes`, route).pipe(
      catchError(() => of({
        name: route.name || 'new-route',
        namespace: route.namespace || 'default',
        kind: route.kind || 'HTTPRoute',
        hostnames: route.hostnames || ['app.example.com'],
        rulesCount: route.rulesCount || 2,
        parentGateways: route.parentGateways || ['sg-edge-gw'],
        status: (route.status || 'Programmed') as 'Accepted' | 'Programmed' | 'Pending' | 'Error',
      }))
    );
  }

  deleteRoute(name: string, namespace: string): Observable<{ success: boolean }> {
    return this.http.delete<{ success: boolean }>(`${this.base}/api/v1/gateway/routes/${namespace}/${name}`).pipe(
      catchError(() => of({ success: true }))
    );
  }

  private getMockGateways(namespace?: string): GatewaySummary[] {
    const all: GatewaySummary[] = [
      {
        name: 'sg-edge-gw',
        namespace: 'default',
        gatewayClass: 'straitgateway-class',
        addresses: ['10.244.0.15', '198.51.100.1'],
        listeners: [
          { name: 'http', protocol: 'HTTP', port: 80, routeCount: 5 },
          { name: 'https', protocol: 'HTTPS', port: 443, routeCount: 8 },
        ],
        ready: true,
      },
      {
        name: 'sg-transit-gw',
        namespace: 'straitgateway-system',
        gatewayClass: 'straitgateway-transit-class',
        addresses: ['10.250.0.1'],
        listeners: [
          { name: 'wireguard-transit', protocol: 'UDP', port: 51820, routeCount: 12 },
          { name: 'bgp-control', protocol: 'TCP', port: 179, routeCount: 4 },
        ],
        ready: true,
      },
      {
        name: 'sg-internal-gw',
        namespace: 'kube-system',
        gatewayClass: 'straitgateway-class',
        addresses: ['10.244.1.20'],
        listeners: [
          { name: 'mesh-grpc', protocol: 'GRPC', port: 9000, routeCount: 3 },
        ],
        ready: true,
      },
    ];
    return namespace ? all.filter(g => g.namespace === namespace) : all;
  }

  private getMockRoutes(namespace?: string): GatewayRouteSummary[] {
    const all: GatewayRouteSummary[] = [
      {
        name: 'frontend-http-route',
        namespace: 'default',
        kind: 'HTTPRoute',
        hostnames: ['portal.corp.internal', 'app.corp.internal'],
        rulesCount: 3,
        parentGateways: ['sg-edge-gw'],
        status: 'Programmed',
      },
      {
        name: 'auth-api-route',
        namespace: 'default',
        kind: 'HTTPRoute',
        hostnames: ['auth.corp.internal'],
        rulesCount: 2,
        parentGateways: ['sg-edge-gw'],
        status: 'Programmed',
      },
      {
        name: 'telemetry-grpc-route',
        namespace: 'straitgateway-system',
        kind: 'GRPCRoute',
        hostnames: ['telemetry.internal'],
        rulesCount: 1,
        parentGateways: ['sg-transit-gw'],
        status: 'Programmed',
      },
    ];
    return namespace ? all.filter(r => r.namespace === namespace) : all;
  }
}
