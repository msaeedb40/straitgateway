import { inject, Injectable } from '@angular/core';
import { Observable } from 'rxjs';
import { ApiClient } from '../api-client';
import { PaginatedResponse, ListParams } from '../api.types';
import { KubernetesService } from '../../models/service.model';

@Injectable({ providedIn: 'root' })
export class ServiceApiService {
  private readonly client = inject(ApiClient);

  list(params: ListParams): Observable<PaginatedResponse<KubernetesService>> {
    return this.client.get<PaginatedResponse<KubernetesService>>('/v1/services', {
      params: params as Record<string, string>,
    });
  }

  get(namespace: string, name: string): Observable<KubernetesService> {
    return this.client.get<KubernetesService>(`/v1/namespaces/${namespace}/services/${name}`);
  }
}
