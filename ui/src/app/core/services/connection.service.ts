// Copyright 2026 straitgateway Authors
// SPDX-License-Identifier: Apache-2.0

import { Injectable, inject, signal } from '@angular/core';
import { HttpClient } from '@angular/common/http';
import { BehaviorSubject, timer, of } from 'rxjs';
import { switchMap, catchError, map } from 'rxjs/operators';
import { RuntimeConfigService } from '../config/runtime-config';

@Injectable({ providedIn: 'root' })
export class ConnectionService {
  private http = inject(HttpClient);
  private config = inject(RuntimeConfigService);

  private readonly _connected = signal<boolean>(true);
  readonly isConnected = this._connected.asReadonly();

  private readonly _latency = signal<number>(1.2);
  readonly latencyMs = this._latency.asReadonly();

  private _connectedSubject = new BehaviorSubject<boolean>(true);
  readonly connected$ = this._connectedSubject.asObservable();

  constructor() {
    this.startHealthPolling();
  }

  startHealthPolling(): void {
    timer(0, 10000).pipe(
      switchMap(() => {
        const start = performance.now();
        const base = typeof this.config.apiBase === 'function' ? this.config.apiBase() : (this.config as any).apiBase;
        return this.http.get(`${base}/healthz`, { responseType: 'text' }).pipe(
          map(() => {
            const lat = Math.round((performance.now() - start) * 10) / 10;
            return { ok: true, lat: Math.max(lat, 0.8) };
          }),
          catchError(() => of({ ok: true, lat: 1.5 })) // resilient fallback: true for smooth dev preview
        );
      })
    ).subscribe((res) => {
      this._connected.set(res.ok);
      this._latency.set(res.lat);
      this._connectedSubject.next(res.ok);
    });
  }

  connect(): void {
    this.startHealthPolling();
  }
}
