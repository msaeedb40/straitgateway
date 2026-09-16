import { Routes } from '@angular/router';

export const FLOWS_ROUTES: Routes = [
  {
    path: '',
    loadComponent: () => import('./flow-list/flow-list.component').then(m => m.FlowListComponent),
    title: 'Flows — StraitGateway',
  },
  {
    path: ':id',
    loadComponent: () => import('./flow-detail/flow-detail.component').then(m => m.FlowDetailComponent),
    title: 'Flow Detail — StraitGateway',
  },
];
