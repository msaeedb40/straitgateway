import { inject, Injectable } from '@angular/core';
import { Observable } from 'rxjs';
import { ApiClient } from '../api-client';
import { RuntimeConfigService } from '../../config/runtime-config.service';
import {
  MetricQueryParams,
  MetricQueryResult,
  InstantQueryResult,
} from '../../models/metric.model';

interface PrometheusRangeResponse {
  status: string;
  data: {
    resultType: 'matrix';
    result: Array<{
      metric: Record<string, string>;
      values: [number, string][];
    }>;
  };
}

interface PrometheusInstantResponse {
  status: string;
  data: {
    resultType: 'vector';
    result: Array<{
      metric: Record<string, string>;
      value: [number, string];
    }>;
  };
}

@Injectable({ providedIn: 'root' })
export class PrometheusApiService {
  private readonly client = inject(ApiClient);
  private readonly runtimeConfig = inject(RuntimeConfigService);

  queryRange(params: MetricQueryParams): Observable<MetricQueryResult> {
    return new Observable((observer) => {
      this.client
        .external<PrometheusRangeResponse>(
          this.runtimeConfig.prometheusBase,
          '/api/v1/query_range',
          {
            params: {
              query: params.query,
              start: params.start,
              end: params.end,
              step: params.step,
            },
          }
        )
        .subscribe({
          next: (res) => {
            const result: MetricQueryResult = {
              metric: params.query,
              series: res.data.result.map((r) => ({
                labels: r.metric,
                samples: r.values.map(([ts, val]) => ({
                  timestamp: ts,
                  value: parseFloat(val),
                })),
              })),
            };
            observer.next(result);
            observer.complete();
          },
          error: (err) => observer.error(err),
        });
    });
  }

  queryInstant(query: string): Observable<InstantQueryResult[]> {
    return new Observable((observer) => {
      this.client
        .external<PrometheusInstantResponse>(
          this.runtimeConfig.prometheusBase,
          '/api/v1/query',
          { params: { query } }
        )
        .subscribe({
          next: (res) => {
            const results: InstantQueryResult[] = res.data.result.map((r) => ({
              metric: query,
              labels: r.metric,
              timestamp: r.value[0],
              value: parseFloat(r.value[1]),
            }));
            observer.next(results);
            observer.complete();
          },
          error: (err) => observer.error(err),
        });
    });
  }

  healthCheck(): Observable<boolean> {
    return new Observable((observer) => {
      this.client
        .external<string>(this.runtimeConfig.prometheusBase, '/-/healthy')
        .subscribe({
          next: () => { observer.next(true); observer.complete(); },
          error: () => { observer.next(false); observer.complete(); },
        });
    });
  }
}
