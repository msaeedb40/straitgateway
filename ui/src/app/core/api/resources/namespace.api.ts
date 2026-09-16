import { inject, Injectable } from '@angular/core';
import { Observable } from 'rxjs';
import { ApiClient } from '../api-client';

export interface Namespace {
  readonly name: string;
  readonly cluster: string;
  readonly status: string;
  readonly labels?: Record<string, string>;
}

@Injectable({ providedIn: 'root' })
export class NamespaceApiService {
  private readonly client = inject(ApiClient);

  list(cluster?: string): Observable<Namespace[]> {
    return this.client.get<Namespace[]>('/v1/namespaces', {
      params: cluster ? { cluster } : undefined,
    });
  }
}
