// Copyright 2026 straitgateway Authors
// SPDX-License-Identifier: Apache-2.0

import { Injectable, inject, signal, computed } from '@angular/core';
import { HttpClient } from '@angular/common/http';
import { firstValueFrom } from 'rxjs';
import { RuntimeConfig, DEFAULT_RUNTIME_CONFIG } from './runtime-config.model';

@Injectable({ providedIn: 'root' })
export class RuntimeConfigService {
  private http = inject(HttpClient);
  private readonly _config = signal<RuntimeConfig>(DEFAULT_RUNTIME_CONFIG);

  readonly config = this._config.asReadonly();
  readonly apiBase = computed(() => this._config().apiBase);
  readonly prometheusBase = computed(() => this._config().prometheusBase);
  readonly grafanaBase = computed(() => this._config().grafanaBase);
  readonly jaegerBase = computed(() => this._config().jaegerBase);
  readonly clusterName = computed(() => this._config().clusterName);
  readonly refreshIntervalMs = computed(() => this._config().refreshIntervalMs);

  get<K extends keyof RuntimeConfig>(key: K): RuntimeConfig[K] {
    return this._config()[key];
  }

  update(partial: Partial<RuntimeConfig>) {
    this._config.update((c) => ({ ...c, ...partial }));
  }

  async load(): Promise<void> {
    try {
      const cfg = await firstValueFrom(
        this.http.get<Partial<RuntimeConfig>>('/runtime-config.json')
      );
      if (cfg) {
        this._config.set({ ...DEFAULT_RUNTIME_CONFIG, ...cfg });
      }
    } catch {
      console.warn('runtime-config.json not loaded, running with defaults');
    }
  }
}
