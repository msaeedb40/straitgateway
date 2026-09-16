import { inject, Injectable } from '@angular/core';
import { Observable } from 'rxjs';
import { ApiClient } from '../api-client';
import { PaginatedResponse } from '../api.types';
import { Flow, FlowFilter, FlowPath } from '../../models/flow.model';

@Injectable({ providedIn: 'root' })
export class FlowApiService {
  private readonly client = inject(ApiClient);

  list(filter: FlowFilter): Observable<PaginatedResponse<Flow>> {
    return this.client.get<PaginatedResponse<Flow>>('/v1/flows', {
      params: filter as Record<string, string>,
    });
  }

  get(id: string): Observable<Flow> {
    return this.client.get<Flow>(`/v1/flows/${id}`);
  }

  getPath(id: string): Observable<FlowPath> {
    return this.client.get<FlowPath>(`/v1/flows/${id}/path`);
  }
}
