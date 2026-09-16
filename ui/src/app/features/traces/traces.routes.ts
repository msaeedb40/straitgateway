import { Routes } from '@angular/router';

export const TRACES_ROUTES: Routes = [
  {
    path: '',
    loadComponent: () => import('./trace-list/trace-list.component').then(m => m.TraceListComponent),
    title: 'Traces — StraitGateway',
  },
  {
    path: ':traceId',
    loadComponent: () => import('./trace-detail/trace-detail.component').then(m => m.TraceDetailComponent),
    title: 'Trace Detail — StraitGateway',
  },
];
