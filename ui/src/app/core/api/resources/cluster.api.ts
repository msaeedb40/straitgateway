import { inject, Injectable } from '@angular/core';
import { Observable } from 'rxjs';
import { ApiClient } from '../api-client';

export interface Cluster {
  readonly name: string;
  readonly apiServer: string;
  readonly version: string;
  readonly status: 'Reachable' | 'Unreachable' | 'Unknown';
}

@Injectable({ providedIn: 'root' })
export class ClusterApiService {
  private readonly client = inject(ApiClient);

  list(): Observable<Cluster[]> {
    return this.client.get<Cluster[]>('/v1/clusters');
  }

  get(name: string): Observable<Cluster> {
    return this.client.get<Cluster>(`/v1/clusters/${name}`);
  }
}
