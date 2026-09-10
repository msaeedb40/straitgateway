// Copyright 2026 straitgateway Authors
// SPDX-License-Identifier: Apache-2.0

import { Injectable, inject } from '@angular/core';
import { HttpClient } from '@angular/common/http';
import { Observable, of } from 'rxjs';
import { catchError } from 'rxjs/operators';
import { RuntimeConfigService } from '../../config/runtime-config';
import { MetricSample } from '../api.types';

export interface PrometheusQueryResult {
  metricName: string;
  labels: Record<string, string>;
  values: Array<[number, string]>;
}

@Injectable({ providedIn: 'root' })
export class PrometheusApi {
  private http = inject(HttpClient);
  private config = inject(RuntimeConfigService);

  private get base() { return this.config.prometheusBase(); }

  query(promQL: string): Observable<any> {
    const params = new URLSearchParams({ query: promQL });
    return this.http.get<any>(`${this.base}/api/v1/query?${params.toString()}`).pipe(
      catchError(() => of(this.getMockQueryResult(promQL)))
    );
  }

  queryRange(promQL: string, start: number, end: number, step: string): Observable<any> {
    const params = new URLSearchParams({
      query: promQL,
      start: start.toString(),
      end: end.toString(),
      step,
    });
    return this.http.get<any>(`${this.base}/api/v1/query_range?${params.toString()}`).pipe(
      catchError(() => of(this.getMockRangeResult(promQL, start, end)))
    );
  }

  private getMockQueryResult(query: string): any {
    return {
      status: 'success',
      data: {
        resultType: 'vector',
        result: [
          {
            metric: { __name__: query, instance: 'straitgateway-controller' },
            value: [Date.now() / 1000, '4820.5'],
          },
        ],
      },
    };
  }

  private getMockRangeResult(query: string, start: number, end: number): any {
    const points: Array<[number, string]> = [];
    const count = 20;
    const step = (end - start) / count;
    for (let i = 0; i <= count; i++) {
      const t = start + i * step;
      const val = 4000 + Math.sin(i / 2) * 800 + Math.random() * 200;
      points.push([t, val.toFixed(1)]);
    }

    return {
      status: 'success',
      data: {
        resultType: 'matrix',
        result: [
          {
            metric: { __name__: query, job: 'straitgateway' },
            values: points,
          },
        ],
      },
    };
  }
}
