import { inject, Injectable } from '@angular/core';
import { Observable } from 'rxjs';
import { ApiClient } from '../api-client';
import { PaginatedResponse } from '../api.types';
import { StraitEvent, EventFilter } from '../../models/event.model';

@Injectable({ providedIn: 'root' })
export class EventApiService {
  private readonly client = inject(ApiClient);

  list(filter: EventFilter): Observable<PaginatedResponse<StraitEvent>> {
    return this.client.get<PaginatedResponse<StraitEvent>>('/v1/events', {
      params: filter as Record<string, string>,
    });
  }

  /** SSE endpoint URL for live event streaming — caller opens EventSource */
  streamUrl(): string {
    return `${this.client.controllerBase}/v1/events/stream`;
  }
}
