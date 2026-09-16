import { inject, Injectable } from '@angular/core';
import { Observable } from 'rxjs';
import { ApiClient } from '../api-client';
import { TopologyGraph } from '../../models/topology.model';

@Injectable({ providedIn: 'root' })
export class TopologyApiService {
  private readonly client = inject(ApiClient);

  getGraph(cluster: string, namespace: string): Observable<TopologyGraph> {
    return this.client.get<TopologyGraph>('/v1/topology', {
      params: { cluster, namespace },
    });
  }
}
