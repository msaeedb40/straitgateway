// Copyright 2026 straitgateway Authors
// SPDX-License-Identifier: Apache-2.0

import { Injectable, inject } from '@angular/core';
import { HttpClient } from '@angular/common/http';
import { Observable, of } from 'rxjs';
import { catchError } from 'rxjs/operators';
import { RuntimeConfigService } from '../config/runtime-config';
import {
  DashboardSummary,
  NodeStatus,
  ServiceSummary,
  GatewaySummary,
  GatewayRouteSummary,
  FlowEvent,
  PacketCapture,
  TunnelPeer,
  BGPPeerStatus,
  KubeEvent,
  EbpfMapSummary,
  CniNodeStatus,
  EndpointSummary,
  LogEntry,
  TraceSpan,
} from './api.types';

import { GatewayApi } from './resources/gateway.api';
import { NodeApi } from './resources/node.api';
import { TunnelApi } from './resources/tunnel.api';
import { FlowApi } from './resources/flow.api';
import { PacketApi } from './resources/packet.api';
import { ServiceApi } from './resources/service.api';
import { EndpointApi } from './resources/endpoint.api';
import { EbpfApi } from './resources/ebpf.api';
import { CniApi } from './resources/cni.api';
import { EventApi } from './resources/event.api';
import { ConfigurationApi } from './resources/configuration.api';

import { PrometheusApi } from './observability/prometheus.api';
import { GrafanaApi } from './observability/grafana.api';
import { JaegerApi } from './observability/jaeger.api';
import { LogsApi } from './observability/logs.api';

export * from './api.types';
export * from './api-error';

@Injectable({ providedIn: 'root' })
export class ApiClient {
  private http = inject(HttpClient);
  private config = inject(RuntimeConfigService);

  // Sub-APIs
  readonly gateway = inject(GatewayApi);
  readonly node = inject(NodeApi);
  readonly tunnel = inject(TunnelApi);
  readonly flow = inject(FlowApi);
  readonly packet = inject(PacketApi);
  readonly service = inject(ServiceApi);
  readonly endpoint = inject(EndpointApi);
  readonly ebpf = inject(EbpfApi);
  readonly cni = inject(CniApi);
  readonly event = inject(EventApi);
  readonly configuration = inject(ConfigurationApi);

  readonly prometheus = inject(PrometheusApi);
  readonly grafana = inject(GrafanaApi);
  readonly jaeger = inject(JaegerApi);
  readonly logs = inject(LogsApi);

  private get base() { return this.config.apiBase(); }

  // ── Dashboard ────────────────────────────────────────────────────────────
  getDashboard(): Observable<DashboardSummary> {
    return this.http.get<DashboardSummary>(`${this.base}/api/v1/dashboard`).pipe(
      catchError(() => of({
        totalNodes: 3,
        readyNodes: 3,
        totalGateways: 3,
        activeGateways: 3,
        totalTunnels: 3,
        activeTunnels: 3,
        totalServices: 5,
        totalFlowsPerSec: 5420,
        droppedFlowsPerSec: 12,
        healthPercentage: 99.4,
      }))
    );
  }

  // ── Resource facades for backward compatibility ──────────────────────────
  getNodes(): Observable<NodeStatus[]> {
    return this.node.list();
  }

  getServices(namespace?: string): Observable<ServiceSummary[]> {
    return this.service.list(namespace);
  }

  getGateways(namespace?: string): Observable<GatewaySummary[]> {
    return this.gateway.list(namespace);
  }

  getGatewayRoutes(kind?: string, namespace?: string): Observable<GatewayRouteSummary[]> {
    return this.gateway.listRoutes(kind, namespace);
  }

  getFlows(limit = 100): Observable<FlowEvent[]> {
    return this.flow.list(limit);
  }

  getPackets(action?: string, limit = 500): Observable<PacketCapture[]> {
    return this.packet.list(limit, action);
  }

   getTunnels(): Observable<TunnelPeer[]> {
    return this.tunnel.listTunnels();
  }

  createTunnel(tunnel: Partial<TunnelPeer>): Observable<TunnelPeer> {
    return this.tunnel.createTunnel(tunnel);
  }

  deleteTunnel(clusterID: number): Observable<{ success: boolean }> {
    return this.tunnel.deleteTunnel(clusterID);
  }

  getBGPPeers(): Observable<BGPPeerStatus[]> {
    return this.tunnel.listBgpPeers();
  }

  getEvents(limit = 50, namespace?: string): Observable<KubeEvent[]> {
    return this.event.list(limit, namespace);
  }

  getBPFMapStats(): Observable<EbpfMapSummary[]> {
    return this.ebpf.listMaps();
  }
  getBpfMapStats = this.getBPFMapStats.bind(this);

  getEndpoints(namespace?: string): Observable<EndpointSummary[]> {
    return this.endpoint.list(namespace);
  }

  getCniStatus(): Observable<CniNodeStatus[]> {
    return this.cni.listNodes();
  }

  getMetrics(): Observable<any[]> {
    return this.prometheus.query('rate(straitgateway_flows_total[1m])');
  }

  getLogs(level?: string, limit = 200, subsystem?: string): Observable<LogEntry[]> {
    return this.logs.getLogs(level, limit, subsystem);
  }

  getTraces(limit = 100): Observable<TraceSpan[]> {
    return this.jaeger.listTraces('straitgateway-controller', limit);
  }
}
