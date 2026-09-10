// Copyright 2026 straitgateway Authors
// SPDX-License-Identifier: Apache-2.0

import { Routes } from '@angular/router';

export const routes: Routes = [
  {
    path: '',
    loadComponent: () => import('./features/dashboard/dashboard.component').then(m => m.DashboardComponent),
    title: 'Dashboard — straitgateway',
  },
  {
    path: 'topology',
    loadComponent: () => import('./features/topology/topology.component').then(m => m.TopologyComponent),
    title: 'Topology — straitgateway',
  },
  {
    path: 'gateways',
    loadComponent: () => import('./features/gateways/gateways.component').then(m => m.GatewaysComponent),
    title: 'Gateways — straitgateway',
  },
  {
    path: 'routes',
    loadComponent: () => import('./features/gateway-routes/gateway-routes.component').then(m => m.GatewayRoutesComponent),
    title: 'Routes — straitgateway',
  },
  {
    path: 'nodes',
    loadComponent: () => import('./features/nodes/nodes.component').then(m => m.NodesComponent),
    title: 'Nodes — straitgateway',
  },
  {
    path: 'services',
    loadComponent: () => import('./features/services/services.component').then(m => m.ServicesComponent),
    title: 'Services — straitgateway',
  },
  {
    path: 'endpoints',
    loadComponent: () => import('./features/endpoints/endpoints.component').then(m => m.EndpointsComponent),
    title: 'Endpoints — straitgateway',
  },
  {
    path: 'tunnels',
    loadComponent: () => import('./features/tunnels/tunnels.component').then(m => m.TunnelsComponent),
    title: 'Tunnels — straitgateway',
  },
  {
    path: 'ebpf',
    loadComponent: () => import('./features/ebpf/ebpf.component').then(m => m.EbpfComponent),
    title: 'eBPF Maps — straitgateway',
  },
  {
    path: 'cni',
    loadComponent: () => import('./features/cni/cni.component').then(m => m.CniComponent),
    title: 'CNI — straitgateway',
  },
  {
    path: 'flows',
    loadComponent: () => import('./features/flows/flows.component').then(m => m.FlowsComponent),
    title: 'Flows — straitgateway',
  },
  {
    path: 'packets',
    loadComponent: () => import('./features/packets/packets.component').then(m => m.PacketsComponent),
    title: 'Packets — straitgateway',
  },
  {
    path: 'events',
    loadComponent: () => import('./features/events/events.component').then(m => m.EventsComponent),
    title: 'Events — straitgateway',
  },
  {
    path: 'logs',
    loadComponent: () => import('./features/logs/logs.component').then(m => m.LogsComponent),
    title: 'Logs — straitgateway',
  },
  {
    path: 'metrics',
    loadComponent: () => import('./features/metrics/metrics.component').then(m => m.MetricsComponent),
    title: 'Metrics — straitgateway',
  },
  {
    path: 'traces',
    loadComponent: () => import('./features/traces/traces.component').then(m => m.TracesComponent),
    title: 'Traces — straitgateway',
  },
  {
    path: 'settings',
    loadComponent: () => import('./features/settings/settings.component').then(m => m.SettingsComponent),
    title: 'Settings — straitgateway',
  },
  {
    path: '**',
    redirectTo: '',
  },
];
