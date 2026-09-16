import { Routes } from '@angular/router';

export const SERVICES_ROUTES: Routes = [
  {
    path: '',
    loadComponent: () => import('./service-list/service-list.component').then(m => m.ServiceListComponent),
    title: 'Services — StraitGateway',
  },
  {
    path: ':namespace/:name',
    loadComponent: () => import('./service-detail/service-detail.component').then(m => m.ServiceDetailComponent),
    title: 'Service Detail — StraitGateway',
  },
];
