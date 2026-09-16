import { Routes } from '@angular/router';

export const SETTINGS_ROUTES: Routes = [
  { path: '', redirectTo: 'general', pathMatch: 'full' },
  {
    path: 'general',
    loadComponent: () => import('./general/settings-general.component').then(m => m.SettingsGeneralComponent),
    title: 'General Settings — StraitGateway',
  },
  {
    path: 'gateway',
    loadComponent: () => import('./gateway/settings-gateway.component').then(m => m.SettingsGatewayComponent),
    title: 'Gateway Settings — StraitGateway',
  },
  {
    path: 'network',
    loadComponent: () => import('./network/settings-network.component').then(m => m.SettingsNetworkComponent),
    title: 'Network Settings — StraitGateway',
  },
  {
    path: 'ebpf',
    loadComponent: () => import('./ebpf/settings-ebpf.component').then(m => m.SettingsEbpfComponent),
    title: 'eBPF Settings — StraitGateway',
  },
  {
    path: 'cni',
    loadComponent: () => import('./cni/settings-cni.component').then(m => m.SettingsCniComponent),
    title: 'CNI Settings — StraitGateway',
  },
  {
    path: 'observability',
    loadComponent: () => import('./observability/settings-observability.component').then(m => m.SettingsObservabilityComponent),
    title: 'Observability Settings — StraitGateway',
  },
  {
    path: 'security',
    loadComponent: () => import('./security/settings-security.component').then(m => m.SettingsSecurityComponent),
    title: 'Security Settings — StraitGateway',
  },
  {
    path: 'advanced',
    loadComponent: () => import('./advanced/settings-advanced.component').then(m => m.SettingsAdvancedComponent),
    title: 'Advanced Settings — StraitGateway',
  },
];
