import { inject, Injectable } from '@angular/core';
import { Observable } from 'rxjs';
import { ApiClient } from '../api-client';
import { PaginatedResponse, ListParams } from '../api.types';
import { Tunnel, TunnelCreateRequest, TunnelUpdateRequest } from '../../models/tunnel.model';

@Injectable({ providedIn: 'root' })
export class TunnelApiService {
  private readonly client = inject(ApiClient);

  list(params: ListParams): Observable<PaginatedResponse<Tunnel>> {
    return this.client.get<PaginatedResponse<Tunnel>>('/v1/tunnels', {
      params: params as Record<string, string>,
    });
  }

  get(namespace: string, name: string): Observable<Tunnel> {
    return this.client.get<Tunnel>(`/v1/namespaces/${namespace}/tunnels/${name}`);
  }

  create(req: TunnelCreateRequest): Observable<Tunnel> {
    return this.client.post<Tunnel>(`/v1/namespaces/${req.namespace}/tunnels`, req);
  }

  update(namespace: string, name: string, req: TunnelUpdateRequest): Observable<Tunnel> {
    return this.client.patch<Tunnel>(`/v1/namespaces/${namespace}/tunnels/${name}`, req);
  }

  delete(namespace: string, name: string): Observable<void> {
    return this.client.delete<void>(`/v1/namespaces/${namespace}/tunnels/${name}`);
  }
}
