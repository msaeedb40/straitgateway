import { inject, Injectable } from '@angular/core';
import { Observable } from 'rxjs';
import { ApiClient } from '../api-client';

export interface Configuration {
  readonly key: string;
  readonly value: unknown;
  readonly description?: string;
  readonly updatedAt?: string;
}

export interface ConfigurationUpdateRequest {
  readonly key: string;
  readonly value: unknown;
}

@Injectable({ providedIn: 'root' })
export class ConfigurationApiService {
  private readonly client = inject(ApiClient);

  list(): Observable<Configuration[]> {
    return this.client.get<Configuration[]>('/v1/configuration');
  }

  get(key: string): Observable<Configuration> {
    return this.client.get<Configuration>(`/v1/configuration/${encodeURIComponent(key)}`);
  }

  update(req: ConfigurationUpdateRequest): Observable<Configuration> {
    return this.client.put<Configuration>(
      `/v1/configuration/${encodeURIComponent(req.key)}`,
      { value: req.value }
    );
  }

  delete(key: string): Observable<void> {
    return this.client.delete<void>(`/v1/configuration/${encodeURIComponent(key)}`);
  }
}
