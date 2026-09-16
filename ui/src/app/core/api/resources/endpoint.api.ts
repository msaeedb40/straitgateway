import { inject, Injectable } from '@angular/core';
import { Observable } from 'rxjs';
import { ApiClient } from '../api-client';
import { PaginatedResponse, ListParams } from '../api.types';
import { Endpoint } from '../../models/endpoint.model';

@Injectable({ providedIn: 'root' })
export class EndpointApiService {
  private readonly client = inject(ApiClient);

  list(params: ListParams): Observable<PaginatedResponse<Endpoint>> {
    return this.client.get<PaginatedResponse<Endpoint>>('/v1/endpoints', {
      params: params as Record<string, string>,
    });
  }

  get(namespace: string, name: string): Observable<Endpoint> {
    return this.client.get<Endpoint>(`/v1/namespaces/${namespace}/endpoints/${name}`);
  }
}
