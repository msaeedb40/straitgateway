import { inject, Injectable } from '@angular/core';
import { Observable } from 'rxjs';
import { ApiClient } from '../api-client';
import { RuntimeConfigService } from '../../config/runtime-config.service';
import { LogEntry, LogFilter } from '../../models/log.model';
import { PaginatedResponse } from '../api.types';

@Injectable({ providedIn: 'root' })
export class LogsApiService {
  private readonly client = inject(ApiClient);
  private readonly runtimeConfig = inject(RuntimeConfigService);

  query(filter: LogFilter): Observable<PaginatedResponse<LogEntry>> {
    return this.client.external<PaginatedResponse<LogEntry>>(
      this.runtimeConfig.logsBase,
      '/api/v1/logs',
      { params: filter as Record<string, string> }
    );
  }

  /** Returns SSE URL for live log tail — caller opens EventSource with filter params */
  tailUrl(filter: LogFilter): string {
    const qs = new URLSearchParams();
    for (const [k, v] of Object.entries(filter)) {
      if (v !== undefined && v !== null) qs.set(k, String(v));
    }
    return `${this.runtimeConfig.logsBase}/api/v1/logs/tail?${qs.toString()}`;
  }

  healthCheck(): Observable<boolean> {
    return new Observable((observer) => {
      this.client
        .external<unknown>(this.runtimeConfig.logsBase, '/health')
        .subscribe({
          next: () => { observer.next(true); observer.complete(); },
          error: () => { observer.next(false); observer.complete(); },
        });
    });
  }
}
