import { inject, Injectable } from '@angular/core';
import { Observable } from 'rxjs';
import { ApiClient } from '../api-client';
import { RuntimeConfigService } from '../../config/runtime-config.service';
import { TraceReference, Trace, TraceFilter } from '../../models/trace.model';

interface JaegerTracesResponse {
  data: Array<{
    traceID: string;
    spans: unknown[];
    processes: Record<string, { serviceName: string }>;
  }>;
}

@Injectable({ providedIn: 'root' })
export class JaegerApiService {
  private readonly client = inject(ApiClient);
  private readonly runtimeConfig = inject(RuntimeConfigService);

  listServices(): Observable<string[]> {
    return new Observable((observer) => {
      this.client
        .external<{ data: string[] }>(this.runtimeConfig.jaegerBase, '/api/services')
        .subscribe({
          next: (res) => { observer.next(res.data); observer.complete(); },
          error: (err) => observer.error(err),
        });
    });
  }

  searchTraces(filter: TraceFilter): Observable<TraceReference[]> {
    const params: Record<string, string> = {
      ...(filter.service ? { service: filter.service } : {}),
      ...(filter.operation ? { operation: filter.operation } : {}),
      ...(filter.since ? { start: String(new Date(filter.since).getTime() * 1000) } : {}),
      ...(filter.until ? { end: String(new Date(filter.until).getTime() * 1000) } : {}),
      ...(filter.minDurationMs ? { minDuration: `${filter.minDurationMs}ms` } : {}),
      ...(filter.limit ? { limit: String(filter.limit) } : {}),
    };
    return new Observable((observer) => {
      this.client
        .external<JaegerTracesResponse>(this.runtimeConfig.jaegerBase, '/api/traces', { params })
        .subscribe({
          next: (res) => {
            const refs: TraceReference[] = res.data.map((t) => {
              const rootSpan = (t.spans as Array<{ operationName: string; startTime: number; duration: number; spanID: string; references: unknown[] }>)
                .find((s) => !s.references?.length);
              return {
                traceId: t.traceID,
                spanId: rootSpan?.spanID ?? '',
                operationName: rootSpan?.operationName ?? '',
                serviceName: Object.values(t.processes)[0]?.serviceName ?? '',
                startTime: rootSpan ? new Date(rootSpan.startTime / 1000).toISOString() : '',
                durationMs: rootSpan ? rootSpan.duration / 1000 : 0,
                status: 'Unset',
                spanCount: t.spans.length,
                errorCount: 0,
              };
            });
            observer.next(refs);
            observer.complete();
          },
          error: (err) => observer.error(err),
        });
    });
  }

  getTrace(traceId: string): Observable<Trace> {
    return new Observable((observer) => {
      this.client
        .external<{ data: unknown[] }>(this.runtimeConfig.jaegerBase, `/api/traces/${traceId}`)
        .subscribe({
          next: (res) => {
            // Delegate raw Jaeger trace → Trace model mapping
            observer.next(this.mapTrace(res.data[0]));
            observer.complete();
          },
          error: (err) => observer.error(err),
        });
    });
  }

  healthCheck(): Observable<boolean> {
    return new Observable((observer) => {
      this.client
        .external<unknown>(this.runtimeConfig.jaegerBase, '/')
        .subscribe({
          next: () => { observer.next(true); observer.complete(); },
          error: () => { observer.next(false); observer.complete(); },
        });
    });
  }

  private mapTrace(raw: unknown): Trace {
    // Cast raw Jaeger API shape → internal Trace model
    const t = raw as {
      traceID: string;
      spans: Array<{
        spanID: string;
        parentSpanIDs?: string[];
        operationName: string;
        startTime: number;
        duration: number;
        tags?: Array<{ key: string; value: string }>;
        logs?: Array<{ timestamp: number; fields: Array<{ key: string; value: string }> }>;
        processID: string;
        warnings?: string[];
      }>;
      processes: Record<string, { serviceName: string }>;
    };

    const spans = t.spans.map((s) => ({
      spanId: s.spanID,
      parentSpanId: s.parentSpanIDs?.[0] ?? null,
      traceId: t.traceID,
      operationName: s.operationName,
      serviceName: t.processes[s.processID]?.serviceName ?? '',
      startTime: new Date(s.startTime / 1000).toISOString(),
      durationMs: s.duration / 1000,
      status: (s.warnings?.length ? 'Error' : 'OK') as import('../../models/trace.model').TraceStatus,
      tags: Object.fromEntries((s.tags ?? []).map((tag) => [tag.key, tag.value])),
      logs: (s.logs ?? []).map((l) => ({
        timestamp: new Date(l.timestamp / 1000).toISOString(),
        fields: Object.fromEntries(l.fields.map((f) => [f.key, f.value])),
      })),
    }));

    const rootSpan = spans.find((s) => !s.parentSpanId) ?? spans[0];
    const services = [...new Set(spans.map((s) => s.serviceName))];

    return {
      traceId: t.traceID,
      spans,
      durationMs: spans.reduce((max, s) => Math.max(max, s.durationMs), 0),
      status: spans.some((s) => s.status === 'Error') ? 'Error' : 'OK',
      services,
      rootSpan,
    };
  }
}
