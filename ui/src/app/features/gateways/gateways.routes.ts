import { Routes } from '@angular/router';

export const GATEWAYS_ROUTES: Routes = [
  {
    path: '',
    loadComponent: () => import('./gateway-list/gateway-list.component').then(m => m.GatewayListComponent),
    title: 'Gateways — StraitGateway',
  },
  {
    path: 'new',
    loadComponent: () => import('./gateway-form/gateway-form.component').then(m => m.GatewayFormComponent),
    title: 'Create Gateway — StraitGateway',
  },
  {
    path: ':name',
    loadComponent: () => import('./gateway-detail/gateway-detail.component').then(m => m.GatewayDetailComponent),
    title: 'Gateway Detail — StraitGateway',
  },
  {
    path: ':name/edit',
    loadComponent: () => import('./gateway-form/gateway-form.component').then(m => m.GatewayFormComponent),
    title: 'Edit Gateway — StraitGateway',
  },
];
