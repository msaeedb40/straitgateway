import { Routes } from '@angular/router';

export const NODES_ROUTES: Routes = [
  {
    path: '',
    loadComponent: () => import('./node-list/node-list.component').then(m => m.NodeListComponent),
    title: 'Nodes — StraitGateway',
  },
  {
    path: ':name',
    loadComponent: () => import('./node-detail/node-detail.component').then(m => m.NodeDetailComponent),
    title: 'Node Detail — StraitGateway',
  },
];
