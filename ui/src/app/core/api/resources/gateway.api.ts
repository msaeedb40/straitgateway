import { inject, Injectable } from '@angular/core';
import { Observable } from 'rxjs';
import { ApiClient } from '../api-client';
import { PaginatedResponse, ListParams } from '../api.types';
import {
  Gateway,
  GatewayCreateRequest,
  GatewayUpdateRequest,
  GatewayDeleteImpact,
} from '../../models/gateway.model';

@Injectable({ providedIn: 'root' })
export class GatewayApiService {
  private readonly client = inject(ApiClient);

  list(params: ListParams): Observable<PaginatedResponse<Gateway>> {
    return this.client.get<PaginatedResponse<Gateway>>('/v1/gateways', {
      params: params as Record<string, string>,
    });
  }

  get(namespace: string, name: string): Observable<Gateway> {
    return this.client.get<Gateway>(`/v1/namespaces/${namespace}/gateways/${name}`);
  }

  create(req: GatewayCreateRequest): Observable<Gateway> {
    return this.client.post<Gateway>(`/v1/namespaces/${req.namespace}/gateways`, req);
  }

  update(namespace: string, name: string, req: GatewayUpdateRequest): Observable<Gateway> {
    return this.client.patch<Gateway>(`/v1/namespaces/${namespace}/gateways/${name}`, req);
  }

  delete(namespace: string, name: string): Observable<void> {
    return this.client.delete<void>(`/v1/namespaces/${namespace}/gateways/${name}`);
  }

  getDeleteImpact(namespace: string, name: string): Observable<GatewayDeleteImpact> {
    return this.client.get<GatewayDeleteImpact>(
      `/v1/namespaces/${namespace}/gateways/${name}/delete-impact`
    );
  }

  reconcile(namespace: string, name: string): Observable<void> {
    return this.client.post<void>(
      `/v1/namespaces/${namespace}/gateways/${name}/reconcile`,
      {}
    );
  }
}
