import { inject, Injectable } from '@angular/core';
import { Observable } from 'rxjs';
import { ApiClient } from '../api-client';
import { PaginatedResponse, ListParams } from '../api.types';
import { Node } from '../../models/node.model';

@Injectable({ providedIn: 'root' })
export class NodeApiService {
  private readonly client = inject(ApiClient);

  list(params: ListParams): Observable<PaginatedResponse<Node>> {
    return this.client.get<PaginatedResponse<Node>>('/v1/nodes', {
      params: params as Record<string, string>,
    });
  }

  get(name: string): Observable<Node> {
    return this.client.get<Node>(`/v1/nodes/${name}`);
  }
}
