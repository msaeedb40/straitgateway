import { Routes } from '@angular/router';

export const ENDPOINTS_ROUTES: Routes = [
  {
    path: '',
    loadComponent: () => import('./endpoint-list/endpoint-list.component').then(m => m.EndpointListComponent),
    title: 'Endpoints — StraitGateway',
  },
  {
    path: ':namespace/:name',
    loadComponent: () => import('./endpoint-detail/endpoint-detail.component').then(m => m.EndpointDetailComponent),
    title: 'Endpoint Detail — StraitGateway',
  },
];
