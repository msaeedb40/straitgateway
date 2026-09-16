import { Routes } from '@angular/router';

export const TUNNELS_ROUTES: Routes = [
  {
    path: '',
    loadComponent: () => import('./tunnel-list/tunnel-list.component').then(m => m.TunnelListComponent),
    title: 'Tunnels — StraitGateway',
  },
  {
    path: 'new',
    loadComponent: () => import('./tunnel-form/tunnel-form.component').then(m => m.TunnelFormComponent),
    title: 'Create Tunnel — StraitGateway',
  },
  {
    path: ':id',
    loadComponent: () => import('./tunnel-detail/tunnel-detail.component').then(m => m.TunnelDetailComponent),
    title: 'Tunnel Detail — StraitGateway',
  },
  {
    path: ':id/edit',
    loadComponent: () => import('./tunnel-form/tunnel-form.component').then(m => m.TunnelFormComponent),
    title: 'Edit Tunnel — StraitGateway',
  },
];
