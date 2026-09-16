import { Routes } from '@angular/router';
import { authGuard } from './core/guards/auth.guard';

export const routes: Routes = [
  {
    path: 'login',
    loadComponent: () =>
      import('./features/auth/login/login.component').then((m) => m.LoginComponent),
  },
  {
    path: 'auth/callback',
    loadComponent: () =>
      import('./features/auth/callback/callback.component').then((m) => m.CallbackComponent),
  },
  {
    path: '',
    loadComponent: () =>
      import('./layout/shell/shell.component').then((m) => m.ShellComponent),
    canActivate: [authGuard],
    children: [
      { path: '', redirectTo: 'dashboard', pathMatch: 'full' },
      {
        path: 'dashboard',
        loadComponent: () =>
          import('./features/dashboard/dashboard.component').then((m) => m.DashboardComponent),
        title: 'Dashboard — StraitGateway',
      },
      {
        path: 'gateways',
        loadChildren: () =>
          import('./features/gateways/gateways.routes').then((m) => m.GATEWAYS_ROUTES),
        title: 'Gateways — StraitGateway',
      },
      {
        path: 'nodes',
        loadChildren: () =>
          import('./features/nodes/nodes.routes').then((m) => m.NODES_ROUTES),
        title: 'Nodes — StraitGateway',
      },
      {
        path: 'tunnels',
        loadChildren: () =>
          import('./features/tunnels/tunnels.routes').then((m) => m.TUNNELS_ROUTES),
        title: 'Tunnels — StraitGateway',
      },
      {
        path: 'flows',
        loadChildren: () =>
          import('./features/flows/flows.routes').then((m) => m.FLOWS_ROUTES),
        title: 'Flows — StraitGateway',
      },
      {
        path: 'packets',
        loadComponent: () =>
          import('./features/packets/packets.component').then((m) => m.PacketsComponent),
        title: 'Packet Capture — StraitGateway',
      },
      {
        path: 'topology',
        loadComponent: () =>
          import('./features/topology/topology.component').then((m) => m.TopologyComponent),
        title: 'Topology — StraitGateway',
      },
      {
        path: 'services',
        loadChildren: () =>
          import('./features/services/services.routes').then((m) => m.SERVICES_ROUTES),
        title: 'Services — StraitGateway',
      },
      {
        path: 'endpoints',
        loadChildren: () =>
          import('./features/endpoints/endpoints.routes').then((m) => m.ENDPOINTS_ROUTES),
        title: 'Endpoints — StraitGateway',
      },
      {
        path: 'ebpf',
        loadComponent: () =>
          import('./features/ebpf/ebpf.component').then((m) => m.EbpfComponent),
        title: 'eBPF — StraitGateway',
      },
      {
        path: 'cni',
        loadComponent: () =>
          import('./features/cni/cni.component').then((m) => m.CniComponent),
        title: 'CNI — StraitGateway',
      },
      {
        path: 'events',
        loadComponent: () =>
          import('./features/events/events.component').then((m) => m.EventsComponent),
        title: 'Events — StraitGateway',
      },
      {
        path: 'logs',
        loadComponent: () =>
          import('./features/logs/logs.component').then((m) => m.LogsComponent),
        title: 'Logs — StraitGateway',
      },
      {
        path: 'metrics',
        loadComponent: () =>
          import('./features/metrics/metrics.component').then((m) => m.MetricsComponent),
        title: 'Metrics — StraitGateway',
      },
      {
        path: 'traces',
        loadChildren: () =>
          import('./features/traces/traces.routes').then((m) => m.TRACES_ROUTES),
        title: 'Traces — StraitGateway',
      },
      {
        path: 'settings',
        loadChildren: () =>
          import('./features/settings/settings.routes').then((m) => m.SETTINGS_ROUTES),
        title: 'Settings — StraitGateway',
      },
    ],
  },
];
