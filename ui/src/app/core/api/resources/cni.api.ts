import { inject, Injectable } from '@angular/core';
import { Observable } from 'rxjs';
import { ApiClient } from '../api-client';
import { CniSummary, CniNodeStatus } from '../../models/cni.model';

@Injectable({ providedIn: 'root' })
export class CniApiService {
  private readonly client = inject(ApiClient);

  getSummary(): Observable<CniSummary> {
    return this.client.get<CniSummary>('/v1/cni');
  }

  getNodeStatus(nodeName: string): Observable<CniNodeStatus> {
    return this.client.get<CniNodeStatus>(`/v1/nodes/${nodeName}/cni`);
  }
}
