// Copyright 2026 straitgateway Authors
// SPDX-License-Identifier: Apache-2.0

import { Injectable, signal, effect, inject } from '@angular/core';
import { StorageService } from './storage.service';

export interface UserPreferences {
  theme: 'dark' | 'light';
  refreshInterval: number; // in seconds
  compactTables: boolean;
  defaultNamespace: string;
  enableSoundAlerts: boolean;
  enableMockFallback: boolean;
}

const STORAGE_KEY = 'straitgateway_prefs_v1';

const DEFAULT_PREFS: UserPreferences = {
  theme: 'dark',
  refreshInterval: 15,
  compactTables: false,
  defaultNamespace: 'default',
  enableSoundAlerts: false,
  enableMockFallback: true,
};

@Injectable({ providedIn: 'root' })
export class PreferencesService {
  private storage = inject(StorageService);
  private readonly _prefs = signal<UserPreferences>(this.storage.getLocal(STORAGE_KEY, DEFAULT_PREFS));
  readonly prefs = this._prefs.asReadonly();

  constructor() {
    effect(() => {
      const p = this._prefs();
      this.storage.setLocal(STORAGE_KEY, p);
    });
  }

  update(partial: Partial<UserPreferences>) {
    this._prefs.update((cur) => ({ ...cur, ...partial }));
  }
}
